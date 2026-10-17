// Package domain holds the charging-network types shared by the generator,
// in-memory store, analytics, and HTTP API. Money is INR. Energy is kWh.
// Power is kW. Times are absolute instants; business hours use Asia/Kolkata.
package domain

import "time"

const (
	CurrencyINR = "INR"
	Timezone    = "Asia/Kolkata"
)

type ChargerType string

const (
	ChargerAC ChargerType = "AC"
	ChargerDC ChargerType = "DC"
)

type ConnectorStatus string

const (
	StatusAvailable   ConnectorStatus = "available"
	StatusOccupied    ConnectorStatus = "occupied"
	StatusUnavailable ConnectorStatus = "unavailable"
	StatusFaulted     ConnectorStatus = "faulted"
)

type SessionStatus string

const (
	SessionCompleted   SessionStatus = "completed"
	SessionFailed      SessionStatus = "failed"
	SessionInterrupted SessionStatus = "interrupted"
	SessionInProgress  SessionStatus = "in_progress"
)

type FaultCategory string

const (
	FaultCommunication FaultCategory = "communication_failure"
	FaultConnector     FaultCategory = "connector_failure"
	FaultPowerSupply   FaultCategory = "power_supply_failure"
	FaultTemperature   FaultCategory = "temperature_alarm"
	FaultPayment       FaultCategory = "payment_error"
	FaultEmergency     FaultCategory = "emergency_shutdown"
)

type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityWarning  AlertSeverity = "warning"
	SeverityInfo     AlertSeverity = "info"
)

type AlertStatus string

const (
	AlertOpen         AlertStatus = "open"
	AlertAcknowledged AlertStatus = "acknowledged"
	AlertResolved     AlertStatus = "resolved"
)

// Dataset is the full synthetic network produced at process start.
// Alerts are not part of the dataset; the store owns them because
// acknowledge and resolve mutate state for the life of the process.
type Dataset struct {
	OperatorName string    `json:"operator_name"`
	NetworkName  string    `json:"network_name"`
	Disclaimer   string    `json:"disclaimer"`
	Seed         int64     `json:"seed"`
	Days         int       `json:"days"`
	Timezone     string    `json:"timezone"`
	AsOf         time.Time `json:"as_of"`
	DataStart    time.Time `json:"data_start"`
	Stations     []Station
	Connectors   []Connector
	Sessions     []Session
	Faults       []Fault
	Maintenance  []MaintenanceWindow
	Telemetry    []TelemetrySample
}

type Station struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Area                 string    `json:"area"`
	Address              string    `json:"address"`
	City                 string    `json:"city"`
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	CommissionedAt       time.Time `json:"commissioned_at"`
	PeakTariffINR        float64   `json:"peak_tariff_inr_per_kwh"`
	OffPeakTariffINR     float64   `json:"off_peak_tariff_inr_per_kwh"`
	PeakEnergyCostINR    float64   `json:"peak_energy_cost_inr_per_kwh"`
	OffPeakEnergyCostINR float64   `json:"off_peak_energy_cost_inr_per_kwh"`
}

type Connector struct {
	ID           string          `json:"id"`
	StationID    string          `json:"station_id"`
	Label        string          `json:"label"`
	ChargerType  ChargerType     `json:"charger_type"`
	Standard     string          `json:"standard"`
	RatedPowerKW float64         `json:"rated_power_kw"`
	Status       ConnectorStatus `json:"status"`
}

type Session struct {
	ID                string        `json:"id"`
	StationID         string        `json:"station_id"`
	ConnectorID       string        `json:"connector_id"`
	Start             time.Time     `json:"start"`
	End               time.Time     `json:"end"`
	DurationSec       int64         `json:"duration_sec"`
	ChargeDurationSec int64         `json:"charge_duration_sec"`
	EnergyKWh         float64       `json:"energy_kwh"`
	AvgPowerKW        float64       `json:"avg_power_kw"`
	PeakPowerKW       float64       `json:"peak_power_kw"`
	TariffINRPerKWh   float64       `json:"tariff_inr_per_kwh"`
	RevenueINR        float64       `json:"revenue_inr"`
	EnergyCostINR     float64       `json:"energy_cost_inr"`
	Status            SessionStatus `json:"status"`
	TerminationReason string        `json:"termination_reason"`
	ChargerType       ChargerType   `json:"charger_type"`
	RatedPowerKW      float64       `json:"rated_power_kw"`
}

type Fault struct {
	ID          string        `json:"id"`
	StationID   string        `json:"station_id"`
	ConnectorID string        `json:"connector_id"`
	Category    FaultCategory `json:"category"`
	Severity    string        `json:"severity"`
	Start       time.Time     `json:"start"`
	End         *time.Time    `json:"end"`
	Description string        `json:"description"`
}

type MaintenanceWindow struct {
	ID          string    `json:"id"`
	StationID   string    `json:"station_id"`
	ConnectorID string    `json:"connector_id"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Reason      string    `json:"reason"`
}

// TelemetrySample is an hourly average derived from a charging session.
// Peak power lives on the session; samples do not invent grid-meter data.
type TelemetrySample struct {
	ID          string    `json:"id"`
	StationID   string    `json:"station_id"`
	ConnectorID string    `json:"connector_id"`
	SessionID   string    `json:"session_id"`
	Timestamp   time.Time `json:"timestamp"`
	PowerKW     float64   `json:"power_kw"`
	EnergyKWh   float64   `json:"energy_kwh"`
}

type Alert struct {
	ID                string     `json:"id"`
	Severity          string     `json:"severity"`
	Category          string     `json:"category"`
	StationID         string     `json:"station_id"`
	StationName       string     `json:"station_name"`
	ConnectorID       string     `json:"connector_id"`
	Timestamp         time.Time  `json:"timestamp"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Metric            string     `json:"metric"`
	ObservedValue     string     `json:"observed_value"`
	RecommendedAction string     `json:"recommended_action"`
	Status            string     `json:"status"`
	UpdatedAt         *time.Time `json:"updated_at"`
}

// Query is the shared filter for analytics endpoints.
// From is inclusive and To is exclusive.
type Query struct {
	From        time.Time
	To          time.Time
	StationID   string
	ChargerType string
}

// FaultEnd returns the recorded end, or asOf when the fault is still open.
func (f Fault) FaultEnd(asOf time.Time) time.Time {
	if f.End == nil || f.End.After(asOf) {
		return asOf
	}
	return *f.End
}

// Ongoing reports whether the fault was still open at asOf.
func (f Fault) Ongoing(asOf time.Time) bool {
	return f.End == nil || f.End.After(asOf)
}

// IsPeakHour is true for 08:00–11:00 and 18:00–21:00 in the given location.
func IsPeakHour(t time.Time, loc *time.Location) bool {
	h := t.In(loc).Hour()
	return (h >= 8 && h < 11) || (h >= 18 && h < 21)
}

// HardwareFault reports categories that take a connector to "faulted"
// rather than "unavailable".
func HardwareFault(c FaultCategory) bool {
	switch c {
	case FaultConnector, FaultPowerSupply, FaultTemperature, FaultEmergency:
		return true
	default:
		return false
	}
}
