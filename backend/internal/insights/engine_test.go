package insights

import (
	"testing"
	"time"
)

func pfloat(v float64) *float64 { return &v }

func TestLowUtilizationInsightUsesActualNumbers(t *testing.T) {
	net := Network{Utilization: pfloat(24), MarginPct: pfloat(50), MarginPerKWh: 10}
	stations := []Station{{
		ID: "BLR-004", Name: "Electronic City Phase 1", Area: "Electronic City",
		HasDC: true, MaxRatedKW: 60, ConnectorCount: 2,
		Utilization: pfloat(8), MarginPct: pfloat(48), MarginPerKWh: 9,
		Energy: 400, Revenue: 8000, Sessions: 40,
		EnergyPerConnectorPerDay: 6.5, AvailableHours: 700,
	}}
	got := Generate(net, stations, time.Time{}, time.Time{})
	if len(got) != 1 {
		t.Fatalf("got %d insights: %+v", len(got), got)
	}
	in := got[0]
	if in.Category != "underutilized_fast_charging" || in.Priority != PriorityHigh {
		t.Fatalf("category=%s priority=%s", in.Category, in.Priority)
	}
	if in.StationIDs[0] != "BLR-004" {
		t.Fatal(in.StationIDs)
	}
	if !contains(in.Title, "8.0%") || !contains(in.Title, "24.0%") {
		t.Fatalf("title %q", in.Title)
	}
}

func TestHealthyStationProducesNoLowUtilizationInsight(t *testing.T) {
	net := Network{Utilization: pfloat(24), MarginPct: pfloat(50), MarginPerKWh: 10}
	stations := []Station{{
		ID: "BLR-001", Name: "Busy", HasDC: true, MaxRatedKW: 60, ConnectorCount: 2,
		Utilization: pfloat(30), MarginPct: pfloat(52), MarginPerKWh: 11,
		Energy: 5000, Revenue: 100000, Sessions: 200, AvailableHours: 700,
		EnergyPerConnectorPerDay: 40,
	}}
	for _, in := range Generate(net, stations, time.Time{}, time.Time{}) {
		if in.Category == "low_utilization" || in.Category == "underutilized_fast_charging" {
			t.Fatalf("unexpected %+v", in)
		}
	}
}

func TestLowMarginAndZeroSessionEdges(t *testing.T) {
	net := Network{Utilization: pfloat(20), MarginPct: pfloat(55), MarginPerKWh: 11}
	busyThin := Station{
		ID: "BLR-007", Name: "Depot", HasDC: true, MaxRatedKW: 150, ConnectorCount: 2,
		Utilization: pfloat(32), MarginPct: pfloat(12), MarginPerKWh: 2,
		Energy: 1000, Revenue: 18000, Sessions: 80, AvailableHours: 600,
	}
	idle := Station{ID: "BLR-017", Name: "Quiet", Utilization: nil, AvailableHours: 0}
	got := Generate(net, []Station{busyThin, idle}, time.Time{}, time.Time{})
	found := false
	for _, in := range got {
		if in.Category == "high_utilization_low_margin" && in.StationIDs[0] == "BLR-007" {
			found = true
			if !contains(in.EstimatedImpact, "kWh") {
				t.Fatalf("impact %q", in.EstimatedImpact)
			}
		}
		if len(in.StationIDs) == 1 && in.StationIDs[0] == "BLR-017" && in.Category == "low_utilization" {
			t.Fatal("nil utilization must not become a low-utilization insight")
		}
	}
	if !found {
		t.Fatalf("missing margin insight: %+v", got)
	}
}

func TestDowntimeConcentration(t *testing.T) {
	net := Network{Utilization: pfloat(20), MarginPct: pfloat(40), DowntimeHours: 100}
	stations := []Station{
		{ID: "A", Name: "A", DowntimeHours: 50, Faults: 2, Uptime: pfloat(80), RevenueAtRisk: 1000, AvailableHours: 100},
		{ID: "B", Name: "B", DowntimeHours: 30, Faults: 1, Uptime: pfloat(90), RevenueAtRisk: 500, AvailableHours: 100},
		{ID: "C", Name: "C", DowntimeHours: 15, Faults: 1, Uptime: pfloat(95), RevenueAtRisk: 100, AvailableHours: 100},
		{ID: "D", Name: "D", DowntimeHours: 5, Uptime: pfloat(99)},
	}
	got := Generate(net, stations, time.Time{}, time.Time{})
	found := false
	for _, in := range got {
		if in.Category == "downtime_concentration" {
			found = true
			if len(in.StationIDs) != 3 {
				t.Fatalf("ids %+v", in.StationIDs)
			}
		}
	}
	if !found {
		t.Fatal("expected downtime concentration")
	}
	alone := Generate(Network{DowntimeHours: 40}, []Station{
		{ID: "A", Name: "A", DowntimeHours: 40, Faults: 1, Uptime: pfloat(80), RevenueAtRisk: 100, AvailableHours: 100},
	}, time.Time{}, time.Time{})
	for _, in := range alone {
		if in.Category == "downtime_concentration" {
			t.Fatal("one station is not a concentration")
		}
	}
}

func contains(s, part string) bool {
	return len(s) >= len(part) && (s == part || len(part) == 0 || indexOf(s, part) >= 0)
}

func indexOf(s, part string) int {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
