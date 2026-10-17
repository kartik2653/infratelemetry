package analytics

import (
	"testing"
	"time"
)

func TestMergeAndSubtract(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	h := func(hours int) time.Time { return base.Add(time.Duration(hours) * time.Hour) }

	merged := Merge([]Interval{
		{h(0), h(2)},
		{h(2), h(4)},
		{h(5), h(6)},
		{h(1), h(3)},
	})
	if len(merged) != 2 {
		t.Fatalf("merged = %+v", merged)
	}
	if got := TotalDuration(merged); got != 5*time.Hour {
		t.Fatalf("duration = %s", got)
	}

	left := Subtract([]Interval{{h(0), h(10)}}, []Interval{{h(2), h(4)}, {h(8), h(12)}})
	if got := TotalDuration(left); got != 6*time.Hour {
		t.Fatalf("subtract duration = %s, parts %+v", got, left)
	}
}

func TestUtilizationAndUptimeEdges(t *testing.T) {
	if _, ok := UtilizationPct(time.Hour, 0); ok {
		t.Fatal("zero available time should be undefined")
	}
	pct, ok := UtilizationPct(0, 4*time.Hour)
	if !ok || pct != 0 {
		t.Fatalf("zero sessions: got %v %v", pct, ok)
	}
	pct, ok = UtilizationPct(time.Hour, 4*time.Hour)
	if !ok || pct != 25 {
		t.Fatalf("util = %v", pct)
	}
	pct, ok = UtilizationPct(5*time.Hour, 4*time.Hour)
	if !ok || pct != 100 {
		t.Fatalf("capped util = %v", pct)
	}

	if _, ok := UptimePct(0, 0); ok {
		t.Fatal("all-maintenance window should be undefined")
	}
	up, ok := UptimePct(18*time.Hour, 24*time.Hour)
	if !ok || up != 75 {
		t.Fatalf("uptime = %v", up)
	}
	up, ok = UptimePct(0, 24*time.Hour)
	if !ok || up != 0 {
		t.Fatalf("full outage uptime = %v", up)
	}
}

func TestGrossMargin(t *testing.T) {
	margin, pct, ok := GrossMargin(200, 80)
	if margin != 120 || !ok || pct != 60 {
		t.Fatalf("margin=%v pct=%v ok=%v", margin, pct, ok)
	}
	margin, _, ok = GrossMargin(0, 10)
	if margin != -10 || ok {
		t.Fatalf("zero revenue: margin=%v ok=%v", margin, ok)
	}
	margin, pct, ok = GrossMargin(0, 0)
	if margin != 0 || ok || pct != 0 {
		t.Fatalf("empty finance: %v %v %v", margin, pct, ok)
	}
}

func TestMTTRAndMTBF(t *testing.T) {
	mttr, ok := MTTR([]time.Duration{2 * time.Hour, 4 * time.Hour})
	if !ok || mttr != 3*time.Hour {
		t.Fatalf("mttr = %s", mttr)
	}
	if _, ok := MTTR(nil); ok {
		t.Fatal("expected undefined MTTR")
	}
	if _, ok := MTTR([]time.Duration{0}); ok {
		t.Fatal("zero-length repairs are not an MTTR")
	}
	mtbf, ok := MTBF(100*time.Hour, 4)
	if !ok || mtbf != 25*time.Hour {
		t.Fatalf("mtbf = %s", mtbf)
	}
	if _, ok := MTBF(100*time.Hour, 0); ok {
		t.Fatal("zero failures should not produce an infinite MTBF")
	}
}

func TestObservedPowerFallback(t *testing.T) {
	avg, peak, ok := ObservedPower(10, 3600, 12)
	if !ok || avg != 10 || peak != 12 {
		t.Fatalf("avg=%v peak=%v", avg, peak)
	}
	avg, peak, ok = ObservedPower(10, 3600, 0)
	if !ok || avg != 10 || peak != 10 {
		t.Fatalf("missing peak fallback avg=%v peak=%v", avg, peak)
	}
	if _, _, ok := ObservedPower(0, 3600, 5); ok {
		t.Fatal("zero energy should not invent power")
	}
	if _, _, ok := ObservedPower(5, 0, 5); ok {
		t.Fatal("zero charge duration should not invent power")
	}
}
