export type Envelope<T> = {
  demo: boolean;
  data: T;
  meta: {
    from?: string;
    to?: string;
    as_of: string;
    timezone: string;
    seed: number;
    disclaimer: string;
    page?: number;
    limit?: number;
    total?: number;
    total_pages?: number;
  };
};

export type StationIndex = { id: string; name: string; area: string };

export type Health = {
  status: string;
  operator_name: string;
  network_name: string;
  as_of: string;
  data_start: string;
  timezone: string;
  seed: number;
  days: number;
  station_count: number;
  connector_count: number;
  session_count: number;
  fault_count: number;
  connectors_by_status: Record<string, number>;
  disclaimer: string;
  stations: StationIndex[];
};

export type KPIs = {
  station_count: number;
  connector_count: number;
  uptime_pct: number | null;
  utilization_pct: number | null;
  session_count: number;
  energy_kwh: number;
  revenue_inr: number;
  electricity_cost_inr: number;
  gross_margin_inr: number;
  gross_margin_pct: number | null;
  avg_session_duration_min: number | null;
  avg_energy_per_session_kwh: number | null;
  charging_hours: number;
  available_hours: number;
  downtime_hours: number;
};

export type SessionStats = {
  completed: number;
  failed: number;
  interrupted: number;
  in_progress: number;
  success_rate_pct: number | null;
  failure_rate_pct: number | null;
  interrupted_rate_pct: number | null;
};

export type StationRow = {
  id: string;
  name: string;
  area: string;
  address: string;
  city: string;
  charger_types: string[];
  connector_count: number;
  rated_power_kw: number;
  max_power_kw: number;
  status: string;
  connectors_by_status: Record<string, number>;
  utilization_pct: number | null;
  uptime_pct: number | null;
  energy_kwh: number;
  revenue_inr: number;
  electricity_cost_inr: number;
  gross_margin_inr: number;
  gross_margin_pct: number | null;
  session_count: number;
  downtime_hours: number;
};

export type DailyPoint = {
  date: string;
  energy_kwh: number;
  revenue_inr: number;
  electricity_cost_inr: number;
  gross_margin_inr: number;
  sessions: number;
  utilization_pct: number | null;
  uptime_pct: number | null;
  downtime_hours: number;
};

export type HourPoint = {
  hour: number;
  label: string;
  utilization_pct: number | null;
  session_starts: number;
  energy_kwh: number;
  revenue_inr: number;
};

export type WeekdayPoint = {
  weekday: number;
  label: string;
  sessions: number;
  energy_kwh: number;
};

export type Insight = {
  id: string;
  category: string;
  priority: string;
  title: string;
  evidence: string;
  why_it_matters: string;
  recommended_action: string;
  estimated_impact: string;
  station_ids: string[];
  station_names: string[];
  metric: string;
  observed_value: string;
};

export type Alert = {
  id: string;
  severity: string;
  category: string;
  station_id: string;
  station_name: string;
  connector_id: string;
  timestamp: string;
  title: string;
  description: string;
  metric: string;
  observed_value: string;
  recommended_action: string;
  status: string;
  updated_at: string | null;
};

export type Summary = {
  operator_name: string;
  network_name: string;
  disclaimer: string;
  as_of: string;
  timezone: string;
  kpis: KPIs;
  session_stats: SessionStats;
  connectors_by_status: Record<string, number>;
  definitions: Record<string, string>;
  assumptions: string[];
  included_cost_components: string[];
  excluded_cost_components: string[];
  top_stations: StationRow[];
  bottom_stations: StationRow[];
};

export type Dashboard = {
  summary: Summary;
  critical_alerts: Alert[];
  insights: Insight[];
};

export type Trends = {
  daily: DailyPoint[];
  hourly: HourPoint[];
  by_weekday: WeekdayPoint[];
  notes: string[];
};

export type ConnectorRow = {
  id: string;
  label: string;
  charger_type: string;
  standard: string;
  rated_power_kw: number;
  status: string;
  utilization_pct: number | null;
  uptime_pct: number | null;
  energy_kwh: number;
  revenue_inr: number;
  gross_margin_inr: number;
  gross_margin_pct: number | null;
  session_count: number;
  downtime_hours: number;
};

export type Fault = {
  id: string;
  station_id: string;
  connector_id: string;
  category: string;
  severity: string;
  start: string;
  end: string | null;
  ongoing: boolean;
  duration_hours: number;
  description: string;
};

export type StationDetail = {
  station: StationRow;
  kpis: KPIs;
  session_stats: SessionStats;
  connectors: ConnectorRow[];
  daily: DailyPoint[];
  hourly: HourPoint[];
  faults: Fault[];
  definitions: Record<string, string>;
  assumptions: string[];
  tariffs: {
    peak_tariff_inr_per_kwh: number;
    off_peak_tariff_inr_per_kwh: number;
    peak_energy_cost_inr_per_kwh: number;
    off_peak_energy_cost_inr_per_kwh: number;
    peak_hours: string;
  };
  insights: Insight[];
};

export type Session = {
  id: string;
  station_id: string;
  connector_id: string;
  start: string;
  end: string;
  duration_min: number;
  charging_min: number;
  energy_kwh: number;
  avg_power_kw: number;
  peak_power_kw: number;
  tariff_inr_per_kwh: number;
  revenue_inr: number;
  electricity_cost_inr: number;
  gross_margin_inr: number;
  status: string;
  termination_reason: string;
  charger_type: string;
  rated_power_kw: number;
};

export type NamedUtil = {
  key: string;
  label: string;
  utilization_pct: number | null;
  charging_hours: number;
  available_hours: number;
  idle_hours: number;
  energy_kwh: number;
  connector_count: number;
  session_count: number;
};

export type UtilizationReport = {
  formula: string;
  notes: string[];
  utilization_pct: number | null;
  charging_hours: number;
  available_hours: number;
  idle_hours: number;
  excluded_maintenance_hours: number;
  excluded_downtime_hours: number;
  energy_per_connector_per_day: number;
  underutilized_assets: number;
  by_station: NamedUtil[];
  by_charger_type: NamedUtil[];
  by_power_rating: NamedUtil[];
  peak: NamedUtil;
  off_peak: NamedUtil;
  daily: DailyPoint[];
  hourly: HourPoint[];
  low_utilization: {
    station_id: string;
    station_name: string;
    connector_id: string;
    charger_type: string;
    rated_power_kw: number;
    utilization_pct: number | null;
    energy_kwh: number;
    available_hours: number;
  }[];
  definitions: Record<string, string>;
};

export type StationFinance = {
  station_id: string;
  station_name: string;
  area: string;
  revenue_inr: number;
  electricity_cost_inr: number;
  gross_margin_inr: number;
  gross_margin_pct: number | null;
  energy_kwh: number;
  sessions: number;
  revenue_per_session_inr: number | null;
  revenue_per_connector_inr: number;
  revenue_per_kwh: number | null;
  energy_cost_per_kwh: number | null;
  margin_per_kwh: number | null;
  utilization_pct: number | null;
  connector_count: number;
};

export type RevenueReport = {
  assumptions: string[];
  included_cost_components: string[];
  excluded_cost_components: string[];
  revenue_inr: number;
  electricity_cost_inr: number;
  gross_margin_inr: number;
  gross_margin_pct: number | null;
  avg_revenue_per_session_inr: number | null;
  revenue_per_connector_inr: number;
  revenue_per_kwh: number | null;
  energy_cost_per_kwh: number | null;
  margin_per_kwh: number | null;
  by_station: StationFinance[];
  daily: DailyPoint[];
  top_stations: StationFinance[];
  bottom_stations: StationFinance[];
  definitions: Record<string, string>;
};

export type DemandSite = {
  station_id: string;
  station_name: string;
  area: string;
  energy_kwh: number;
  energy_per_connector_per_day: number;
  utilization_pct: number | null;
  peak_hour_label: string;
  peak_hour_utilization_pct: number | null;
  note: string;
};

export type EnergyReport = {
  notes: string[];
  total_energy_kwh: number;
  avg_session_energy_kwh: number | null;
  avg_charging_power_kw: number | null;
  peak_observed_power_kw: number | null;
  by_station: DemandSite[];
  by_connector: {
    connector_id: string;
    station_id: string;
    station_name: string;
    charger_type: string;
    rated_power_kw: number;
    energy_kwh: number;
    session_count: number;
  }[];
  daily: DailyPoint[];
  hourly: HourPoint[];
  rated_vs_observed: {
    label: string;
    charger_type: string;
    rated_power_kw: number;
    avg_observed_kw: number | null;
    peak_observed_kw: number | null;
    observed_to_rated: number | null;
    energy_kwh: number;
    connector_count: number;
  }[];
  high_demand: DemandSite[];
  low_demand: DemandSite[];
  definitions: Record<string, string>;
};

export type ReliabilityReport = {
  formula: string;
  notes: string[];
  uptime_pct: number | null;
  downtime_hours: number;
  maintenance_hours_excluded: number;
  fault_count: number;
  open_fault_count: number;
  mttr_hours: number | null;
  mtbf_hours: number | null;
  longest_ongoing: {
    fault_id: string;
    station_id: string;
    connector_id: string;
    category: string;
    start: string;
    duration_hours: number;
    description: string;
  } | null;
  revenue_at_risk_inr: number;
  revenue_at_risk_assumption: string;
  by_station: {
    station_id: string;
    station_name: string;
    area: string;
    uptime_pct: number | null;
    downtime_hours: number;
    fault_count: number;
    open_fault_count: number;
    mttr_hours: number | null;
    mtbf_hours: number | null;
    revenue_at_risk_inr: number;
    revenue_at_risk_rate_source: string;
  }[];
  by_category: { category: string; count: number }[];
  faults: Fault[];
  definitions: Record<string, string>;
};

export type SessionReport = {
  notes: string[];
  session_count: number;
  completed: number;
  failed: number;
  interrupted: number;
  in_progress: number;
  success_rate_pct: number | null;
  failure_rate_pct: number | null;
  interrupted_rate_pct: number | null;
  avg_duration_min: number | null;
  avg_energy_kwh: number | null;
  by_hour: HourPoint[];
  by_weekday: WeekdayPoint[];
  peak_windows: { label: string; sessions: number; share_pct: number }[];
  daily: DailyPoint[];
  failed_by_reason: { reason: string; count: number }[];
  definitions: Record<string, string>;
};
