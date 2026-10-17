package domain

import "math"

// Round2 rounds to paise (two decimal places) half away from zero.
func Round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Round1 rounds to one decimal, used for percentages in API payloads.
func Round1(v float64) float64 {
	return math.Round(v*10) / 10
}
