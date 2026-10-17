package analytics

import (
	"context"
	"math"
	"testing"
	"time"
	_ "time/tzdata"

	"chargingdatainfra/backend/internal/domain"
	"chargingdatainfra/backend/internal/store"
)

func ist(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func at(loc *time.Location, hour, min int) time.Time {
	return time.Date(2026, 10, 17, hour, min, 0, 0, loc)
}

type fixture struct {
	asOf time.Time
	st   *store.Store
	svc  *Service
}

func newFixture(t *testing.T, sessions []domain.Session, faults []domain.Fault, maint []domain.MaintenanceWindow) fixture {
	t.Helper()
	loc := ist(t)
	asOf := at(loc, 18, 0)
	ds := domain.Dataset{
		OperatorName: "Test", NetworkName: "Test", Disclaimer: domain.Disclaimer,
		Timezone: domain.Timezone, AsOf: asOf, DataStart: at(loc, 0, 0),
		Stations: []domain.Station{
			{ID: "BLR-001", Name: "One", Area: "Indiranagar", City: "Bengaluru", PeakTariffINR: 20, OffPeakTariffINR: 20, PeakEnergyCostINR: 8, OffPeakEnergyCostINR: 8},
			{ID: "BLR-002", Name: "Two", Area: "Hebbal", City: "Bengaluru", PeakTariffINR: 20, OffPeakTariffINR: 20, PeakEnergyCostINR: 8, OffPeakEnergyCostINR: 8},
		},
		Connectors: []domain.Connector{
			{ID: "BLR-001-A", StationID: "BLR-001", Label: "A", ChargerType: domain.ChargerDC, RatedPowerKW: 50, Status: domain.StatusAvailable},
			{ID: "BLR-002-A", StationID: "BLR-002", Label: "A", ChargerType: domain.ChargerAC, RatedPowerKW: 22, Status: domain.StatusFaulted},
		},
		Sessions: sessions, Faults: faults, Maintenance: maint,
	}
	st := store.New(ds)
	return fixture{asOf: asOf, st: st, svc: New(st)}
}

func session(id, station, conn string, start, end time.Time, energy, tariff, cost float64, status domain.SessionStatus) domain.Session {
	sec := int64(end.Sub(start).Seconds())
	charge := sec
	if status == domain.SessionFailed {
		charge = 0
		energy = 0
		tariff = 0
		cost = 0
	}
	return domain.Session{
		ID: id, StationID: station, ConnectorID: conn, Start: start, End: end,
		DurationSec: sec, ChargeDurationSec: charge, EnergyKWh: energy,
		TariffINRPerKWh: tariff, RevenueINR: domain.Round2(energy * tariff), EnergyCostINR: domain.Round2(energy * cost),
		Status: status, ChargerType: domain.ChargerDC, RatedPowerKW: 50,
	}
}

func TestUtilizationBasicAndEdges(t *testing.T) {
	loc := ist(t)
	start := at(loc, 9, 0)
	end := at(loc, 10, 0)
	fx := newFixture(t, []domain.Session{
		session("S1", "BLR-001", "BLR-001-A", start, end, 10, 20, 8, domain.SessionCompleted),
	}, nil, nil)
	q := domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0)}
	snap, err := fx.svc.Compute(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	util, ok := snap.Network.Utilization()
	// Two connectors are in service and only one hour is charging, so the
	// network rate is 1h / 8h.
	if !ok || math.Abs(util-12.5) > 0.01 {
		t.Fatalf("util=%v ok=%v charging=%s available=%s", util, ok, snap.Network.Charging, snap.Network.Available)
	}
	q.StationID = "BLR-001"
	snap, err = fx.svc.Compute(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	util, ok = snap.Network.Utilization()
	if !ok || math.Abs(util-25) > 0.01 {
		t.Fatalf("station util=%v charging=%s available=%s", util, snap.Network.Charging, snap.Network.Available)
	}
	margin, pct, mok := snap.Network.Margin()
	if !mok || math.Abs(margin-120) > 0.01 || math.Abs(pct-60) > 0.01 {
		t.Fatalf("margin=%v pct=%v", margin, pct)
	}
	if snap.Network.Sessions != 1 || snap.Network.Energy != 10 {
		t.Fatalf("sessions=%d energy=%v", snap.Network.Sessions, snap.Network.Energy)
	}
}

func TestZeroSessionsAndZeroAvailableTime(t *testing.T) {
	loc := ist(t)
	fx := newFixture(t, nil, nil, nil)
	q := domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0), StationID: "BLR-001"}
	snap, err := fx.svc.Compute(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	util, ok := snap.Network.Utilization()
	if !ok || util != 0 {
		t.Fatalf("expected 0%% utilization with available time, got %v %v", util, ok)
	}
	up, ok := snap.Network.Uptime()
	if !ok || up != 100 {
		t.Fatalf("uptime=%v", up)
	}
	_, _, mok := snap.Network.Margin()
	if mok {
		t.Fatal("margin percent should be undefined when revenue is zero")
	}

	faultEnd := at(loc, 12, 0)
	fx = newFixture(t, nil, []domain.Fault{{
		ID: "F1", StationID: "BLR-001", ConnectorID: "BLR-001-A",
		Category: domain.FaultPowerSupply, Severity: "critical",
		Start: at(loc, 8, 0), End: &faultEnd, Description: "down",
	}}, nil)
	snap, err = fx.svc.Compute(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := snap.Network.Utilization(); ok {
		t.Fatal("full-window fault should leave utilization undefined")
	}
	up, ok = snap.Network.Uptime()
	if !ok || up != 0 {
		t.Fatalf("full outage uptime=%v ok=%v", up, ok)
	}

	fx = newFixture(t, nil, nil, []domain.MaintenanceWindow{{
		ID: "M1", StationID: "BLR-001", Start: at(loc, 8, 0), End: at(loc, 12, 0), Reason: "planned",
	}})
	snap, err = fx.svc.Compute(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := snap.Network.Utilization(); ok {
		t.Fatal("full-window maintenance should leave utilization undefined")
	}
	if _, ok := snap.Network.Uptime(); ok {
		t.Fatal("full-window maintenance should leave uptime undefined")
	}
}

func TestMaintenanceExcludedFromUtilization(t *testing.T) {
	loc := ist(t)
	// 4h window, 1h maintenance, 1h charging outside maintenance => available 3h, util 33.3%.
	fx := newFixture(t, []domain.Session{
		session("S1", "BLR-001", "BLR-001-A", at(loc, 10, 0), at(loc, 11, 0), 5, 20, 8, domain.SessionCompleted),
	}, nil, []domain.MaintenanceWindow{{
		ID: "M1", StationID: "BLR-001", Start: at(loc, 8, 0), End: at(loc, 9, 0), Reason: "planned",
	}})
	snap, err := fx.svc.Compute(context.Background(), domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0), StationID: "BLR-001"})
	if err != nil {
		t.Fatal(err)
	}
	util, ok := snap.Network.Utilization()
	if !ok || math.Abs(util-100.0/3.0) > 0.2 {
		t.Fatalf("util=%v charging=%s available=%s maint=%s", util, snap.Network.Charging, snap.Network.Available, snap.Network.Maintenance)
	}
	up, ok := snap.Network.Uptime()
	if !ok || math.Abs(up-100) > 0.01 {
		t.Fatalf("uptime should ignore planned maintenance, got %v", up)
	}
}

func TestFailedSessionDoesNotCountAsCharging(t *testing.T) {
	loc := ist(t)
	fx := newFixture(t, []domain.Session{
		session("S1", "BLR-001", "BLR-001-A", at(loc, 9, 0), at(loc, 9, 3), 0, 0, 0, domain.SessionFailed),
	}, nil, nil)
	snap, err := fx.svc.Compute(context.Background(), domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0), StationID: "BLR-001"})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Network.Charging != 0 {
		t.Fatalf("failed session charged for %s", snap.Network.Charging)
	}
	if snap.Network.Sessions != 1 || snap.Network.Failed != 1 || snap.Network.Energy != 0 {
		t.Fatalf("session stats %+v", snap.Network.Sessions)
	}
	util, ok := snap.Network.Utilization()
	if !ok || util != 0 {
		t.Fatalf("util=%v", util)
	}
}

func TestDateFilterAndStationSum(t *testing.T) {
	loc := ist(t)
	inside := session("S1", "BLR-001", "BLR-001-A", at(loc, 9, 0), at(loc, 10, 0), 10, 20, 8, domain.SessionCompleted)
	outside := session("S2", "BLR-001", "BLR-001-A", at(loc, 13, 0), at(loc, 14, 0), 10, 20, 8, domain.SessionCompleted)
	other := session("S3", "BLR-002", "BLR-002-A", at(loc, 9, 0), at(loc, 11, 0), 4, 20, 8, domain.SessionCompleted)
	other.ChargerType = domain.ChargerAC
	fx := newFixture(t, []domain.Session{inside, outside, other}, nil, nil)
	q := domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0)}
	snap, err := fx.svc.Compute(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Network.Sessions != 2 || math.Abs(snap.Network.Energy-14) > 0.01 {
		t.Fatalf("filtered energy=%v sessions=%d", snap.Network.Energy, snap.Network.Sessions)
	}
	var sumE float64
	var sumSessions int
	for _, sm := range snap.Stations {
		sumE += sm.Energy
		sumSessions += sm.Sessions
	}
	if math.Abs(sumE-snap.Network.Energy) > 0.01 || sumSessions != snap.Network.Sessions {
		t.Fatalf("station sum energy=%v network=%v sessions %d vs %d", sumE, snap.Network.Energy, sumSessions, snap.Network.Sessions)
	}
	var sumCharge time.Duration
	for _, sm := range snap.Stations {
		sumCharge += sm.Charging
	}
	if sumCharge != snap.Network.Charging {
		t.Fatalf("charging sum %s vs %s", sumCharge, snap.Network.Charging)
	}
}

func TestMTTRExcludesOngoingFault(t *testing.T) {
	loc := ist(t)
	resolvedEnd := at(loc, 12, 0)
	fx := newFixture(t, nil, []domain.Fault{
		{ID: "F1", StationID: "BLR-001", ConnectorID: "BLR-001-A", Category: domain.FaultConnector, Start: at(loc, 10, 0), End: &resolvedEnd},
		{ID: "F2", StationID: "BLR-001", ConnectorID: "BLR-001-A", Category: domain.FaultConnector, Start: at(loc, 16, 0), End: nil},
	}, nil)
	snap, err := fx.svc.Compute(context.Background(), domain.Query{From: at(loc, 8, 0), To: at(loc, 18, 0), StationID: "BLR-001"})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Network.RepairCount != 1 || snap.Network.RepairTime != 2*time.Hour {
		t.Fatalf("repairs=%d time=%s", snap.Network.RepairCount, snap.Network.RepairTime)
	}
	if snap.Network.OpenFaults != 1 || snap.Network.Failures != 2 {
		t.Fatalf("open=%d failures=%d", snap.Network.OpenFaults, snap.Network.Failures)
	}
	mtbf, ok := MTBF(snap.Network.Up, snap.Network.Failures)
	if !ok {
		t.Fatal("expected MTBF")
	}
	// 10h window, 2h resolved + 2h open = 4h down, 6h up, MTBF = 3h.
	if mtbf != 3*time.Hour {
		t.Fatalf("mtbf=%s up=%s down=%s", mtbf, snap.Network.Up, snap.Network.Downtime)
	}
	if snap.LongestOpen == nil || snap.LongestOpen.ID != "F2" {
		t.Fatalf("longest=%v", snap.LongestOpen)
	}
}

func TestMissingTelemetryPowerAndEmptyFilter(t *testing.T) {
	loc := ist(t)
	ses := session("S1", "BLR-001", "BLR-001-A", at(loc, 9, 0), at(loc, 10, 0), 10, 20, 8, domain.SessionCompleted)
	ses.PeakPowerKW = 0
	ses.AvgPowerKW = 0
	fx := newFixture(t, []domain.Session{ses}, nil, nil)
	rep, err := fx.svc.Energy(context.Background(), domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0), StationID: "BLR-001"})
	if err != nil {
		t.Fatal(err)
	}
	if rep.AvgChargingPowerKW == nil || math.Abs(*rep.AvgChargingPowerKW-10) > 0.05 {
		t.Fatalf("avg power=%v", rep.AvgChargingPowerKW)
	}
	if _, err := fx.svc.Compute(context.Background(), domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0), StationID: "NOPE"}); err != ErrUnknownStation {
		t.Fatalf("err=%v", err)
	}
	empty, err := fx.svc.Compute(context.Background(), domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0), ChargerType: "AC", StationID: "BLR-001"})
	if err != nil {
		t.Fatal(err)
	}
	if empty.ConnectorCount != 0 || empty.Network.Sessions != 0 {
		t.Fatalf("expected empty AC view at a DC station, connectors=%d", empty.ConnectorCount)
	}
	if _, err := fx.svc.Compute(context.Background(), domain.Query{From: at(loc, 12, 0), To: at(loc, 8, 0)}); err != ErrInvalidRange {
		t.Fatalf("err=%v", err)
	}
}

func TestChargingDuringFaultIsClipped(t *testing.T) {
	loc := ist(t)
	faultEnd := at(loc, 10, 0)
	fx := newFixture(t, []domain.Session{
		session("S1", "BLR-001", "BLR-001-A", at(loc, 9, 0), at(loc, 11, 0), 20, 20, 8, domain.SessionCompleted),
	}, []domain.Fault{{
		ID: "F1", StationID: "BLR-001", ConnectorID: "BLR-001-A", Category: domain.FaultConnector,
		Start: at(loc, 9, 0), End: &faultEnd,
	}}, nil)
	snap, err := fx.svc.Compute(context.Background(), domain.Query{From: at(loc, 8, 0), To: at(loc, 12, 0), StationID: "BLR-001"})
	if err != nil {
		t.Fatal(err)
	}
	util, ok := snap.Network.Utilization()
	if !ok || util > 100.01 {
		t.Fatalf("util=%v ok=%v charge=%s avail=%s", util, ok, snap.Network.Charging, snap.Network.Available)
	}
	// Fault 9-10, charge would have been 9-11, so only 10-11 remains. Available is 4h-1h=3h. Util ~33%.
	if math.Abs(util-100.0/3.0) > 1 {
		t.Fatalf("clipped util=%v charge=%s avail=%s", util, snap.Network.Charging, snap.Network.Available)
	}
}
