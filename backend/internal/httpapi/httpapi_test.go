package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"

	"chargingdatainfra/backend/internal/alerts"
	"chargingdatainfra/backend/internal/analytics"
	"chargingdatainfra/backend/internal/domain"
	"chargingdatainfra/backend/internal/seed"
	"chargingdatainfra/backend/internal/store"
)

func testServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	ds := seed.Generate(seed.Options{Seed: 20261017, Days: 30, AsOf: time.Date(2026, 10, 17, 18, 30, 0, 0, loc), Location: loc})
	st := store.New(ds)
	svc := analytics.New(st)
	asOf := ds.AsOf
	last30, err := svc.Compute(context.Background(), domain.Query{From: asOf.AddDate(0, 0, -30), To: asOf})
	if err != nil {
		t.Fatal(err)
	}
	last7, err := svc.Compute(context.Background(), domain.Query{From: asOf.AddDate(0, 0, -7), To: asOf})
	if err != nil {
		t.Fatal(err)
	}
	prior, err := svc.Compute(context.Background(), domain.Query{From: asOf.AddDate(0, 0, -30), To: asOf.AddDate(0, 0, -7)})
	if err != nil {
		t.Fatal(err)
	}
	st.SetAlerts(alerts.Generate(asOf, st.Faults(), last30, last30, last7, prior))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httptest.NewServer(New(logger, st, svc)), st
}

func TestHealthSummaryAndAlertPatch(t *testing.T) {
	srv, st := testServer(t)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("health %d", res.StatusCode)
	}
	var health struct {
		Demo bool `json:"demo"`
		Data struct {
			StationCount int `json:"station_count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		t.Fatal(err)
	}
	if !health.Demo || health.Data.StationCount != 20 {
		t.Fatalf("%+v", health)
	}

	bad, err := http.Get(srv.URL + "/api/v1/dashboard/summary?from=2026-10-17T00:00:00%2B05:30&to=2026-10-01T00:00:00%2B05:30")
	if err != nil {
		t.Fatal(err)
	}
	defer bad.Body.Close()
	if bad.StatusCode != 400 {
		t.Fatalf("bad range %d", bad.StatusCode)
	}

	missing, err := http.Get(srv.URL + "/api/v1/stations/NOPE")
	if err != nil {
		t.Fatal(err)
	}
	defer missing.Body.Close()
	if missing.StatusCode != 404 {
		t.Fatalf("missing station %d", missing.StatusCode)
	}

	sum, err := http.Get(srv.URL + "/api/v1/dashboard/summary?from=2026-10-01T00:00:00%2B05:30&to=2026-10-17T18:30:00%2B05:30")
	if err != nil {
		t.Fatal(err)
	}
	defer sum.Body.Close()
	if sum.StatusCode != 200 {
		t.Fatalf("summary %d", sum.StatusCode)
	}

	alertsRes, err := http.Get(srv.URL + "/api/v1/alerts?severity=critical")
	if err != nil {
		t.Fatal(err)
	}
	defer alertsRes.Body.Close()
	var listed struct {
		Data struct {
			Alerts []domain.Alert `json:"alerts"`
		} `json:"data"`
	}
	if err := json.NewDecoder(alertsRes.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Data.Alerts) == 0 {
		t.Fatal("expected critical alerts")
	}
	id := listed.Data.Alerts[0].ID
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/alerts/"+id, strings.NewReader(`{"status":"acknowledged"}`))
	req.Header.Set("Content-Type", "application/json")
	patched, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer patched.Body.Close()
	if patched.StatusCode != 200 {
		t.Fatalf("patch %d", patched.StatusCode)
	}
	found := false
	for _, al := range st.Alerts() {
		if al.ID == id && al.Status == "acknowledged" {
			found = true
		}
	}
	if !found {
		t.Fatal("status was not stored")
	}

	badPatch, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/alerts/"+id, strings.NewReader(`{"status":"nope"}`))
	badRes, err := http.DefaultClient.Do(badPatch)
	if err != nil {
		t.Fatal(err)
	}
	defer badRes.Body.Close()
	if badRes.StatusCode != 400 {
		t.Fatalf("bad status %d", badRes.StatusCode)
	}
}
