package analytics

import "time"

// UtilizationPct returns charging / available × 100.
// ok is false when there is no available operating time, so callers can
// distinguish "never charged" (0%) from "no time in service" (undefined).
func UtilizationPct(charging, available time.Duration) (float64, bool) {
	if available <= 0 {
		return 0, false
	}
	if charging < 0 {
		charging = 0
	}
	if charging > available {
		charging = available
	}
	return float64(charging) / float64(available) * 100, true
}

// UptimePct returns uptime / scheduled × 100.
// scheduled should already exclude planned maintenance.
// ok is false when the entire window was planned maintenance.
func UptimePct(up, scheduled time.Duration) (float64, bool) {
	if scheduled <= 0 {
		return 0, false
	}
	if up < 0 {
		up = 0
	}
	if up > scheduled {
		up = scheduled
	}
	return float64(up) / float64(scheduled) * 100, true
}

// GrossMargin returns revenue − electricity cost, and the margin as a
// percentage of revenue. The percentage is undefined when revenue is zero.
func GrossMargin(revenue, electricityCost float64) (margin float64, pct float64, pctOK bool) {
	margin = revenue - electricityCost
	if revenue <= 0 {
		return margin, 0, false
	}
	return margin, margin / revenue * 100, true
}

// MTTR returns the mean repair duration. Ongoing faults must be omitted
// by the caller. ok is false when there are no resolved repairs.
func MTTR(resolved []time.Duration) (time.Duration, bool) {
	if len(resolved) == 0 {
		return 0, false
	}
	var sum time.Duration
	n := 0
	for _, d := range resolved {
		if d <= 0 {
			continue
		}
		sum += d
		n++
	}
	if n == 0 {
		return 0, false
	}
	return sum / time.Duration(n), true
}

// MTBF returns operating time / failure count.
// ok is false when there are no failures; an empty history is not reported
// as an infinite MTBF.
func MTBF(operating time.Duration, failures int) (time.Duration, bool) {
	if failures <= 0 || operating < 0 {
		return 0, false
	}
	return operating / time.Duration(failures), true
}

// Ratio returns num/den. ok is false when den is zero.
func Ratio(num, den float64) (float64, bool) {
	if den <= 0 {
		return 0, false
	}
	return num / den, true
}

// ObservedPower derives average kW from energy and charging duration.
// A stored peak of zero falls back to the average so missing telemetry
// does not invent a higher peak. ok is false when the session cannot
// support a power reading.
func ObservedPower(energyKWh float64, chargeSec int64, storedPeakKW float64) (avgKW, peakKW float64, ok bool) {
	if energyKWh <= 0 || chargeSec <= 0 {
		return 0, 0, false
	}
	hours := float64(chargeSec) / 3600
	avgKW = energyKWh / hours
	peakKW = storedPeakKW
	if peakKW <= 0 {
		peakKW = avgKW
	}
	return avgKW, peakKW, true
}
