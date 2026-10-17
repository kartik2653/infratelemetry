package domain

// Metric definitions are returned with API payloads so the UI does not
// invent formulas that could drift from the calculations.
const (
	Disclaimer = "Simulated demo data for a fictional Bengaluru charging network. These are not live chargers, not a real operator, and not measured grid load."

	FormulaUtilization = "Utilization = charging time ÷ available operating time × 100. Available operating time is the selected window minus scheduled maintenance and unplanned fault downtime. Charging time counts completed, interrupted, and in-progress sessions, clipped so it cannot fall inside excluded periods. Plug-in time after charging has stopped does not count. Failed attempts that delivered no energy do not count. Utilization is not availability and it is not uptime."

	FormulaUptime = "Uptime = time outside unplanned faults ÷ (window − scheduled maintenance) × 100. Planned maintenance is removed from both sides. Fault time that overlaps maintenance stays in the maintenance exclusion and is not double-counted as unplanned downtime."

	FormulaMargin = "Gross margin = charging revenue − estimated electricity procurement cost. Revenue is session energy × the driver tariff at session start. Cost is session energy × that station's procurement rate at session start. Peak tariff and peak procurement apply at 08:00–11:00 and 18:00–21:00 IST; all other hours are off-peak. This is not net profit."

	FormulaRevenueAtRisk = "Revenue at risk = unplanned downtime connector-hours × (station revenue ÷ available connector-hours) in the same window. It assumes demand during the outage would have matched observed utilization while the charger was available. If a station has no available time, the network revenue per available connector-hour is used. This is an estimate, not recorded lost revenue."

	FormulaMTTR = "MTTR = total duration of resolved faults that overlap the window ÷ the number of those faults. Duration is the full recorded repair, not clipped to the window. A station-wide fault counts once. Open faults are excluded from MTTR."

	FormulaMTBF = "MTBF = connector operating hours ÷ fault incidents that overlap the window. Operating hours exclude planned maintenance and unplanned downtime. A station-wide incident counts once. MTBF is omitted when the window has no faults. It describes this synthetic history and is not a forecast."

	FormulaSessionAttribution = "Session counts, energy, revenue, electricity cost, and average duration include sessions that start inside the window. The whole session is attributed to its start. Utilization, uptime, and hourly energy throughput use the minutes that actually fall inside each bucket. Hourly throughput therefore matches KPI energy only when every included session also ends inside the window."

	FormulaPower = "Average charging power is session energy ÷ charging hours. Peak power is the maximum within-session reading stored on the session. Hourly telemetry samples are hourly averages derived from those sessions, not a separate utility meter. Rated-versus-observed comparisons use nameplate kW. They do not imply grid import, transformer loading, or electrical efficiency."
)

// IncludedCostComponents are the only costs in gross margin.
func IncludedCostComponents() []string {
	return []string{
		"Driver tariff billed on energy (peak 08:00–11:00 and 18:00–21:00 IST, otherwise off-peak), stored on each session",
		"Electricity procurement cost at the station peak or off-peak rate for the session start hour",
	}
}

// ExcludedCostComponents are deliberately absent from gross margin.
func ExcludedCostComponents() []string {
	return []string{
		"Site rent and common-area charges",
		"Maintenance contracts and spare parts",
		"Staffing and field-service labour",
		"Payment gateway and roaming fees",
		"Taxes, depreciation, and insurance",
		"Demand charges, reactive power, and transformer losses",
	}
}

// StatusNote explains why connector status ignores the date filter.
const StatusNote = "Connector status counts are the current state at the demo clock, limited to connectors that match the station and charger-type filters. They are not a historical replay of the selected dates. Energy, revenue, utilization, uptime, sessions, and faults do follow the date range."
