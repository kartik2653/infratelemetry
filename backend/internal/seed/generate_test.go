package seed

import (
	"math"
	"testing"
	"time"
	_ "time/tzdata"

	"chargingdatainfra/backend/internal/domain"
)

func demo() domain.Dataset {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return Generate(Options{
		Seed: 20261017, Days: 90,
		AsOf:     time.Date(2026, 10, 17, 18, 30, 0, 0, loc),
		Location: loc,
	})
}

func TestDeterministicAndConsistent(t *testing.T) {
	a := demo()
	b := demo()
	if len(a.Stations) != 20 {
		t.Fatalf("stations=%d", len(a.Stations))
	}
	if len(a.Connectors) < 40 || len(a.Connectors) > 60 {
		t.Fatalf("connectors=%d", len(a.Connectors))
	}
	if len(a.Sessions) < 5000 {
		t.Fatalf("sessions=%d", len(a.Sessions))
	}
	if len(a.Sessions) != len(b.Sessions) || a.Sessions[0].ID != b.Sessions[0].ID || a.Sessions[len(a.Sessions)-1].RevenueINR != b.Sessions[len(b.Sessions)-1].RevenueINR {
		t.Fatal("generation is not deterministic")
	}

	connectors := map[string]domain.Connector{}
	stations := map[string]bool{}
	for _, st := range a.Stations {
		stations[st.ID] = true
	}
	for _, c := range a.Connectors {
		connectors[c.ID] = c
		if !stations[c.StationID] {
			t.Fatalf("connector %s has unknown station", c.ID)
		}
	}
	status := map[string]int{}
	for _, c := range a.Connectors {
		status[string(c.Status)]++
	}
	for _, want := range []string{"available", "occupied", "unavailable", "faulted"} {
		if status[want] == 0 {
			t.Fatalf("missing status %s in %+v", want, status)
		}
	}

	byConn := map[string][]domain.Session{}
	var telEnergy = map[string]float64{}
	for _, sample := range a.Telemetry {
		telEnergy[sample.SessionID] += sample.EnergyKWh
	}
	for _, s := range a.Sessions {
		if !stations[s.StationID] {
			t.Fatalf("bad station %s", s.StationID)
		}
		if _, ok := connectors[s.ConnectorID]; !ok {
			t.Fatalf("bad connector %s", s.ConnectorID)
		}
		if s.End.Before(s.Start) || s.Start.Before(a.DataStart) || s.End.After(a.AsOf) {
			t.Fatalf("bad bounds %s %s %s", s.ID, s.Start, s.End)
		}
		if s.DurationSec != int64(s.End.Sub(s.Start).Seconds()) {
			t.Fatalf("duration %s", s.ID)
		}
		if s.ChargeDurationSec > s.DurationSec {
			t.Fatalf("charge longer than plug %s", s.ID)
		}
		if s.Status == domain.SessionFailed && (s.EnergyKWh != 0 || s.RevenueINR != 0) {
			t.Fatalf("failed session has revenue %s", s.ID)
		}
		wantRev := domain.Round2(s.EnergyKWh * s.TariffINRPerKWh)
		if math.Abs(wantRev-s.RevenueINR) > 0.001 {
			t.Fatalf("revenue %s got %v want %v", s.ID, s.RevenueINR, wantRev)
		}
		if s.EnergyKWh > 0 {
			if math.Abs(telEnergy[s.ID]-s.EnergyKWh) > 0.05 {
				t.Fatalf("telemetry energy %s got %v session %v", s.ID, telEnergy[s.ID], s.EnergyKWh)
			}
		}
		byConn[s.ConnectorID] = append(byConn[s.ConnectorID], s)
	}
	for id, list := range byConn {
		for i := 1; i < len(list); i++ {
			if list[i].Start.Before(list[i-1].End) {
				t.Fatalf("overlap on %s %s and %s", id, list[i-1].ID, list[i].ID)
			}
		}
	}
}
