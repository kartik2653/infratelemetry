package analytics

import (
	"sort"
	"time"
)

// Interval is a half-open time range [Start, End).
type Interval struct {
	Start time.Time
	End   time.Time
}

func (iv Interval) Duration() time.Duration {
	if !iv.End.After(iv.Start) {
		return 0
	}
	return iv.End.Sub(iv.Start)
}

func (iv Interval) Valid() bool {
	return iv.End.After(iv.Start)
}

// Merge sorts and unions overlapping or touching intervals.
func Merge(in []Interval) []Interval {
	if len(in) == 0 {
		return nil
	}
	sorted := append([]Interval(nil), in...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start.Equal(sorted[j].Start) {
			return sorted[i].End.Before(sorted[j].End)
		}
		return sorted[i].Start.Before(sorted[j].Start)
	})
	out := make([]Interval, 0, len(sorted))
	cur := sorted[0]
	for _, iv := range sorted[1:] {
		if !iv.Valid() {
			continue
		}
		if !cur.Valid() {
			cur = iv
			continue
		}
		if !iv.Start.After(cur.End) {
			if iv.End.After(cur.End) {
				cur.End = iv.End
			}
			continue
		}
		out = append(out, cur)
		cur = iv
	}
	if cur.Valid() {
		out = append(out, cur)
	}
	return out
}

// Subtract removes cuts from base. Both inputs may be unsorted.
// The result is merged and contains only time that is in base and not in cuts.
func Subtract(base, cuts []Interval) []Interval {
	bases := Merge(base)
	cutm := Merge(cuts)
	if len(bases) == 0 {
		return nil
	}
	var out []Interval
	for _, b := range bases {
		cur := []Interval{b}
		for _, c := range cutm {
			next := make([]Interval, 0, len(cur)*2)
			for _, iv := range cur {
				next = append(next, subtractOne(iv, c)...)
			}
			cur = next
			if len(cur) == 0 {
				break
			}
		}
		out = append(out, cur...)
	}
	return Merge(out)
}

func subtractOne(a, c Interval) []Interval {
	if !a.Valid() {
		return nil
	}
	if !c.Valid() || !c.End.After(a.Start) || !c.Start.Before(a.End) {
		return []Interval{a}
	}
	var out []Interval
	if c.Start.After(a.Start) {
		out = append(out, Interval{Start: a.Start, End: c.Start})
	}
	if c.End.Before(a.End) {
		out = append(out, Interval{Start: c.End, End: a.End})
	}
	return out
}

// Overlap returns the intersection duration of two intervals.
func Overlap(a, b Interval) time.Duration {
	s := a.Start
	if b.Start.After(s) {
		s = b.Start
	}
	e := a.End
	if b.End.Before(e) {
		e = b.End
	}
	if !e.After(s) {
		return 0
	}
	return e.Sub(s)
}

// OverlapSum is the time in slot that is covered by merged intervals.
func OverlapSum(slot Interval, ivs []Interval) time.Duration {
	if !slot.Valid() {
		return 0
	}
	var d time.Duration
	for _, iv := range ivs {
		d += Overlap(slot, iv)
	}
	if d > slot.Duration() {
		return slot.Duration()
	}
	return d
}

// TotalDuration sums interval lengths. Intervals must not overlap.
func TotalDuration(ivs []Interval) time.Duration {
	var d time.Duration
	for _, iv := range ivs {
		d += iv.Duration()
	}
	return d
}
