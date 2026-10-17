package analytics

import (
	"fmt"
	"sort"
	"time"

	"chargingdatainfra/backend/internal/domain"
)

type DailyPoint struct {
	Date               string   `json:"date"`
	EnergyKWh          float64  `json:"energy_kwh"`
	RevenueINR         float64  `json:"revenue_inr"`
	ElectricityCostINR float64  `json:"electricity_cost_inr"`
	GrossMarginINR     float64  `json:"gross_margin_inr"`
	Sessions           int      `json:"sessions"`
	UtilizationPct     *float64 `json:"utilization_pct"`
	UptimePct          *float64 `json:"uptime_pct"`
	DowntimeHours      float64  `json:"downtime_hours"`
}

type HourPoint struct {
	Hour           int      `json:"hour"`
	Label          string   `json:"label"`
	UtilizationPct *float64 `json:"utilization_pct"`
	SessionStarts  int      `json:"session_starts"`
	EnergyKWh      float64  `json:"energy_kwh"`
	RevenueINR     float64  `json:"revenue_inr"`
}

type WeekdayPoint struct {
	Weekday   int     `json:"weekday"`
	Label     string  `json:"label"`
	Sessions  int     `json:"sessions"`
	EnergyKWh float64 `json:"energy_kwh"`
}

type KPISet struct {
	StationCount           int      `json:"station_count"`
	ConnectorCount         int      `json:"connector_count"`
	UptimePct              *float64 `json:"uptime_pct"`
	UtilizationPct         *float64 `json:"utilization_pct"`
	SessionCount           int      `json:"session_count"`
	EnergyKWh              float64  `json:"energy_kwh"`
	RevenueINR             float64  `json:"revenue_inr"`
	ElectricityCostINR     float64  `json:"electricity_cost_inr"`
	GrossMarginINR         float64  `json:"gross_margin_inr"`
	GrossMarginPct         *float64 `json:"gross_margin_pct"`
	AvgSessionDurationMin  *float64 `json:"avg_session_duration_min"`
	AvgEnergyPerSessionKWh *float64 `json:"avg_energy_per_session_kwh"`
	ChargingHours          float64  `json:"charging_hours"`
	AvailableHours         float64  `json:"available_hours"`
	DowntimeHours          float64  `json:"downtime_hours"`
}

type SessionBreakdown struct {
	Completed          int      `json:"completed"`
	Failed             int      `json:"failed"`
	Interrupted        int      `json:"interrupted"`
	InProgress         int      `json:"in_progress"`
	SuccessRatePct     *float64 `json:"success_rate_pct"`
	FailureRatePct     *float64 `json:"failure_rate_pct"`
	InterruptedRatePct *float64 `json:"interrupted_rate_pct"`
}

type Summary struct {
	OperatorName       string            `json:"operator_name"`
	NetworkName        string            `json:"network_name"`
	Disclaimer         string            `json:"disclaimer"`
	AsOf               time.Time         `json:"as_of"`
	Timezone           string            `json:"timezone"`
	From               time.Time         `json:"from"`
	To                 time.Time         `json:"to"`
	KPIs               KPISet            `json:"kpis"`
	SessionStats       SessionBreakdown  `json:"session_stats"`
	ConnectorsByStatus map[string]int    `json:"connectors_by_status"`
	Definitions        map[string]string `json:"definitions"`
	Assumptions        []string          `json:"assumptions"`
	IncludedCosts      []string          `json:"included_cost_components"`
	ExcludedCosts      []string          `json:"excluded_cost_components"`
	TopStations        []StationRow      `json:"top_stations"`
	BottomStations     []StationRow      `json:"bottom_stations"`
}

type Trends struct {
	Daily     []DailyPoint   `json:"daily"`
	Hourly    []HourPoint    `json:"hourly"`
	ByWeekday []WeekdayPoint `json:"by_weekday"`
	Notes     []string       `json:"notes"`
}

type StationRow struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Area               string         `json:"area"`
	Address            string         `json:"address"`
	City               string         `json:"city"`
	Latitude           float64        `json:"latitude"`
	Longitude          float64        `json:"longitude"`
	ChargerTypes       []string       `json:"charger_types"`
	ConnectorCount     int            `json:"connector_count"`
	RatedPowerKW       float64        `json:"rated_power_kw"`
	MaxPowerKW         float64        `json:"max_power_kw"`
	Status             string         `json:"status"`
	ConnectorsByStatus map[string]int `json:"connectors_by_status"`
	UtilizationPct     *float64       `json:"utilization_pct"`
	UptimePct          *float64       `json:"uptime_pct"`
	EnergyKWh          float64        `json:"energy_kwh"`
	RevenueINR         float64        `json:"revenue_inr"`
	ElectricityCostINR float64        `json:"electricity_cost_inr"`
	GrossMarginINR     float64        `json:"gross_margin_inr"`
	GrossMarginPct     *float64       `json:"gross_margin_pct"`
	SessionCount       int            `json:"session_count"`
	DowntimeHours      float64        `json:"downtime_hours"`
}

type ConnectorRow struct {
	ID             string   `json:"id"`
	Label          string   `json:"label"`
	ChargerType    string   `json:"charger_type"`
	Standard       string   `json:"standard"`
	RatedPowerKW   float64  `json:"rated_power_kw"`
	Status         string   `json:"status"`
	UtilizationPct *float64 `json:"utilization_pct"`
	UptimePct      *float64 `json:"uptime_pct"`
	EnergyKWh      float64  `json:"energy_kwh"`
	RevenueINR     float64  `json:"revenue_inr"`
	GrossMarginINR float64  `json:"gross_margin_inr"`
	GrossMarginPct *float64 `json:"gross_margin_pct"`
	SessionCount   int      `json:"session_count"`
	DowntimeHours  float64  `json:"downtime_hours"`
}

type FaultDTO struct {
	ID            string     `json:"id"`
	StationID     string     `json:"station_id"`
	ConnectorID   string     `json:"connector_id"`
	Category      string     `json:"category"`
	Severity      string     `json:"severity"`
	Start         time.Time  `json:"start"`
	End           *time.Time `json:"end"`
	Ongoing       bool       `json:"ongoing"`
	DurationHours float64    `json:"duration_hours"`
	Description   string     `json:"description"`
}

type StationDetail struct {
	Station      StationRow        `json:"station"`
	KPIs         KPISet            `json:"kpis"`
	SessionStats SessionBreakdown  `json:"session_stats"`
	Connectors   []ConnectorRow    `json:"connectors"`
	Daily        []DailyPoint      `json:"daily"`
	Hourly       []HourPoint       `json:"hourly"`
	Faults       []FaultDTO        `json:"faults"`
	Definitions  map[string]string `json:"definitions"`
	Assumptions  []string          `json:"assumptions"`
	Tariffs      TariffView        `json:"tariffs"`
}

type TariffView struct {
	PeakTariffINRPerKWh        float64 `json:"peak_tariff_inr_per_kwh"`
	OffPeakTariffINRPerKWh     float64 `json:"off_peak_tariff_inr_per_kwh"`
	PeakEnergyCostINRPerKWh    float64 `json:"peak_energy_cost_inr_per_kwh"`
	OffPeakEnergyCostINRPerKWh float64 `json:"off_peak_energy_cost_inr_per_kwh"`
	PeakHours                  string  `json:"peak_hours"`
}

type NamedUtil struct {
	Key            string   `json:"key"`
	Label          string   `json:"label"`
	UtilizationPct *float64 `json:"utilization_pct"`
	ChargingHours  float64  `json:"charging_hours"`
	AvailableHours float64  `json:"available_hours"`
	IdleHours      float64  `json:"idle_hours"`
	EnergyKWh      float64  `json:"energy_kwh"`
	ConnectorCount int      `json:"connector_count"`
	SessionCount   int      `json:"session_count"`
}

type LowUtilAsset struct {
	StationID      string   `json:"station_id"`
	StationName    string   `json:"station_name"`
	ConnectorID    string   `json:"connector_id"`
	ChargerType    string   `json:"charger_type"`
	RatedPowerKW   float64  `json:"rated_power_kw"`
	UtilizationPct *float64 `json:"utilization_pct"`
	EnergyKWh      float64  `json:"energy_kwh"`
	AvailableHours float64  `json:"available_hours"`
}

type UtilizationReport struct {
	Formula                  string            `json:"formula"`
	Notes                    []string          `json:"notes"`
	UtilizationPct           *float64          `json:"utilization_pct"`
	ChargingHours            float64           `json:"charging_hours"`
	AvailableHours           float64           `json:"available_hours"`
	IdleHours                float64           `json:"idle_hours"`
	ExcludedMaintenanceHours float64           `json:"excluded_maintenance_hours"`
	ExcludedDowntimeHours    float64           `json:"excluded_downtime_hours"`
	EnergyPerConnectorPerDay float64           `json:"energy_per_connector_per_day"`
	UnderutilizedAssets      int               `json:"underutilized_assets"`
	ByStation                []NamedUtil       `json:"by_station"`
	ByChargerType            []NamedUtil       `json:"by_charger_type"`
	ByPowerRating            []NamedUtil       `json:"by_power_rating"`
	Peak                     NamedUtil         `json:"peak"`
	OffPeak                  NamedUtil         `json:"off_peak"`
	Daily                    []DailyPoint      `json:"daily"`
	Hourly                   []HourPoint       `json:"hourly"`
	LowUtilization           []LowUtilAsset    `json:"low_utilization"`
	Definitions              map[string]string `json:"definitions"`
}

type StationFinance struct {
	StationID           string   `json:"station_id"`
	StationName         string   `json:"station_name"`
	Area                string   `json:"area"`
	RevenueINR          float64  `json:"revenue_inr"`
	ElectricityCostINR  float64  `json:"electricity_cost_inr"`
	GrossMarginINR      float64  `json:"gross_margin_inr"`
	GrossMarginPct      *float64 `json:"gross_margin_pct"`
	EnergyKWh           float64  `json:"energy_kwh"`
	Sessions            int      `json:"sessions"`
	RevenuePerSession   *float64 `json:"revenue_per_session_inr"`
	RevenuePerConnector float64  `json:"revenue_per_connector_inr"`
	RevenuePerKWh       *float64 `json:"revenue_per_kwh"`
	EnergyCostPerKWh    *float64 `json:"energy_cost_per_kwh"`
	MarginPerKWh        *float64 `json:"margin_per_kwh"`
	UtilizationPct      *float64 `json:"utilization_pct"`
	ConnectorCount      int      `json:"connector_count"`
}

type RevenueReport struct {
	Assumptions            []string          `json:"assumptions"`
	IncludedCostComponents []string          `json:"included_cost_components"`
	ExcludedCostComponents []string          `json:"excluded_cost_components"`
	RevenueINR             float64           `json:"revenue_inr"`
	ElectricityCostINR     float64           `json:"electricity_cost_inr"`
	GrossMarginINR         float64           `json:"gross_margin_inr"`
	GrossMarginPct         *float64          `json:"gross_margin_pct"`
	AvgRevenuePerSession   *float64          `json:"avg_revenue_per_session_inr"`
	RevenuePerConnector    float64           `json:"revenue_per_connector_inr"`
	RevenuePerKWh          *float64          `json:"revenue_per_kwh"`
	EnergyCostPerKWh       *float64          `json:"energy_cost_per_kwh"`
	MarginPerKWh           *float64          `json:"margin_per_kwh"`
	ByStation              []StationFinance  `json:"by_station"`
	Daily                  []DailyPoint      `json:"daily"`
	TopStations            []StationFinance  `json:"top_stations"`
	BottomStations         []StationFinance  `json:"bottom_stations"`
	Definitions            map[string]string `json:"definitions"`
}

type PowerCompare struct {
	ChargerType     string   `json:"charger_type"`
	RatedPowerKW    float64  `json:"rated_power_kw"`
	Label           string   `json:"label"`
	AvgObservedKW   *float64 `json:"avg_observed_kw"`
	PeakObservedKW  *float64 `json:"peak_observed_kw"`
	ObservedToRated *float64 `json:"observed_to_rated"`
	EnergyKWh       float64  `json:"energy_kwh"`
	ConnectorCount  int      `json:"connector_count"`
}

type DemandSite struct {
	StationID                string   `json:"station_id"`
	StationName              string   `json:"station_name"`
	Area                     string   `json:"area"`
	EnergyKWh                float64  `json:"energy_kwh"`
	EnergyPerConnectorPerDay float64  `json:"energy_per_connector_per_day"`
	UtilizationPct           *float64 `json:"utilization_pct"`
	PeakHour                 int      `json:"peak_hour"`
	PeakHourLabel            string   `json:"peak_hour_label"`
	PeakHourUtilizationPct   *float64 `json:"peak_hour_utilization_pct"`
	Note                     string   `json:"note"`
}

type EnergyReport struct {
	Notes               []string          `json:"notes"`
	TotalEnergyKWh      float64           `json:"total_energy_kwh"`
	AvgSessionEnergyKWh *float64          `json:"avg_session_energy_kwh"`
	AvgChargingPowerKW  *float64          `json:"avg_charging_power_kw"`
	PeakObservedPowerKW *float64          `json:"peak_observed_power_kw"`
	ByStation           []DemandSite      `json:"by_station"`
	ByConnector         []ConnectorEnergy `json:"by_connector"`
	ByChargerType       []NamedUtil       `json:"by_charger_type"`
	Daily               []DailyPoint      `json:"daily"`
	Hourly              []HourPoint       `json:"hourly"`
	RatedVsObserved     []PowerCompare    `json:"rated_vs_observed"`
	HighDemand          []DemandSite      `json:"high_demand"`
	LowDemand           []DemandSite      `json:"low_demand"`
	Definitions         map[string]string `json:"definitions"`
}

type ConnectorEnergy struct {
	ConnectorID  string  `json:"connector_id"`
	StationID    string  `json:"station_id"`
	StationName  string  `json:"station_name"`
	ChargerType  string  `json:"charger_type"`
	RatedPowerKW float64 `json:"rated_power_kw"`
	EnergyKWh    float64 `json:"energy_kwh"`
	SessionCount int     `json:"session_count"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type Outage struct {
	FaultID       string    `json:"fault_id"`
	StationID     string    `json:"station_id"`
	ConnectorID   string    `json:"connector_id"`
	Category      string    `json:"category"`
	Start         time.Time `json:"start"`
	DurationHours float64   `json:"duration_hours"`
	Description   string    `json:"description"`
}

type StationReliability struct {
	StationID        string   `json:"station_id"`
	StationName      string   `json:"station_name"`
	Area             string   `json:"area"`
	UptimePct        *float64 `json:"uptime_pct"`
	DowntimeHours    float64  `json:"downtime_hours"`
	FaultCount       int      `json:"fault_count"`
	OpenFaultCount   int      `json:"open_fault_count"`
	MTTRHours        *float64 `json:"mttr_hours"`
	MTBFHours        *float64 `json:"mtbf_hours"`
	RevenueAtRiskINR float64  `json:"revenue_at_risk_inr"`
	RateSource       string   `json:"revenue_at_risk_rate_source"`
}

type ReliabilityReport struct {
	Formula                  string               `json:"formula"`
	Notes                    []string             `json:"notes"`
	UptimePct                *float64             `json:"uptime_pct"`
	DowntimeHours            float64              `json:"downtime_hours"`
	MaintenanceHoursExcluded float64              `json:"maintenance_hours_excluded"`
	FaultCount               int                  `json:"fault_count"`
	OpenFaultCount           int                  `json:"open_fault_count"`
	MTTRHours                *float64             `json:"mttr_hours"`
	MTBFHours                *float64             `json:"mtbf_hours"`
	LongestOngoing           *Outage              `json:"longest_ongoing"`
	RevenueAtRiskINR         float64              `json:"revenue_at_risk_inr"`
	RevenueAtRiskAssumption  string               `json:"revenue_at_risk_assumption"`
	ByStation                []StationReliability `json:"by_station"`
	ByCategory               []CategoryCount      `json:"by_category"`
	Faults                   []FaultDTO           `json:"faults"`
	Definitions              map[string]string    `json:"definitions"`
}

type ReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type PeakWindow struct {
	StartHour int     `json:"start_hour"`
	EndHour   int     `json:"end_hour"`
	Label     string  `json:"label"`
	Sessions  int     `json:"sessions"`
	SharePct  float64 `json:"share_pct"`
}

type SessionReport struct {
	Notes              []string          `json:"notes"`
	SessionCount       int               `json:"session_count"`
	Completed          int               `json:"completed"`
	Failed             int               `json:"failed"`
	Interrupted        int               `json:"interrupted"`
	InProgress         int               `json:"in_progress"`
	SuccessRatePct     *float64          `json:"success_rate_pct"`
	FailureRatePct     *float64          `json:"failure_rate_pct"`
	InterruptedRatePct *float64          `json:"interrupted_rate_pct"`
	AvgDurationMin     *float64          `json:"avg_duration_min"`
	AvgEnergyKWh       *float64          `json:"avg_energy_kwh"`
	ByHour             []HourPoint       `json:"by_hour"`
	ByWeekday          []WeekdayPoint    `json:"by_weekday"`
	PeakWindows        []PeakWindow      `json:"peak_windows"`
	Daily              []DailyPoint      `json:"daily"`
	FailedByReason     []ReasonCount     `json:"failed_by_reason"`
	Definitions        map[string]string `json:"definitions"`
}

type SessionDTO struct {
	ID                 string    `json:"id"`
	StationID          string    `json:"station_id"`
	ConnectorID        string    `json:"connector_id"`
	Start              time.Time `json:"start"`
	End                time.Time `json:"end"`
	DurationMin        float64   `json:"duration_min"`
	ChargingMin        float64   `json:"charging_min"`
	EnergyKWh          float64   `json:"energy_kwh"`
	AvgPowerKW         float64   `json:"avg_power_kw"`
	PeakPowerKW        float64   `json:"peak_power_kw"`
	TariffINRPerKWh    float64   `json:"tariff_inr_per_kwh"`
	RevenueINR         float64   `json:"revenue_inr"`
	ElectricityCostINR float64   `json:"electricity_cost_inr"`
	GrossMarginINR     float64   `json:"gross_margin_inr"`
	Status             string    `json:"status"`
	TerminationReason  string    `json:"termination_reason"`
	ChargerType        string    `json:"charger_type"`
	RatedPowerKW       float64   `json:"rated_power_kw"`
}

func definitions() map[string]string {
	return map[string]string{
		"utilization":         domain.FormulaUtilization,
		"uptime":              domain.FormulaUptime,
		"gross_margin":        domain.FormulaMargin,
		"revenue_at_risk":     domain.FormulaRevenueAtRisk,
		"mttr":                domain.FormulaMTTR,
		"mtbf":                domain.FormulaMTBF,
		"session_attribution": domain.FormulaSessionAttribution,
		"power":               domain.FormulaPower,
		"connector_status":    domain.StatusNote,
	}
}

func assumptionLines() []string {
	return []string{
		domain.FormulaMargin,
		domain.FormulaSessionAttribution,
		domain.StatusNote,
		"Gross margin is not net profit. " + joinCosts(),
	}
}

func joinCosts() string {
	return "Excluded: rent, maintenance, staffing, payment fees, taxes, depreciation, and demand charges."
}

func r2(v float64) float64 { return domain.Round2(v) }

func r1ptr(v float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	x := domain.Round1(v)
	return &x
}

func r2ptr(v float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	x := domain.Round2(v)
	return &x
}

func hrs(d time.Duration) float64 { return r2(d.Hours()) }

func kpis(t Totals, stations, connectors int) KPISet {
	margin, marginPct, marginOK := t.Margin()
	util, utilOK := t.Utilization()
	up, upOK := t.Uptime()
	avgDur, durOK := t.AvgDurationMin()
	avgE, eOK := t.AvgEnergy()
	return KPISet{
		StationCount: stations, ConnectorCount: connectors,
		UptimePct: r1ptr(up, upOK), UtilizationPct: r1ptr(util, utilOK),
		SessionCount: t.Sessions, EnergyKWh: r2(t.Energy),
		RevenueINR: r2(t.Revenue), ElectricityCostINR: r2(t.Cost),
		GrossMarginINR: r2(margin), GrossMarginPct: r1ptr(marginPct, marginOK),
		AvgSessionDurationMin: r1ptr(avgDur, durOK), AvgEnergyPerSessionKWh: r2ptr(avgE, eOK),
		ChargingHours: hrs(t.Charging), AvailableHours: hrs(t.Available), DowntimeHours: hrs(t.Downtime),
	}
}

func sessionsOf(t Totals) SessionBreakdown {
	return SessionBreakdown{
		Completed: t.Completed, Failed: t.Failed, Interrupted: t.Interrupted, InProgress: t.InProgress,
		SuccessRatePct:     ratePtr(t.Completed, t.Sessions),
		FailureRatePct:     ratePtr(t.Failed, t.Sessions),
		InterruptedRatePct: ratePtr(t.Interrupted, t.Sessions),
	}
}

func ratePtr(n, d int) *float64 {
	if d <= 0 {
		return nil
	}
	return r1ptr(float64(n)/float64(d)*100, true)
}

func dailyPoints(days []DayStat) []DailyPoint {
	out := make([]DailyPoint, 0, len(days))
	for _, d := range days {
		util, uok := d.Utilization()
		up, pok := d.Uptime()
		out = append(out, DailyPoint{
			Date: d.Date, EnergyKWh: r2(d.Energy), RevenueINR: r2(d.Revenue),
			ElectricityCostINR: r2(d.Cost), GrossMarginINR: r2(d.Revenue - d.Cost),
			Sessions: d.Sessions, UtilizationPct: r1ptr(util, uok), UptimePct: r1ptr(up, pok),
			DowntimeHours: hrs(d.Downtime),
		})
	}
	return out
}

func hourPoints(t Totals) []HourPoint {
	out := make([]HourPoint, 24)
	for h := 0; h < 24; h++ {
		util, ok := UtilizationPct(t.HourCharging[h], t.HourAvailable[h])
		out[h] = HourPoint{
			Hour: h, Label: fmt.Sprintf("%02d:00", h),
			UtilizationPct: r1ptr(util, ok),
			SessionStarts:  t.HourSessionStarts[h],
			EnergyKWh:      r2(t.HourEnergy[h]),
			RevenueINR:     r2(t.HourRevenue[h]),
		}
	}
	return out
}

func weekdayPoints(t Totals) []WeekdayPoint {
	labels := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	order := []int{1, 2, 3, 4, 5, 6, 0}
	out := make([]WeekdayPoint, 0, 7)
	for _, wd := range order {
		out = append(out, WeekdayPoint{
			Weekday: wd, Label: labels[wd], Sessions: t.DOWSessions[wd], EnergyKWh: r2(t.DOWEnergy[wd]),
		})
	}
	return out
}

func stationRow(sm StationMetric) StationRow {
	margin, pct, ok := sm.Margin()
	util, uok := sm.Utilization()
	up, pok := sm.Uptime()
	types := append([]string(nil), sm.ChargerTypes...)
	sort.Strings(types)
	return StationRow{
		ID: sm.Station.ID, Name: sm.Station.Name, Area: sm.Station.Area,
		Address: sm.Station.Address, City: sm.Station.City,
		Latitude: sm.Station.Latitude, Longitude: sm.Station.Longitude,
		ChargerTypes: types, ConnectorCount: len(sm.Connectors),
		RatedPowerKW: r2(sm.SumRatedKW), MaxPowerKW: r2(sm.MaxRatedKW),
		Status: sm.WorstStatus, ConnectorsByStatus: sm.StatusCounts,
		UtilizationPct: r1ptr(util, uok), UptimePct: r1ptr(up, pok),
		EnergyKWh: r2(sm.Energy), RevenueINR: r2(sm.Revenue), ElectricityCostINR: r2(sm.Cost),
		GrossMarginINR: r2(margin), GrossMarginPct: r1ptr(pct, ok),
		SessionCount: sm.Sessions, DowntimeHours: hrs(sm.Downtime),
	}
}

func connectorRow(cm ConnectorMetric) ConnectorRow {
	margin, pct, ok := cm.Margin()
	util, uok := cm.Utilization()
	up, pok := cm.Uptime()
	return ConnectorRow{
		ID: cm.Connector.ID, Label: cm.Connector.Label,
		ChargerType: string(cm.Connector.ChargerType), Standard: cm.Connector.Standard,
		RatedPowerKW: cm.Connector.RatedPowerKW, Status: string(cm.Connector.Status),
		UtilizationPct: r1ptr(util, uok), UptimePct: r1ptr(up, pok),
		EnergyKWh: r2(cm.Energy), RevenueINR: r2(cm.Revenue),
		GrossMarginINR: r2(margin), GrossMarginPct: r1ptr(pct, ok),
		SessionCount: cm.Sessions, DowntimeHours: hrs(cm.Downtime),
	}
}

func faultDTO(f domain.Fault, asOf time.Time) FaultDTO {
	end := f.FaultEnd(asOf)
	ongoing := f.Ongoing(asOf)
	var endPtr *time.Time
	if !ongoing && f.End != nil {
		cp := *f.End
		endPtr = &cp
	}
	return FaultDTO{
		ID: f.ID, StationID: f.StationID, ConnectorID: f.ConnectorID,
		Category: string(f.Category), Severity: f.Severity,
		Start: f.Start, End: endPtr, Ongoing: ongoing,
		DurationHours: r2(end.Sub(f.Start).Hours()), Description: f.Description,
	}
}

func financeRow(sm StationMetric) StationFinance {
	margin, pct, ok := sm.Margin()
	revKWh, revOK := sm.RevenuePerKWh()
	costKWh, costOK := sm.CostPerKWh()
	var marginPer *float64
	if revOK && costOK {
		marginPer = r2ptr(revKWh-costKWh, true)
	}
	n := len(sm.Connectors)
	perConn := 0.0
	if n > 0 {
		perConn = sm.Revenue / float64(n)
	}
	util, uok := sm.Utilization()
	return StationFinance{
		StationID: sm.Station.ID, StationName: sm.Station.Name, Area: sm.Station.Area,
		RevenueINR: r2(sm.Revenue), ElectricityCostINR: r2(sm.Cost),
		GrossMarginINR: r2(margin), GrossMarginPct: r1ptr(pct, ok),
		EnergyKWh: r2(sm.Energy), Sessions: sm.Sessions,
		RevenuePerSession:   r2ptr(safeDiv(sm.Revenue, float64(sm.Sessions))),
		RevenuePerConnector: r2(perConn),
		RevenuePerKWh:       r2ptr(revKWh, revOK), EnergyCostPerKWh: r2ptr(costKWh, costOK),
		MarginPerKWh: marginPer, UtilizationPct: r1ptr(util, uok), ConnectorCount: n,
	}
}

func safeDiv(n, d float64) (float64, bool) {
	if d <= 0 {
		return 0, false
	}
	return n / d, true
}

func (s *Snapshot) rows() []StationRow {
	out := make([]StationRow, 0, len(s.Stations))
	for _, sm := range s.Stations {
		out = append(out, stationRow(sm))
	}
	return out
}

func summaryFrom(s *Snapshot) Summary {
	rows := s.rows()
	byRev := append([]StationRow(nil), rows...)
	sort.Slice(byRev, func(i, j int) bool {
		if byRev[i].RevenueINR == byRev[j].RevenueINR {
			return byRev[i].ID < byRev[j].ID
		}
		return byRev[i].RevenueINR > byRev[j].RevenueINR
	})
	byUtil := append([]StationRow(nil), rows...)
	sort.Slice(byUtil, func(i, j int) bool {
		ui, uj := -1.0, -1.0
		if byUtil[i].UtilizationPct != nil {
			ui = *byUtil[i].UtilizationPct
		}
		if byUtil[j].UtilizationPct != nil {
			uj = *byUtil[j].UtilizationPct
		}
		if ui == uj {
			return byUtil[i].ID < byUtil[j].ID
		}
		return ui < uj
	})
	return Summary{
		OperatorName: s.OperatorName, NetworkName: s.NetworkName, Disclaimer: s.Disclaimer,
		AsOf: s.AsOf, Timezone: s.Timezone, From: s.Query.From, To: s.Query.To,
		KPIs:         kpis(s.Network, s.StationCount, s.ConnectorCount),
		SessionStats: sessionsOf(s.Network), ConnectorsByStatus: s.StatusCounts,
		Definitions: definitions(), Assumptions: assumptionLines(),
		IncludedCosts: domain.IncludedCostComponents(), ExcludedCosts: domain.ExcludedCostComponents(),
		TopStations: firstN(byRev, 5), BottomStations: firstN(byUtil, 5),
	}
}

func filled[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func firstN[T any](in []T, n int) []T {
	if len(in) < n {
		n = len(in)
	}
	out := make([]T, n)
	copy(out, in[:n])
	return out
}

func trendsFrom(s *Snapshot) Trends {
	return Trends{
		Daily: dailyPoints(s.Daily), Hourly: hourPoints(s.Network), ByWeekday: weekdayPoints(s.Network),
		Notes: []string{domain.FormulaSessionAttribution, "Hourly energy and revenue are spread across the hours a session was actually charging. Daily totals attribute the whole session to its start date, so a session that crosses midnight stays on the start date."},
	}
}

func detailFrom(s *Snapshot, sm StationMetric) StationDetail {
	conns := make([]ConnectorRow, 0, len(sm.Connectors))
	for _, c := range sm.Connectors {
		conns = append(conns, connectorRow(c))
	}
	faults := make([]FaultDTO, 0, len(sm.Faults))
	for _, f := range sm.Faults {
		faults = append(faults, faultDTO(f, s.AsOf))
	}
	sort.Slice(faults, func(i, j int) bool { return faults[i].Start.After(faults[j].Start) })
	return StationDetail{
		Station: stationRow(sm), KPIs: kpis(sm.Totals, 1, len(sm.Connectors)),
		SessionStats: sessionsOf(sm.Totals), Connectors: conns,
		Daily: dailyPoints(sm.Daily), Hourly: hourPoints(sm.Totals), Faults: faults,
		Definitions: definitions(), Assumptions: assumptionLines(),
		Tariffs: TariffView{
			PeakTariffINRPerKWh: sm.Station.PeakTariffINR, OffPeakTariffINRPerKWh: sm.Station.OffPeakTariffINR,
			PeakEnergyCostINRPerKWh: sm.Station.PeakEnergyCostINR, OffPeakEnergyCostINRPerKWh: sm.Station.OffPeakEnergyCostINR,
			PeakHours: "08:00–11:00 and 18:00–21:00 IST",
		},
	}
}

func namedFromTotals(key, label string, t Totals, connectors int) NamedUtil {
	util, ok := t.Utilization()
	idle := t.Available - t.Charging
	if idle < 0 {
		idle = 0
	}
	return NamedUtil{
		Key: key, Label: label, UtilizationPct: r1ptr(util, ok),
		ChargingHours: hrs(t.Charging), AvailableHours: hrs(t.Available), IdleHours: hrs(idle),
		EnergyKWh: r2(t.Energy), ConnectorCount: connectors, SessionCount: t.Sessions,
	}
}

func utilizationFrom(s *Snapshot) UtilizationReport {
	days := s.Query.To.Sub(s.Query.From).Hours() / 24
	if days <= 0 {
		days = 1
	}
	perDay := 0.0
	if s.ConnectorCount > 0 {
		perDay = s.Network.Energy / float64(s.ConnectorCount) / days
	}
	byType := map[string]*Totals{}
	typeN := map[string]int{}
	byPower := map[string]*Totals{}
	powerN := map[string]int{}
	powerLabel := map[string]string{}
	var low []LowUtilAsset
	for _, cm := range s.Connectors {
		key := string(cm.Connector.ChargerType)
		if byType[key] == nil {
			byType[key] = &Totals{}
		}
		accumulate(byType[key], cm.Totals)
		typeN[key]++
		pk := fmt.Sprintf("%s|%g", cm.Connector.ChargerType, cm.Connector.RatedPowerKW)
		if byPower[pk] == nil {
			byPower[pk] = &Totals{}
			powerLabel[pk] = fmt.Sprintf("%s %s kW", cm.Connector.ChargerType, trimKW(cm.Connector.RatedPowerKW))
		}
		accumulate(byPower[pk], cm.Totals)
		powerN[pk]++
		util, ok := cm.Utilization()
		if ok && util < 10 && cm.Available >= 24*time.Hour {
			u := util
			low = append(low, LowUtilAsset{
				StationID: cm.Connector.StationID, StationName: cm.StationName, ConnectorID: cm.Connector.ID,
				ChargerType: string(cm.Connector.ChargerType), RatedPowerKW: cm.Connector.RatedPowerKW,
				UtilizationPct: r1ptr(u, true), EnergyKWh: r2(cm.Energy), AvailableHours: hrs(cm.Available),
			})
		}
	}
	sort.Slice(low, func(i, j int) bool { return *low[i].UtilizationPct < *low[j].UtilizationPct })
	util, uok := s.Network.Utilization()
	idle := s.Network.Available - s.Network.Charging
	if idle < 0 {
		idle = 0
	}
	peakUtil, pok := UtilizationPct(s.Network.PeakCharging, s.Network.PeakAvailable)
	offUtil, ook := UtilizationPct(s.Network.OffCharging, s.Network.OffAvailable)
	report := UtilizationReport{
		Formula: domain.FormulaUtilization,
		Notes: []string{
			domain.FormulaUtilization,
			"Peak hours are 08:00–11:00 and 18:00–21:00 IST. Off-peak is every other hour.",
			"A connector is listed as underutilized when utilization is below 10% and it had at least 24 hours of available operating time.",
		},
		UtilizationPct: r1ptr(util, uok),
		ChargingHours:  hrs(s.Network.Charging), AvailableHours: hrs(s.Network.Available), IdleHours: hrs(idle),
		ExcludedMaintenanceHours: hrs(s.Network.Maintenance), ExcludedDowntimeHours: hrs(s.Network.Downtime),
		EnergyPerConnectorPerDay: r2(perDay), UnderutilizedAssets: len(low),
		Peak:    NamedUtil{Key: "peak", Label: "Peak hours", UtilizationPct: r1ptr(peakUtil, pok), ChargingHours: hrs(s.Network.PeakCharging), AvailableHours: hrs(s.Network.PeakAvailable)},
		OffPeak: NamedUtil{Key: "off_peak", Label: "Off-peak hours", UtilizationPct: r1ptr(offUtil, ook), ChargingHours: hrs(s.Network.OffCharging), AvailableHours: hrs(s.Network.OffAvailable)},
		Daily:   dailyPoints(s.Daily), Hourly: hourPoints(s.Network), LowUtilization: filled(low),
		Definitions: definitions(),
	}
	for _, sm := range s.Stations {
		report.ByStation = append(report.ByStation, namedFromTotals(sm.Station.ID, sm.Station.Name, sm.Totals, len(sm.Connectors)))
	}
	sort.Slice(report.ByStation, func(i, j int) bool {
		return ptrOrNeg(report.ByStation[i].UtilizationPct) > ptrOrNeg(report.ByStation[j].UtilizationPct)
	})
	for _, key := range []string{"AC", "DC"} {
		if byType[key] == nil {
			continue
		}
		report.ByChargerType = append(report.ByChargerType, namedFromTotals(key, key, *byType[key], typeN[key]))
	}
	keys := make([]string, 0, len(byPower))
	for k := range byPower {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		report.ByPowerRating = append(report.ByPowerRating, namedFromTotals(k, powerLabel[k], *byPower[k], powerN[k]))
	}
	return report
}

func accumulate(dst *Totals, src Totals) {
	dst.Charging += src.Charging
	dst.Available += src.Available
	dst.Energy += src.Energy
	dst.Sessions += src.Sessions
	dst.Revenue += src.Revenue
	dst.Cost += src.Cost
	dst.Maintenance += src.Maintenance
	dst.Downtime += src.Downtime
}

func ptrOrNeg(p *float64) float64 {
	if p == nil {
		return -1
	}
	return *p
}

func trimKW(v float64) string {
	if v == float64(int(v)) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.1f", v)
}

func revenueFrom(s *Snapshot) RevenueReport {
	var rows []StationFinance
	for _, sm := range s.Stations {
		rows = append(rows, financeRow(sm))
	}
	byRev := append([]StationFinance(nil), rows...)
	sort.Slice(byRev, func(i, j int) bool { return byRev[i].RevenueINR > byRev[j].RevenueINR })
	byMargin := make([]StationFinance, 0, len(rows))
	for _, row := range rows {
		if row.EnergyKWh > 0 && row.GrossMarginPct != nil {
			byMargin = append(byMargin, row)
		}
	}
	sort.Slice(byMargin, func(i, j int) bool { return *byMargin[i].GrossMarginPct < *byMargin[j].GrossMarginPct })
	margin, pct, ok := s.Network.Margin()
	revKWh, revOK := s.Network.RevenuePerKWh()
	costKWh, costOK := s.Network.CostPerKWh()
	var marginPer *float64
	if revOK && costOK {
		marginPer = r2ptr(revKWh-costKWh, true)
	}
	perConn := 0.0
	if s.ConnectorCount > 0 {
		perConn = s.Network.Revenue / float64(s.ConnectorCount)
	}
	return RevenueReport{
		Assumptions:            assumptionLines(),
		IncludedCostComponents: domain.IncludedCostComponents(),
		ExcludedCostComponents: domain.ExcludedCostComponents(),
		RevenueINR:             r2(s.Network.Revenue), ElectricityCostINR: r2(s.Network.Cost),
		GrossMarginINR: r2(margin), GrossMarginPct: r1ptr(pct, ok),
		AvgRevenuePerSession: r2ptr(safeDiv(s.Network.Revenue, float64(s.Network.Sessions))),
		RevenuePerConnector:  r2(perConn),
		RevenuePerKWh:        r2ptr(revKWh, revOK), EnergyCostPerKWh: r2ptr(costKWh, costOK),
		MarginPerKWh: marginPer, ByStation: filled(byRev), Daily: dailyPoints(s.Daily),
		TopStations: firstN(byRev, 5), BottomStations: firstN(byMargin, 5),
		Definitions: definitions(),
	}
}

func demandFrom(sm StationMetric, days float64) DemandSite {
	per := 0.0
	n := len(sm.Connectors)
	if n > 0 && days > 0 {
		per = sm.Energy / float64(n) / days
	}
	hour, util, ok := busiestHour(sm.Totals)
	note := "Observed charging demand only. This does not measure grid import or transformer loading."
	if ok && util >= 65 {
		note = "Charging bays are busy in this hour. That can mean queues. It is not evidence of a grid or transformer limit."
	}
	u, uok := sm.Utilization()
	return DemandSite{
		StationID: sm.Station.ID, StationName: sm.Station.Name, Area: sm.Station.Area,
		EnergyKWh: r2(sm.Energy), EnergyPerConnectorPerDay: r2(per),
		UtilizationPct: r1ptr(u, uok), PeakHour: hour, PeakHourLabel: fmt.Sprintf("%02d:00", hour),
		PeakHourUtilizationPct: r1ptr(util, ok), Note: note,
	}
}

func busiestHour(t Totals) (hour int, util float64, ok bool) {
	best := -1.0
	hour = 0
	for h := 0; h < 24; h++ {
		if t.HourAvailable[h] < 10*time.Hour {
			continue
		}
		u, good := UtilizationPct(t.HourCharging[h], t.HourAvailable[h])
		if good && u > best {
			best = u
			hour = h
			ok = true
			util = u
		}
	}
	return hour, util, ok
}

func energyFrom(s *Snapshot) EnergyReport {
	days := s.Query.To.Sub(s.Query.From).Hours() / 24
	if days <= 0 {
		days = 1
	}
	var sites []DemandSite
	for _, sm := range s.Stations {
		sites = append(sites, demandFrom(sm, days))
	}
	byEnergy := append([]DemandSite(nil), sites...)
	sort.Slice(byEnergy, func(i, j int) bool { return byEnergy[i].EnergyKWh > byEnergy[j].EnergyKWh })
	byLow := append([]DemandSite(nil), sites...)
	sort.Slice(byLow, func(i, j int) bool {
		return byLow[i].EnergyPerConnectorPerDay < byLow[j].EnergyPerConnectorPerDay
	})
	var conns []ConnectorEnergy
	for _, cm := range s.Connectors {
		conns = append(conns, ConnectorEnergy{
			ConnectorID: cm.Connector.ID, StationID: cm.Connector.StationID, StationName: cm.StationName,
			ChargerType: string(cm.Connector.ChargerType), RatedPowerKW: cm.Connector.RatedPowerKW,
			EnergyKWh: r2(cm.Energy), SessionCount: cm.Sessions,
		})
	}
	sort.Slice(conns, func(i, j int) bool { return conns[i].EnergyKWh > conns[j].EnergyKWh })
	if len(conns) > 20 {
		conns = conns[:20]
	}
	avgE, eok := s.Network.AvgEnergy()
	avgP, pok := s.Network.AvgPower()
	var peak *float64
	if s.Network.PeakObservedKW > 0 {
		peak = r2ptr(s.Network.PeakObservedKW, true)
	}
	utilReport := utilizationFrom(s)
	return EnergyReport{
		Notes:          []string{domain.FormulaPower, domain.FormulaSessionAttribution, "High-demand and low-demand lists compare energy delivered. They are not a grid study."},
		TotalEnergyKWh: r2(s.Network.Energy), AvgSessionEnergyKWh: r2ptr(avgE, eok),
		AvgChargingPowerKW: r2ptr(avgP, pok), PeakObservedPowerKW: peak,
		ByStation: filled(byEnergy), ByConnector: filled(conns), ByChargerType: utilReport.ByChargerType,
		Daily: dailyPoints(s.Daily), Hourly: hourPoints(s.Network),
		RatedVsObserved: filled(powerCompare(s.Connectors)),
		HighDemand:      firstN(byEnergy, 5), LowDemand: firstN(byLow, 5),
		Definitions: definitions(),
	}
}

func powerCompare(conns []ConnectorMetric) []PowerCompare {
	type g struct {
		t     Totals
		n     int
		rated float64
		kind  string
		peak  float64
	}
	groups := map[string]*g{}
	for _, cm := range conns {
		key := fmt.Sprintf("%s|%g", cm.Connector.ChargerType, cm.Connector.RatedPowerKW)
		if groups[key] == nil {
			groups[key] = &g{rated: cm.Connector.RatedPowerKW, kind: string(cm.Connector.ChargerType)}
		}
		groups[key].t.Energy += cm.Energy
		groups[key].t.ChargeSec += cm.ChargeSec
		groups[key].n++
		if cm.PeakObservedKW > groups[key].peak {
			groups[key].peak = cm.PeakObservedKW
		}
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []PowerCompare
	for _, k := range keys {
		gr := groups[k]
		avg, ok := gr.t.AvgPower()
		var ratio *float64
		if ok && gr.rated > 0 {
			ratio = r2ptr(avg/gr.rated, true)
		}
		var peak *float64
		if gr.peak > 0 {
			peak = r2ptr(gr.peak, true)
		}
		out = append(out, PowerCompare{
			ChargerType: gr.kind, RatedPowerKW: gr.rated,
			Label:         fmt.Sprintf("%s %s kW", gr.kind, trimKW(gr.rated)),
			AvgObservedKW: r2ptr(avg, ok), PeakObservedKW: peak, ObservedToRated: ratio,
			EnergyKWh: r2(gr.t.Energy), ConnectorCount: gr.n,
		})
	}
	return out
}

func reliabilityFrom(s *Snapshot) ReliabilityReport {
	up, ok := s.Network.Uptime()
	mttr, mttrOK := MTTR([]time.Duration{s.Network.RepairTime})
	if s.Network.RepairCount > 0 {
		mttr = s.Network.RepairTime / time.Duration(s.Network.RepairCount)
		mttrOK = true
	} else {
		mttrOK = false
	}
	mtbf, mtbfOK := MTBF(s.Network.Up, s.Network.Failures)
	var rows []StationReliability
	for _, sm := range s.Stations {
		sup, sok := sm.Uptime()
		var smttr *float64
		if sm.RepairCount > 0 {
			smttr = r2ptr((sm.RepairTime / time.Duration(sm.RepairCount)).Hours(), true)
		}
		smbf, bok := MTBF(sm.Up, sm.Failures)
		source := "station"
		if sm.UsedNetworkRate {
			source = "network"
		}
		if sm.Downtime <= 0 {
			source = "not_applicable"
		}
		rows = append(rows, StationReliability{
			StationID: sm.Station.ID, StationName: sm.Station.Name, Area: sm.Station.Area,
			UptimePct: r1ptr(sup, sok), DowntimeHours: hrs(sm.Downtime),
			FaultCount: sm.Totals.Faults, OpenFaultCount: sm.OpenFaults,
			MTTRHours: smttr, MTBFHours: r2ptr(smbf.Hours(), bok),
			RevenueAtRiskINR: r2(sm.RevenueAtRisk), RateSource: source,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].DowntimeHours > rows[j].DowntimeHours })
	var cats []CategoryCount
	for cat, n := range s.Network.FaultsByCategory {
		cats = append(cats, CategoryCount{Category: cat, Count: n})
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].Count > cats[j].Count })
	var faults []FaultDTO
	for _, f := range s.Faults {
		faults = append(faults, faultDTO(f, s.AsOf))
	}
	sort.Slice(faults, func(i, j int) bool { return faults[i].Start.After(faults[j].Start) })
	var longest *Outage
	if s.LongestOpen != nil {
		f := s.LongestOpen
		longest = &Outage{
			FaultID: f.ID, StationID: f.StationID, ConnectorID: f.ConnectorID,
			Category: string(f.Category), Start: f.Start,
			DurationHours: r2(s.AsOf.Sub(f.Start).Hours()), Description: f.Description,
		}
	}
	var mttrPtr *float64
	if mttrOK {
		mttrPtr = r2ptr(mttr.Hours(), true)
	}
	return ReliabilityReport{
		Formula:   domain.FormulaUptime,
		Notes:     []string{domain.FormulaUptime, domain.FormulaMTTR, domain.FormulaMTBF, domain.FormulaRevenueAtRisk},
		UptimePct: r1ptr(up, ok), DowntimeHours: hrs(s.Network.Downtime),
		MaintenanceHoursExcluded: hrs(s.Network.Maintenance),
		FaultCount:               s.Network.Faults, OpenFaultCount: s.Network.OpenFaults,
		MTTRHours: mttrPtr, MTBFHours: r2ptr(mtbf.Hours(), mtbfOK),
		LongestOngoing: longest, RevenueAtRiskINR: r2(s.Network.RevenueAtRisk),
		RevenueAtRiskAssumption: domain.FormulaRevenueAtRisk,
		ByStation:               filled(rows), ByCategory: filled(cats), Faults: filled(faults), Definitions: definitions(),
	}
}

func sessionReportFrom(s *Snapshot) SessionReport {
	reasons := map[string]int{}
	for _, ses := range s.Sessions {
		if ses.Status == domain.SessionFailed || ses.Status == domain.SessionInterrupted {
			reason := ses.TerminationReason
			if reason == "" {
				reason = string(ses.Status)
			}
			reasons[reason]++
		}
	}
	var reasonRows []ReasonCount
	for reason, n := range reasons {
		reasonRows = append(reasonRows, ReasonCount{Reason: reason, Count: n})
	}
	sort.Slice(reasonRows, func(i, j int) bool { return reasonRows[i].Count > reasonRows[j].Count })
	avgD, dok := s.Network.AvgDurationMin()
	avgE, eok := s.Network.AvgEnergy()
	return SessionReport{
		Notes:        []string{domain.FormulaSessionAttribution, "Success rate is completed sessions divided by all sessions that started in the window, including failed and interrupted attempts."},
		SessionCount: s.Network.Sessions, Completed: s.Network.Completed, Failed: s.Network.Failed,
		Interrupted: s.Network.Interrupted, InProgress: s.Network.InProgress,
		SuccessRatePct:     ratePtr(s.Network.Completed, s.Network.Sessions),
		FailureRatePct:     ratePtr(s.Network.Failed, s.Network.Sessions),
		InterruptedRatePct: ratePtr(s.Network.Interrupted, s.Network.Sessions),
		AvgDurationMin:     r1ptr(avgD, dok), AvgEnergyKWh: r2ptr(avgE, eok),
		ByHour: hourPoints(s.Network), ByWeekday: weekdayPoints(s.Network),
		PeakWindows: filled(peakWindows(s.Network.HourSessionStarts)),
		Daily:       dailyPoints(s.Daily), FailedByReason: filled(reasonRows), Definitions: definitions(),
	}
}

func peakWindows(starts [24]int) []PeakWindow {
	total := 0
	for _, n := range starts {
		total += n
	}
	type cand struct{ h, n int }
	var cands []cand
	for h := 0; h <= 21; h++ {
		n := starts[h] + starts[h+1] + starts[h+2]
		cands = append(cands, cand{h, n})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].n > cands[j].n })
	var out []PeakWindow
	for _, c := range cands {
		if len(out) == 3 || c.n == 0 {
			break
		}
		overlap := false
		for _, existing := range out {
			if c.h < existing.EndHour && existing.StartHour < c.h+3 {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		share := 0.0
		if total > 0 {
			share = float64(c.n) / float64(total) * 100
		}
		out = append(out, PeakWindow{
			StartHour: c.h, EndHour: c.h + 3,
			Label:    fmt.Sprintf("%02d:00–%02d:00 IST", c.h, c.h+3),
			Sessions: c.n, SharePct: domain.Round1(share),
		})
	}
	return out
}

func sessionDTO(ses domain.Session) SessionDTO {
	chargeMin := float64(ses.ChargeDurationSec) / 60
	avg := ses.AvgPowerKW
	peak := ses.PeakPowerKW
	if ses.EnergyKWh > 0 {
		sec := ses.ChargeDurationSec
		if sec <= 0 {
			sec = ses.DurationSec
		}
		if a, p, ok := ObservedPower(ses.EnergyKWh, sec, ses.PeakPowerKW); ok {
			if avg <= 0 {
				avg = domain.Round2(a)
			}
			if peak <= 0 {
				peak = domain.Round2(p)
			}
		}
	}
	return SessionDTO{
		ID: ses.ID, StationID: ses.StationID, ConnectorID: ses.ConnectorID,
		Start: ses.Start, End: ses.End,
		DurationMin: r2(float64(ses.DurationSec) / 60), ChargingMin: r2(chargeMin),
		EnergyKWh: ses.EnergyKWh, AvgPowerKW: avg, PeakPowerKW: peak,
		TariffINRPerKWh: ses.TariffINRPerKWh, RevenueINR: ses.RevenueINR,
		ElectricityCostINR: ses.EnergyCostINR, GrossMarginINR: r2(ses.RevenueINR - ses.EnergyCostINR),
		Status: string(ses.Status), TerminationReason: ses.TerminationReason,
		ChargerType: string(ses.ChargerType), RatedPowerKW: ses.RatedPowerKW,
	}
}

func (s *Snapshot) sessionDTOs(status string) []SessionDTO {
	out := make([]SessionDTO, 0, len(s.Sessions))
	for _, ses := range s.Sessions {
		if status != "" && string(ses.Status) != status {
			continue
		}
		out = append(out, sessionDTO(ses))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].ID > out[j].ID
		}
		return out[i].Start.After(out[j].Start)
	})
	return out
}
