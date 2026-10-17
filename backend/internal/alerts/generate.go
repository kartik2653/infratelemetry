// Package alerts builds operational alerts from calculated windows and keeps
// no state of its own. The store persists acknowledge and resolve in memory.
package alerts

import (
	"fmt"
	"sort"
	"time"

	"chargingdatainfra/backend/internal/analytics"
	"chargingdatainfra/backend/internal/domain"
)

// Generate builds the demo alert list for the trailing windows.
// last30, last14, last7, and prior must cover the same station set.
// prior is the window immediately before last7, used for drop detection.
func Generate(asOf time.Time, faults []domain.Fault, last30, last14, last7, prior *analytics.Snapshot) []domain.Alert {
	var out []domain.Alert
	out = append(out, offlineAlerts(asOf, faults, names(last30))...)
	if last30 != nil {
		for _, sm := range last30.Stations {
			if sm.ConnectorFaults >= 4 {
				out = append(out, domain.Alert{
					ID: "ALT-RPT-" + sm.Station.ID, Severity: string(domain.SeverityWarning),
					Category: "repeated_connector_failure", StationID: sm.Station.ID, StationName: sm.Station.Name,
					Timestamp:   asOf,
					Title:       fmt.Sprintf("Repeated connector faults at %s", sm.Station.ID),
					Description: fmt.Sprintf("%s recorded %d connector faults in the trailing 30 days, with %.0f connector-hours of unplanned downtime.", sm.Station.Name, sm.ConnectorFaults, sm.Downtime.Hours()),
					Metric:      "connector_faults_30d", ObservedValue: fmt.Sprintf("%d", sm.ConnectorFaults),
					RecommendedAction: "Inspect the connector that is failing before another driver is turned away. A reset without a parts check is likely to recur.",
					Status:            string(domain.AlertOpen),
				})
			}
			up, upOK := sm.Uptime()
			if sm.Downtime >= 48*time.Hour && (!upOK || up < 92) {
				sev := string(domain.SeverityWarning)
				if upOK && up < 85 {
					sev = string(domain.SeverityCritical)
				}
				out = append(out, domain.Alert{
					ID: "ALT-DOWN-" + sm.Station.ID, Severity: sev,
					Category: "high_downtime", StationID: sm.Station.ID, StationName: sm.Station.Name,
					Timestamp:   asOf,
					Title:       fmt.Sprintf("High downtime at %s", sm.Station.ID),
					Description: fmt.Sprintf("Unplanned downtime is %.0f connector-hours and uptime is %s over 30 days. Scheduled maintenance is excluded.", sm.Downtime.Hours(), pct(up, upOK)),
					Metric:      "uptime_pct_30d", ObservedValue: pct(up, upOK),
					RecommendedAction: "Rank this site above routine visits. Confirm whether the lost hours are one long supply fault or many short connector faults.",
					Status:            string(domain.AlertOpen),
				})
			}
			if sm.RevenueAtRisk >= 4000 {
				sev := string(domain.SeverityWarning)
				if sm.RevenueAtRisk >= 15000 {
					sev = string(domain.SeverityCritical)
				}
				out = append(out, domain.Alert{
					ID: "ALT-RISK-" + sm.Station.ID, Severity: sev,
					Category: "revenue_at_risk", StationID: sm.Station.ID, StationName: sm.Station.Name,
					Timestamp:   asOf,
					Title:       fmt.Sprintf("Revenue at risk at %s", sm.Station.ID),
					Description: fmt.Sprintf("Estimated revenue at risk is ₹%.0f. The estimate is downtime connector-hours times revenue per available hour, and it assumes demand would have matched recent utilization. It is not recorded lost revenue.", sm.RevenueAtRisk),
					Metric:      "revenue_at_risk_inr", ObservedValue: fmt.Sprintf("₹%.0f", sm.RevenueAtRisk),
					RecommendedAction: "Use the estimate to order repairs. Do not post it as a financial loss.",
					Status:            string(domain.AlertOpen),
				})
			}
		}
	}
	if last7 != nil && prior != nil {
		priorByID := indexStations(prior)
		for _, sm := range last7.Stations {
			prev, ok := priorByID[sm.Station.ID]
			if !ok {
				continue
			}
			recent, rok := sm.Utilization()
			before, bok := prev.Utilization()
			if rok && bok && before >= 12 && recent < before*0.6 {
				out = append(out, domain.Alert{
					ID: "ALT-UDROP-" + sm.Station.ID, Severity: string(domain.SeverityWarning),
					Category: "utilization_drop", StationID: sm.Station.ID, StationName: sm.Station.Name,
					Timestamp:   asOf,
					Title:       fmt.Sprintf("Utilization dropped at %s", sm.Station.ID),
					Description: fmt.Sprintf("Utilization over the last 7 days is %.1f%%, down from %.1f%% in the previous 23 days. Both figures exclude maintenance and fault time.", recent, before),
					Metric:      "utilization_pct", ObservedValue: fmt.Sprintf("%.1f%% vs %.1f%%", recent, before),
					RecommendedAction: "Check for a new outage, a tariff change, or a nearby site opening before treating this as a demand shift.",
					Status:            string(domain.AlertOpen),
				})
			}
			re, reOK := sm.AvgEnergy()
			pe, peOK := prev.AvgEnergy()
			if reOK && peOK && sm.Sessions >= 8 && prev.Sessions >= 8 && pe > 5 && re < pe*0.6 {
				out = append(out, domain.Alert{
					ID: "ALT-ENERGY-" + sm.Station.ID, Severity: string(domain.SeverityWarning),
					Category: "energy_anomaly", StationID: sm.Station.ID, StationName: sm.Station.Name,
					Timestamp:   asOf,
					Title:       fmt.Sprintf("Energy per session fell at %s", sm.Station.ID),
					Description: fmt.Sprintf("Average energy per session is %.1f kWh over 7 days, versus %.1f kWh in the previous 23 days. Session counts are %d and %d.", re, pe, sm.Sessions, prev.Sessions),
					Metric:      "avg_energy_kwh", ObservedValue: fmt.Sprintf("%.1f kWh vs %.1f kWh", re, pe),
					RecommendedAction: "Check for power derating, vehicles ending sessions early, or a connector that is no longer delivering its usual power.",
					Status:            string(domain.AlertOpen),
				})
			}
			if sm.LongSessions >= 4 {
				out = append(out, domain.Alert{
					ID: "ALT-LONG-" + sm.Station.ID, Severity: string(domain.SeverityInfo),
					Category: "abnormal_session_duration", StationID: sm.Station.ID, StationName: sm.Station.Name,
					Timestamp:   asOf,
					Title:       fmt.Sprintf("Long sessions at %s", sm.Station.ID),
					Description: fmt.Sprintf("%d sessions in 7 days exceeded 90 minutes on DC or 5 hours on AC. Plug time can include idle time after charging stops.", sm.LongSessions),
					Metric:      "long_sessions_7d", ObservedValue: fmt.Sprintf("%d", sm.LongSessions),
					RecommendedAction: "See whether vehicles are left plugged in after the charge completes. Utilization itself ignores that idle tail.",
					Status:            string(domain.AlertOpen),
				})
			}
		}
	}
	if last14 != nil {
		for _, sm := range last14.Stations {
			if sm.Sessions < 20 || sm.PaymentFails < 6 {
				continue
			}
			rate := float64(sm.PaymentFails) / float64(sm.Sessions) * 100
			if rate < 8 {
				continue
			}
			out = append(out, domain.Alert{
				ID: "ALT-PAY-" + sm.Station.ID, Severity: string(domain.SeverityWarning),
				Category: "repeated_payment_failure", StationID: sm.Station.ID, StationName: sm.Station.Name,
				Timestamp:   asOf,
				Title:       fmt.Sprintf("Repeated payment failures at %s", sm.Station.ID),
				Description: fmt.Sprintf("%d of %d sessions in 14 days ended as payment_failure (%.0f%%). Those attempts delivered no energy.", sm.PaymentFails, sm.Sessions, rate),
				Metric:      "payment_failures_14d", ObservedValue: fmt.Sprintf("%d (%.0f%%)", sm.PaymentFails, rate),
				RecommendedAction: "Check the payment terminal and the authorisation path. Compare with a nearby station on the same tariff.",
				Status:            string(domain.AlertOpen),
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		si, sj := sevRank(out[i].Severity), sevRank(out[j].Severity)
		if si != sj {
			return si < sj
		}
		if !out[i].Timestamp.Equal(out[j].Timestamp) {
			return out[i].Timestamp.After(out[j].Timestamp)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func offlineAlerts(asOf time.Time, faults []domain.Fault, names map[string]string) []domain.Alert {
	var out []domain.Alert
	for _, f := range faults {
		if !f.Ongoing(asOf) || f.Start.After(asOf) {
			continue
		}
		age := asOf.Sub(f.Start)
		sev := string(domain.SeverityWarning)
		if domain.HardwareFault(f.Category) || age >= 4*time.Hour {
			sev = string(domain.SeverityCritical)
		}
		if !domain.HardwareFault(f.Category) && age < 4*time.Hour {
			sev = string(domain.SeverityWarning)
		}
		name := names[f.StationID]
		if name == "" {
			name = f.StationID
		}
		where := f.StationID
		if f.ConnectorID != "" {
			where = f.ConnectorID
		}
		out = append(out, domain.Alert{
			ID: "ALT-OFF-" + f.ID, Severity: sev, Category: "charger_offline",
			StationID: f.StationID, StationName: name, ConnectorID: f.ConnectorID,
			Timestamp:   f.Start,
			Title:       fmt.Sprintf("%s is offline (%s)", where, labelCategory(f.Category)),
			Description: fmt.Sprintf("%s Fault has been open for %.1f hours. %s", f.Description, age.Hours(), "Status will show faulted for hardware faults and unavailable for communication or payment faults."),
			Metric:      "open_fault_hours", ObservedValue: fmt.Sprintf("%.1f hours", age.Hours()),
			RecommendedAction: actionFor(f.Category),
			Status:            string(domain.AlertOpen),
		})
	}
	return out
}

func actionFor(c domain.FaultCategory) string {
	switch c {
	case domain.FaultConnector:
		return "Dispatch a technician to the connector. Do not return it to service on a reset alone if the contactor or latch has already failed repeatedly."
	case domain.FaultPowerSupply:
		return "Confirm the site supply and breaker before replacing charger parts."
	case domain.FaultTemperature:
		return "Keep the cabinet out of service until the temperature alarm clears and the cooling path has been checked."
	case domain.FaultEmergency:
		return "Verify the emergency stop on site and restart only after a physical inspection."
	case domain.FaultCommunication:
		return "Check the network link and charger modem. The bay should stay unavailable until it reports status again."
	case domain.FaultPayment:
		return "Test the payment terminal with a known-good method. Leave the connector unavailable while authorisations fail."
	default:
		return "Confirm the fault on site and update the alert once the connector is back in service."
	}
}

func labelCategory(c domain.FaultCategory) string {
	switch c {
	case domain.FaultCommunication:
		return "communication failure"
	case domain.FaultConnector:
		return "connector failure"
	case domain.FaultPowerSupply:
		return "power supply failure"
	case domain.FaultTemperature:
		return "temperature alarm"
	case domain.FaultPayment:
		return "payment error"
	case domain.FaultEmergency:
		return "emergency shutdown"
	default:
		return string(c)
	}
}

func names(s *analytics.Snapshot) map[string]string {
	out := map[string]string{}
	if s == nil {
		return out
	}
	for _, sm := range s.Stations {
		out[sm.Station.ID] = sm.Station.Name
	}
	return out
}

func indexStations(s *analytics.Snapshot) map[string]analytics.StationMetric {
	out := map[string]analytics.StationMetric{}
	if s == nil {
		return out
	}
	for _, sm := range s.Stations {
		out[sm.Station.ID] = sm
	}
	return out
}

func pct(v float64, ok bool) string {
	if !ok {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", v)
}

func sevRank(s string) int {
	switch s {
	case string(domain.SeverityCritical):
		return 0
	case string(domain.SeverityWarning):
		return 1
	default:
		return 2
	}
}
