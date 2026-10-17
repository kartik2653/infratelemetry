package analytics_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
	_ "time/tzdata"

	"chargingdatainfra/backend/internal/analytics"
	"chargingdatainfra/backend/internal/domain"
	"chargingdatainfra/backend/internal/insights"
	"chargingdatainfra/backend/internal/seed"
	"chargingdatainfra/backend/internal/store"
)

func TestDemoNetworkTellsTheOperatingStories(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	asOf := time.Date(2026, 10, 17, 18, 30, 0, 0, loc)
	ds := seed.Generate(seed.Options{Seed: 20261017, Days: 90, AsOf: asOf, Location: loc})
	svc := analytics.New(store.New(ds))
	snap, err := svc.Compute(context.Background(), domain.Query{From: asOf.AddDate(0, 0, -30), To: asOf})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := snap.Network.Utilization(); !ok {
		t.Fatal("network utilization undefined")
	}
	var sumEnergy float64
	var sumCharge time.Duration
	for _, sm := range snap.Stations {
		sumEnergy += sm.Energy
		sumCharge += sm.Charging
		util, ok := sm.Utilization()
		if ok && util > 100.01 {
			t.Fatalf("%s utilization %.2f exceeds 100", sm.Station.ID, util)
		}
		t.Logf("%s util=%s uptime=%s margin=%s energy=%.0f downtime=%.1fh sessions=%d pay=%d faults=%d",
			sm.Station.ID, fmtPct(sm.Utilization()), fmtPct(sm.Uptime()), fmtPct(marginOnly(sm)),
			sm.Energy, sm.Downtime.Hours(), sm.Sessions, sm.PaymentFails, sm.ConnectorFaults)
	}
	if math.Abs(sumEnergy-snap.Network.Energy) > 0.05 {
		t.Fatalf("energy sum %.2f network %.2f", sumEnergy, snap.Network.Energy)
	}
	if sumCharge != snap.Network.Charging {
		t.Fatalf("charging sum mismatch")
	}
	found := map[string]bool{}
	for _, in := range insights.FromSnapshot(snap) {
		found[in.Category] = true
		t.Logf("insight %s %s", in.Category, in.Title)
	}
	for _, cat := range []string{
		"underutilized_fast_charging",
		"high_utilization_low_margin",
		"repeated_faults",
		"downtime_concentration",
		"payment_failures",
	} {
		if !found[cat] {
			t.Errorf("missing insight category %s", cat)
		}
	}
}

func marginOnly(sm analytics.StationMetric) (float64, bool) {
	_, pct, ok := sm.Margin()
	return pct, ok
}

func fmtPct(v float64, ok bool) string {
	if !ok {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", v)
}
