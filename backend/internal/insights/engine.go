// Package insights is a rule engine. It does not call a language model.
// Every recommendation is a comparison against the snapshot that was just calculated.
package insights

import (
	"fmt"
	"sort"
	"time"
)

const (
	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityMedium   = "medium"
	PriorityLow      = "low"
)

type Insight struct {
	ID                string   `json:"id"`
	Category          string   `json:"category"`
	Priority          string   `json:"priority"`
	Title             string   `json:"title"`
	Evidence          string   `json:"evidence"`
	WhyItMatters      string   `json:"why_it_matters"`
	RecommendedAction string   `json:"recommended_action"`
	EstimatedImpact   string   `json:"estimated_impact"`
	StationIDs        []string `json:"station_ids"`
	StationNames      []string `json:"station_names"`
	Metric            string   `json:"metric"`
	ObservedValue     string   `json:"observed_value"`
}

type Network struct {
	Utilization   *float64
	MarginPct     *float64
	MarginPerKWh  float64
	DowntimeHours float64
}

type Station struct {
	ID, Name, Area           string
	HasDC                    bool
	MaxRatedKW               float64
	ConnectorCount           int
	Utilization              *float64
	Uptime                   *float64
	MarginPct                *float64
	MarginPerKWh             float64
	Revenue                  float64
	Energy                   float64
	Sessions                 int
	DowntimeHours            float64
	ConnectorFaults          int
	Faults                   int
	PaymentFails             int
	AvgPlugMin               float64
	AvgChargeMin             float64
	SessionsByHour           [24]int
	PeakHour                 int
	PeakHourUtil             *float64
	EnergyPerConnectorPerDay float64
	RevenueAtRisk            float64
	WeekdayPerDay            float64
	WeekendPerDay            float64
	WeekdaySessions          int
	WeekendSessions          int
	AvailableHours           float64
}

// Generate ranks explainable recommendations for the supplied facts.
func Generate(net Network, stations []Station, from, to time.Time) []Insight {
	var out []Insight
	out = append(out, lowUtilization(net, stations)...)
	out = append(out, lowMargin(net, stations)...)
	out = append(out, repeatedFaults(stations)...)
	out = append(out, highDowntime(stations)...)
	out = append(out, downtimeConcentration(net, stations)...)
	out = append(out, peakConcentration(stations)...)
	out = append(out, capacityTight(stations)...)
	out = append(out, paymentFailures(stations)...)
	out = append(out, irregularDemand(stations)...)
	out = append(out, revenueAtRisk(stations)...)
	out = append(out, idlePlug(stations)...)
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := rank(out[i].Priority), rank(out[j].Priority)
		if ri != rj {
			return ri < rj
		}
		return out[i].ID < out[j].ID
	})
	if len(out) > 24 {
		out = out[:24]
	}
	return out
}

func rank(p string) int {
	switch p {
	case PriorityCritical:
		return 0
	case PriorityHigh:
		return 1
	case PriorityMedium:
		return 2
	default:
		return 3
	}
}

func lowUtilization(net Network, stations []Station) []Insight {
	if net.Utilization == nil {
		return nil
	}
	var out []Insight
	for _, st := range stations {
		if st.Utilization == nil || st.AvailableHours < 48 {
			continue
		}
		if *st.Utilization >= 10 || *st.Utilization >= *net.Utilization*0.55 {
			continue
		}
		category := "low_utilization"
		priority := PriorityMedium
		asset := "charging assets"
		if st.HasDC && st.MaxRatedKW >= 50 {
			category = "underutilized_fast_charging"
			priority = PriorityHigh
			asset = "DC fast-charging assets"
		}
		out = append(out, Insight{
			ID: category + "-" + st.ID, Category: category, Priority: priority,
			Title:             fmt.Sprintf("%s utilization is %s, against a network average of %s", st.ID, pct(st.Utilization), pct(net.Utilization)),
			Evidence:          fmt.Sprintf("%s (%s) delivered %.0f kWh per connector per day across %d connectors. Available operating time in the window was %.0f hours. Utilization uses charging time only and already excludes maintenance and fault downtime.", st.Name, st.Area, st.EnergyPerConnectorPerDay, st.ConnectorCount, st.AvailableHours),
			WhyItMatters:      fmt.Sprintf("Low utilization means these %s are tying up capital without a matching amount of energy sales. Adding capacity at this site would put more hardware against demonstrated weak demand.", asset),
			RecommendedAction: "Review site demand, signage, tariffs, and nearby alternatives before investing in additional connectors. Compare operating cost with the energy actually sold.",
			EstimatedImpact:   fmt.Sprintf("Energy delivery is %.0f kWh per connector per day at this site. That is the demand the current history supports; this prototype does not forecast a utilization recovery.", st.EnergyPerConnectorPerDay),
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "utilization_pct", ObservedValue: pct(st.Utilization),
		})
	}
	return out
}

func lowMargin(net Network, stations []Station) []Insight {
	if net.MarginPct == nil || net.Utilization == nil {
		return nil
	}
	var out []Insight
	for _, st := range stations {
		if st.Utilization == nil || st.MarginPct == nil || st.Energy < 50 || st.Revenue < 1000 {
			continue
		}
		if *st.Utilization < *net.Utilization {
			continue
		}
		if *st.MarginPct+12 >= *net.MarginPct {
			continue
		}
		gapPerKWh := net.MarginPerKWh - st.MarginPerKWh
		gap := gapPerKWh * st.Energy
		impact := "The gap is descriptive of this window's tariff and procurement rates. It is not a forecast and it is not net profit."
		if gap > 0 {
			impact = fmt.Sprintf("If gross margin per kWh had matched the network (₹%.2f/kWh versus ₹%.2f/kWh here), gross margin on the %.0f kWh sold would have been about ₹%.0f higher. Rent, maintenance, and fees are still excluded.", net.MarginPerKWh, st.MarginPerKWh, st.Energy, gap)
		}
		out = append(out, Insight{
			ID: "low-margin-" + st.ID, Category: "high_utilization_low_margin", Priority: PriorityHigh,
			Title:             fmt.Sprintf("%s utilization is %s, above the network average of %s, but gross margin is %s versus %s", st.ID, pct(st.Utilization), pct(net.Utilization), pct(st.MarginPct), pct(net.MarginPct)),
			Evidence:          fmt.Sprintf("Revenue ₹%.0f, electricity cost leaves a gross margin of %s (₹%.2f per kWh). Network gross margin is %s (₹%.2f per kWh). Only driver tariffs and procurement rates are included.", st.Revenue, pct(st.MarginPct), st.MarginPerKWh, pct(net.MarginPct), net.MarginPerKWh),
			WhyItMatters:      "A busy station can still be a poor use of power if the tariff does not cover procurement. High utilization hides that until margin is looked at separately from uptime.",
			RecommendedAction: "Compare this station's peak and off-peak procurement rates with its driver tariff. Revisit the supply contract or the tariff before adding connectors.",
			EstimatedImpact:   impact,
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "gross_margin_pct", ObservedValue: pct(st.MarginPct),
		})
	}
	return out
}

func repeatedFaults(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.ConnectorFaults < 4 {
			continue
		}
		out = append(out, Insight{
			ID: "faults-" + st.ID, Category: "repeated_faults", Priority: PriorityHigh,
			Title:             fmt.Sprintf("%s recorded %d connector faults in this window", st.ID, st.ConnectorFaults),
			Evidence:          fmt.Sprintf("%s lost %.0f connector-hours to unplanned downtime. Uptime for the station is %s. Connector faults are a subset of all fault categories.", st.Name, st.DowntimeHours, pct(st.Uptime)),
			WhyItMatters:      "A connector that fails repeatedly drops sessions and trains drivers to avoid the site. Each visit also consumes field-service time that this margin model does not cost.",
			RecommendedAction: "Inspect the affected connector hardware and recent work orders before the next failure. Do not treat a reset as a repair when the same category keeps recurring.",
			EstimatedImpact:   fmt.Sprintf("Estimated revenue at risk from this station's unplanned downtime is ₹%.0f in the window, assuming demand would have matched recent utilization. That figure is not recorded lost revenue.", st.RevenueAtRisk),
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "connector_faults", ObservedValue: fmt.Sprintf("%d", st.ConnectorFaults),
		})
	}
	return out
}

func highDowntime(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.DowntimeHours < 48 {
			continue
		}
		if st.Uptime != nil && *st.Uptime >= 92 && st.ConnectorFaults >= 4 {
			continue
		}
		priority := PriorityHigh
		if st.Uptime != nil && *st.Uptime < 85 {
			priority = PriorityCritical
		}
		out = append(out, Insight{
			ID: "downtime-" + st.ID, Category: "high_downtime", Priority: priority,
			Title:             fmt.Sprintf("%s lost %.0f connector-hours to unplanned downtime", st.ID, st.DowntimeHours),
			Evidence:          fmt.Sprintf("Uptime is %s after scheduled maintenance is excluded. Fault incidents recorded at the station: %d.", pct(st.Uptime), st.Faults),
			WhyItMatters:      "Downtime removes the asset from both energy sales and driver trust. Utilization of the hours that remain can look acceptable while the site is still unreliable.",
			RecommendedAction: "Prioritise the open or longest faults at this site. Confirm whether the outage was supply-side, connector hardware, or communications before dispatching the wrong part.",
			EstimatedImpact:   fmt.Sprintf("Estimated revenue at risk is ₹%.0f for this window under the assumption that lost hours would have earned the station's observed revenue per available hour.", st.RevenueAtRisk),
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "downtime_hours", ObservedValue: fmt.Sprintf("%.0f connector-hours", st.DowntimeHours),
		})
	}
	return out
}

func downtimeConcentration(net Network, stations []Station) []Insight {
	if net.DowntimeHours < 24 {
		return nil
	}
	sorted := append([]Station(nil), stations...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].DowntimeHours > sorted[j].DowntimeHours })
	var top []Station
	for _, st := range sorted {
		if st.DowntimeHours < 8 || len(top) == 3 {
			continue
		}
		if len(top) < 3 {
			top = append(top, st)
		}
	}
	// A single site, or every site in the filter, is not a concentration.
	if len(top) < 2 || len(stations) <= len(top) {
		return nil
	}
	sum := 0.0
	ids := make([]string, 0, len(top))
	names := make([]string, 0, len(top))
	for _, st := range top {
		sum += st.DowntimeHours
		ids = append(ids, st.ID)
		names = append(names, st.Name)
	}
	share := sum / net.DowntimeHours * 100
	if share < 55 {
		return nil
	}
	return []Insight{{
		ID: "downtime-concentration", Category: "downtime_concentration", Priority: PriorityHigh,
		Title:             fmt.Sprintf("%d stations account for %.0f%% of network downtime", len(top), share),
		Evidence:          fmt.Sprintf("%s lost %.0f of %.0f unplanned connector-hours in the network. Scheduled maintenance is not included in these hours.", joinIDs(ids), sum, net.DowntimeHours),
		WhyItMatters:      "Maintenance effort spread evenly across the network will miss the sites that create most of the lost operating time.",
		RecommendedAction: "Put these sites at the front of the field-service queue and review whether the same fault category is repeating.",
		EstimatedImpact:   fmt.Sprintf("They hold %.0f%% of unplanned connector-hours. Revenue at risk is estimated per station elsewhere; this insight does not add those estimates together a second time as a forecast.", share),
		StationIDs:        ids, StationNames: names,
		Metric: "downtime_share_pct", ObservedValue: fmt.Sprintf("%.0f%%", share),
	}}
}

func peakConcentration(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.Sessions < 25 {
			continue
		}
		start, n, share := best3(st.SessionsByHour)
		if share < 0.42 {
			continue
		}
		out = append(out, Insight{
			ID: "peak-hours-" + st.ID, Category: "peak_concentration", Priority: PriorityMedium,
			Title:             fmt.Sprintf("Most sessions at %s start between %02d:00 and %02d:00", st.ID, start, start+3),
			Evidence:          fmt.Sprintf("%d of %d sessions (%.0f%%) start in that three-hour window. Session starts are attributed to the clock hour they begin, in IST.", n, st.Sessions, share*100),
			WhyItMatters:      "A short peak can congest bays and leave the rest of the day idle. Tariff or dwell changes are a possible response; this history does not say how drivers would react.",
			RecommendedAction: "Check whether peak-hour pricing or a bay-time limit would move sessions, or whether the peak needs another connector. Do not treat this as a measured price elasticity.",
			EstimatedImpact:   "No revenue impact is estimated. The prototype does not model how demand would shift if the tariff changed.",
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "session_start_share", ObservedValue: fmt.Sprintf("%.0f%% in %02d:00–%02d:00", share*100, start, start+3),
		})
	}
	return out
}

func capacityTight(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.PeakHourUtil == nil || st.Utilization == nil {
			continue
		}
		if *st.PeakHourUtil < 55 || *st.Utilization < 12 {
			continue
		}
		out = append(out, Insight{
			ID: "capacity-" + st.ID, Category: "capacity_bottleneck", Priority: PriorityMedium,
			Title:             fmt.Sprintf("%s is tight around %02d:00 (%s of available time charging)", st.ID, st.PeakHour, pct(st.PeakHourUtil)),
			Evidence:          fmt.Sprintf("All-day utilization is %s. The busiest clock hour with at least 10 hours of available time in the window is %02d:00 at %s. This is bay occupancy, not grid load.", pct(st.Utilization), st.PeakHour, pct(st.PeakHourUtil)),
			WhyItMatters:      "Drivers who arrive in that hour are more likely to find a bay busy even when the daily average looks comfortable.",
			RecommendedAction: "Review queue complaints and session start times before adding a connector. A peak tariff is an alternative only if the commercial team accepts that elasticity is unknown here.",
			EstimatedImpact:   "No additional energy or revenue is forecast. The figure above is observed occupancy.",
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "peak_hour_utilization_pct", ObservedValue: pct(st.PeakHourUtil),
		})
	}
	return out
}

func paymentFailures(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.Sessions < 20 || st.PaymentFails < 8 {
			continue
		}
		rate := float64(st.PaymentFails) / float64(st.Sessions) * 100
		if rate < 8 {
			continue
		}
		out = append(out, Insight{
			ID: "payments-" + st.ID, Category: "payment_failures", Priority: PriorityMedium,
			Title:             fmt.Sprintf("%s failed %d payments (%.0f%% of sessions)", st.ID, st.PaymentFails, rate),
			Evidence:          fmt.Sprintf("%d sessions started with a payment_failure termination out of %d sessions. These attempts deliver no energy and are excluded from charging time.", st.PaymentFails, st.Sessions),
			WhyItMatters:      "Drivers who cannot authorise a session leave without charging. The lost sale is not fully visible in utilization because the attempt is short.",
			RecommendedAction: "Check the payment terminal, acquirer, and RFID or app path at this site. Compare with stations that use the same tariff but fail less often.",
			EstimatedImpact:   "No recovered-revenue figure is claimed. Failed attempts in this dataset record zero energy and zero revenue.",
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "payment_failure_rate", ObservedValue: fmt.Sprintf("%d failures, %.0f%%", st.PaymentFails, rate),
		})
	}
	return out
}

func irregularDemand(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.WeekendSessions < 12 || st.WeekdayPerDay <= 0 || st.WeekendPerDay <= 0 {
			continue
		}
		if st.WeekendPerDay < st.WeekdayPerDay*2.5 {
			continue
		}
		out = append(out, Insight{
			ID: "irregular-" + st.ID, Category: "irregular_demand", Priority: PriorityLow,
			Title:             fmt.Sprintf("%s demand is concentrated on weekends", st.ID),
			Evidence:          fmt.Sprintf("Weekend sessions average %.1f per day versus %.1f on weekdays (%.0f weekend sessions, %.0f weekday sessions).", st.WeekendPerDay, st.WeekdayPerDay, float64(st.WeekendSessions), float64(st.WeekdaySessions)),
			WhyItMatters:      "A staffing pattern or tariff copied from a commuter site will miss when this location is actually busy, and a weekday site visit will understate demand.",
			RecommendedAction: "Judge this site on weekend throughput. Avoid using a weekday snapshot to decide whether the location has failed.",
			EstimatedImpact:   "No forward demand is estimated.",
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "weekend_to_weekday_sessions", ObservedValue: fmt.Sprintf("%.1f×", st.WeekendPerDay/st.WeekdayPerDay),
		})
	}
	return out
}

func revenueAtRisk(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.RevenueAtRisk < 5000 {
			continue
		}
		priority := PriorityMedium
		if st.RevenueAtRisk >= 15000 {
			priority = PriorityHigh
		}
		out = append(out, Insight{
			ID: "risk-" + st.ID, Category: "revenue_at_risk", Priority: priority,
			Title:             fmt.Sprintf("%s has about ₹%.0f of revenue at risk from downtime", st.ID, st.RevenueAtRisk),
			Evidence:          fmt.Sprintf("Unplanned downtime was %.0f connector-hours. The estimate multiplies those hours by this station's revenue per available connector-hour in the same window.", st.DowntimeHours),
			WhyItMatters:      "The estimate shows which outages were expensive if demand stayed at the observed rate. It is not cash that was invoiced and then lost.",
			RecommendedAction: "Use the figure to rank repairs. Do not book it as lost revenue in a financial statement.",
			EstimatedImpact:   fmt.Sprintf("₹%.0f estimated, assuming outage hours would have earned the same revenue per available hour the station earned while it was up.", st.RevenueAtRisk),
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "revenue_at_risk_inr", ObservedValue: fmt.Sprintf("₹%.0f", st.RevenueAtRisk),
		})
	}
	return out
}

func idlePlug(stations []Station) []Insight {
	var out []Insight
	for _, st := range stations {
		if st.Sessions < 15 || st.AvgChargeMin <= 0 || st.AvgPlugMin <= 0 {
			continue
		}
		if st.AvgPlugMin < st.AvgChargeMin*1.45 {
			continue
		}
		out = append(out, Insight{
			ID: "idle-plug-" + st.ID, Category: "idle_plugin", Priority: PriorityLow,
			Title:             fmt.Sprintf("Vehicles at %s stay plugged in after charging", st.ID),
			Evidence:          fmt.Sprintf("Average plug time is %.0f minutes and average charging time is %.0f minutes across %d sessions. Utilization counts charging time only, so the bay can be blocked without the utilization rate moving as much.", st.AvgPlugMin, st.AvgChargeMin, st.Sessions),
			WhyItMatters:      "A finished vehicle holding a bay blocks the next session. That lost throughput does not show up as a fault.",
			RecommendedAction: "Review idle fees or a session-end notification. This dataset does not include a fee schedule, so no idle-fee revenue is calculated.",
			EstimatedImpact:   "No recovered-session count is estimated.",
			StationIDs:        []string{st.ID}, StationNames: []string{st.Name},
			Metric: "plug_to_charge_ratio", ObservedValue: fmt.Sprintf("%.0f min plugged, %.0f min charging", st.AvgPlugMin, st.AvgChargeMin),
		})
	}
	return out
}

func best3(hours [24]int) (start, count int, share float64) {
	total := 0
	for _, n := range hours {
		total += n
	}
	bestN := -1
	for h := 0; h <= 21; h++ {
		n := hours[h] + hours[h+1] + hours[h+2]
		if n > bestN {
			bestN = n
			start = h
		}
	}
	if total > 0 && bestN > 0 {
		share = float64(bestN) / float64(total)
	}
	return start, bestN, share
}

func pct(p *float64) string {
	if p == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", *p)
}

func joinIDs(ids []string) string {
	switch len(ids) {
	case 0:
		return ""
	case 1:
		return ids[0]
	case 2:
		return ids[0] + " and " + ids[1]
	default:
		out := ""
		for i, id := range ids {
			if i == len(ids)-1 {
				out += "and " + id
			} else if i == 0 {
				out += id
			} else {
				out += ", " + id
			}
			if i < len(ids)-2 {
				// commas already handled above except we need comma after first items
			}
		}
		// Rebuild cleanly.
		s := ids[0]
		for i := 1; i < len(ids)-1; i++ {
			s += ", " + ids[i]
		}
		return s + ", and " + ids[len(ids)-1]
	}
}
