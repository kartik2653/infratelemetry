package alerts

import (
	"testing"
	"time"

	"chargingdatainfra/backend/internal/analytics"
	"chargingdatainfra/backend/internal/domain"
)

func TestOfflineAndPaymentAlerts(t *testing.T) {
	asOf := time.Date(2026, 10, 17, 18, 0, 0, 0, time.UTC)
	faults := []domain.Fault{
		{ID: "FLT-1", StationID: "BLR-006", ConnectorID: "BLR-006-B", Category: domain.FaultConnector, Start: asOf.Add(-5 * time.Hour), Description: "open"},
		{ID: "FLT-2", StationID: "BLR-001", ConnectorID: "BLR-001-A", Category: domain.FaultCommunication, Start: asOf.Add(-48 * time.Hour), End: ptr(asOf.Add(-47 * time.Hour))},
	}
	last14 := &analytics.Snapshot{Stations: []analytics.StationMetric{{
		Station: domain.Station{ID: "BLR-010", Name: "Marathahalli Bridge"},
		Totals:  analytics.Totals{Sessions: 40, PaymentFails: 8},
	}}}
	got := Generate(asOf, faults, nil, last14, nil, nil)
	var offline, payment int
	for _, al := range got {
		if al.Category == "charger_offline" {
			offline++
			if al.Status != "open" || al.ConnectorID != "BLR-006-B" || al.Severity != "critical" {
				t.Fatalf("%+v", al)
			}
		}
		if al.Category == "repeated_payment_failure" {
			payment++
		}
	}
	if offline != 1 || payment != 1 {
		t.Fatalf("offline=%d payment=%d alerts=%d", offline, payment, len(got))
	}
}

func TestUtilizationDropAndEmptyWindows(t *testing.T) {
	asOf := time.Date(2026, 10, 17, 18, 0, 0, 0, time.UTC)
	seven := 7.0
	twenty := 20.0
	last7 := snapUtil("BLR-004", "Electronic City", seven, 10, 12)
	prior := snapUtil("BLR-004", "Electronic City", twenty, 30, 12)
	got := Generate(asOf, nil, nil, nil, last7, prior)
	if len(got) != 1 || got[0].Category != "utilization_drop" {
		t.Fatalf("%+v", got)
	}
	if got[0].ObservedValue == "" || got[0].RecommendedAction == "" {
		t.Fatal("alert missing evidence")
	}
	if len(Generate(asOf, nil, nil, nil, nil, nil)) != 0 {
		t.Fatal("empty input should produce no alerts")
	}
}

func snapUtil(id, name string, util float64, sessions int, avg float64) *analytics.Snapshot {
	// Utilization is derived from durations, so set charging and available directly.
	available := 100 * time.Hour
	charging := time.Duration(util / 100 * float64(available))
	return &analytics.Snapshot{Stations: []analytics.StationMetric{{
		Station: domain.Station{ID: id, Name: name},
		Totals: analytics.Totals{
			Charging: charging, Available: available,
			Sessions: sessions, Energy: avg * float64(sessions),
		},
	}}}
}

func ptr(t time.Time) *time.Time { return &t }
