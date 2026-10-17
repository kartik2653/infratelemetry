"use client";

import Link from "next/link";
import { BarSeries, TrendChart, dailyLabels } from "@/components/charts";
import { RangeFilter, StationSelect } from "@/components/filters";
import { ErrorState, Loading } from "@/components/states";
import { Card, Definitions, Kpi, SectionTitle } from "@/components/ui";
import { qs } from "@/lib/api";
import { useFilters } from "@/lib/filters";
import { hours, inr, kwh, minutes, num, pct, titleCase } from "@/lib/format";
import { useGet } from "@/lib/use-get";
import type { EnergyReport, ReliabilityReport, RevenueReport, SessionReport, UtilizationReport } from "@/lib/types";

const tabs = [
  { id: "utilization", label: "Utilization" },
  { id: "revenue", label: "Revenue" },
  { id: "energy", label: "Energy" },
  { id: "sessions", label: "Sessions" },
  { id: "reliability", label: "Reliability" },
] as const;

type Tab = (typeof tabs)[number]["id"];

export default function AnalyticsPage() {
  const { range, chargerType, stationId, from, to, update, href, clock, sp } = useFilters();
  const requested = sp.get("tab");
  const tab: Tab = tabs.some((item) => item.id === requested) ? (requested as Tab) : "utilization";
  const query = qs({ from, to, charger_type: chargerType, station_id: stationId });
  const util = useGet<UtilizationReport>(tab === "utilization" ? `/api/v1/analytics/utilization${query}` : null);
  const revenue = useGet<RevenueReport>(tab === "revenue" ? `/api/v1/analytics/revenue${query}` : null);
  const energy = useGet<EnergyReport>(tab === "energy" ? `/api/v1/analytics/energy${query}` : null);
  const sessions = useGet<SessionReport>(tab === "sessions" ? `/api/v1/analytics/sessions${query}` : null);
  const reliability = useGet<ReliabilityReport>(tab === "reliability" ? `/api/v1/analytics/reliability${query}` : null);
  const active = tab === "utilization" ? util : tab === "revenue" ? revenue : tab === "energy" ? energy : tab === "sessions" ? sessions : reliability;

  return (
    <div className="space-y-4">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">Analytics</h1>
          <p className="mt-1 max-w-3xl text-sm text-ink-soft">Filters apply to every chart on the active tab. Power charts describe charging sessions, not grid import or transformer loading.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <RangeFilter range={range} chargerType={chargerType} onRange={(value) => update({ range: value })} onCharger={(value) => update({ charger_type: value || null })} />
          <StationSelect stations={clock.stations} value={stationId} onChange={(id) => update({ station_id: id || null })} />
        </div>
      </header>
      <div className="flex flex-wrap gap-2">
        {tabs.map((item) => (
          <button key={item.id} type="button" aria-pressed={tab === item.id} onClick={() => update({ tab: item.id })} className={`rounded-full px-3 py-1.5 text-sm ${tab === item.id ? "bg-pine text-white" : "border border-line bg-white"}`}>
            {item.label}
          </button>
        ))}
      </div>
      {active.loading && <Loading />}
      {active.error && <ErrorState message={active.error} onRetry={active.reload} />}
      {tab === "utilization" && util.data && <Utilization data={util.data.data} href={href} />}
      {tab === "revenue" && revenue.data && <Revenue data={revenue.data.data} href={href} />}
      {tab === "energy" && energy.data && <Energy data={energy.data.data} href={href} />}
      {tab === "sessions" && sessions.data && <Sessions data={sessions.data.data} />}
      {tab === "reliability" && reliability.data && <Reliability data={reliability.data.data} href={href} />}
    </div>
  );
}

function Utilization({ data, href }: { data: UtilizationReport; href: (path: string) => string }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Kpi label="Utilization" value={pct(data.utilization_pct)} detail={`${hours(data.charging_hours)} / ${hours(data.available_hours)}`} />
        <Kpi label="Idle available time" value={hours(data.idle_hours)} detail="Available but not charging" />
        <Kpi label="Energy / connector / day" value={`${num(data.energy_per_connector_per_day, 1)} kWh`} />
        <Kpi label="Under 10% utilization" value={num(data.underutilized_assets)} detail="Connectors with 24h+ available" />
      </div>
      <div className="grid gap-4 xl:grid-cols-2">
        <Card>
          <SectionTitle title="Daily utilization" detail={data.formula} />
          <TrendChart data={dailyLabels(data.daily)} lines={[{ key: "utilization_pct", name: "Utilization", color: "#1c6b45" }]} />
        </Card>
        <Card>
          <SectionTitle title="Hour of day" detail="Peak tariff hours are 08:00–11:00 and 18:00–21:00 IST." />
          <BarSeries data={data.hourly} dataKey="utilization_pct" name="Utilization" />
          <p className="mt-2 text-sm text-ink-soft">Peak {pct(data.peak.utilization_pct)} · off-peak {pct(data.off_peak.utilization_pct)}. Maintenance excluded {hours(data.excluded_maintenance_hours)}. Fault time excluded {hours(data.excluded_downtime_hours)}.</p>
        </Card>
      </div>
      <div className="grid gap-4 xl:grid-cols-2">
        <NamedTable title="By charger type" rows={data.by_charger_type} />
        <NamedTable title="By power rating" rows={data.by_power_rating} />
      </div>
      <Card className="overflow-x-auto">
        <SectionTitle title="Connectors under 10% utilization" />
        <table className="w-full text-left text-sm">
          <thead className="text-xs uppercase text-ink-faint"><tr><th className="py-2">Connector</th><th>Type</th><th className="text-right">Utilization</th><th className="text-right">Energy</th></tr></thead>
          <tbody>
            {(data.low_utilization ?? []).slice(0, 12).map((row) => (
              <tr key={row.connector_id} className="border-t border-line">
                <td className="py-2"><Link className="font-medium hover:underline" href={href(`/stations/${row.station_id}`)}>{row.connector_id}</Link><div className="text-xs text-ink-faint">{row.station_name}</div></td>
                <td>{row.charger_type} {row.rated_power_kw} kW</td>
                <td className="text-right tabular">{pct(row.utilization_pct)}</td>
                <td className="text-right tabular">{kwh(row.energy_kwh)}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {(data.low_utilization ?? []).length === 0 && <p className="text-sm text-ink-soft">No connector in this filter is under 10% with a full day of available time.</p>}
      </Card>
      <Definitions items={data.definitions} />
    </div>
  );
}

function NamedTable({ title, rows }: { title: string; rows: UtilizationReport["by_charger_type"] }) {
  return (
    <Card>
      <SectionTitle title={title} />
      <table className="w-full text-left text-sm">
        <tbody>
          {rows.map((row) => (
            <tr key={row.key} className="border-t border-line">
              <td className="py-2">{row.label}<div className="text-xs text-ink-faint">{row.connector_count} connectors</div></td>
              <td className="text-right tabular">{pct(row.utilization_pct)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}

function Revenue({ data, href }: { data: RevenueReport; href: (path: string) => string }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Kpi label="Revenue" value={inr(data.revenue_inr)} />
        <Kpi label="Electricity cost" value={inr(data.electricity_cost_inr)} detail={data.energy_cost_per_kwh === null ? undefined : `${inr(data.energy_cost_per_kwh, 2)} / kWh`} />
        <Kpi label="Gross margin" value={inr(data.gross_margin_inr)} detail={pct(data.gross_margin_pct)} />
        <Kpi label="Revenue / kWh" value={data.revenue_per_kwh === null ? "—" : inr(data.revenue_per_kwh, 2)} detail={`Per connector ${inr(data.revenue_per_connector_inr)}`} />
      </div>
      <Card>
        <SectionTitle title="Revenue and margin by day" detail="Attributed to the session start date." />
        <TrendChart data={dailyLabels(data.daily)} lines={[{ key: "revenue_inr", name: "Revenue", color: "#1e4f86" }, { key: "gross_margin_inr", name: "Gross margin", color: "#1c6b45" }]} />
      </Card>
      <Card className="overflow-x-auto">
        <SectionTitle title="Stations" detail="Sorted by revenue. Bottom of the margin list is the lowest gross-margin percentage among stations that delivered energy." />
        <table className="w-full min-w-[760px] text-left text-sm">
          <thead className="text-xs uppercase text-ink-faint"><tr><th className="py-2">Station</th><th className="text-right">Revenue</th><th className="text-right">Cost</th><th className="text-right">Margin</th><th className="text-right">₹/kWh</th></tr></thead>
          <tbody>
            {(data.by_station ?? []).map((row) => (
              <tr key={row.station_id} className="border-t border-line">
                <td className="py-2"><Link className="font-medium hover:underline" href={href(`/stations/${row.station_id}`)}>{row.station_name}</Link></td>
                <td className="text-right tabular">{inr(row.revenue_inr)}</td>
                <td className="text-right tabular">{inr(row.electricity_cost_inr)}</td>
                <td className="text-right tabular">{pct(row.gross_margin_pct)}</td>
                <td className="text-right tabular">{row.revenue_per_kwh === null ? "—" : inr(row.revenue_per_kwh, 2)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
      <Card>
        <SectionTitle title="Included in gross margin" />
        <ul className="list-disc space-y-1 pl-5 text-sm text-ink-soft">{data.included_cost_components.map((item) => <li key={item}>{item}</li>)}</ul>
        <p className="mt-3 text-sm font-medium">Not included, so this is not net profit</p>
        <ul className="mt-1 list-disc space-y-1 pl-5 text-sm text-ink-soft">{data.excluded_cost_components.map((item) => <li key={item}>{item}</li>)}</ul>
      </Card>
      <Definitions items={data.definitions} />
    </div>
  );
}

function Energy({ data, href }: { data: EnergyReport; href: (path: string) => string }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Kpi label="Energy delivered" value={kwh(data.total_energy_kwh)} />
        <Kpi label="Avg session" value={data.avg_session_energy_kwh === null ? "—" : `${num(data.avg_session_energy_kwh, 1)} kWh`} />
        <Kpi label="Avg charging power" value={data.avg_charging_power_kw === null ? "—" : `${num(data.avg_charging_power_kw, 1)} kW`} detail="Energy ÷ charging hours" />
        <Kpi label="Peak observed" value={data.peak_observed_power_kw === null ? "—" : `${num(data.peak_observed_power_kw, 1)} kW`} detail="Highest session reading" />
      </div>
      <div className="grid gap-4 xl:grid-cols-2">
        <Card>
          <SectionTitle title="Daily throughput" />
          <TrendChart data={dailyLabels(data.daily)} lines={[{ key: "energy_kwh", name: "Energy", color: "#1c6b45" }]} />
        </Card>
        <Card>
          <SectionTitle title="Energy by hour of day" detail="Each session's energy is spread across the hours it was charging." />
          <BarSeries data={data.hourly} dataKey="energy_kwh" name="Energy" color="#1e4f86" />
        </Card>
      </div>
      <Card className="overflow-x-auto">
        <SectionTitle title="Rated power versus observed charging power" detail="Observed power is session energy divided by charging hours. It is not a utility-meter reading." />
        <table className="w-full text-left text-sm">
          <thead className="text-xs uppercase text-ink-faint"><tr><th className="py-2">Group</th><th className="text-right">Nameplate</th><th className="text-right">Avg observed</th><th className="text-right">Peak</th><th className="text-right">Observed / rated</th></tr></thead>
          <tbody>
            {(data.rated_vs_observed ?? []).map((row) => (
              <tr key={row.label} className="border-t border-line">
                <td className="py-2">{row.label}</td>
                <td className="text-right tabular">{num(row.rated_power_kw, 1)} kW</td>
                <td className="text-right tabular">{row.avg_observed_kw === null ? "—" : `${num(row.avg_observed_kw, 1)} kW`}</td>
                <td className="text-right tabular">{row.peak_observed_kw === null ? "—" : `${num(row.peak_observed_kw, 1)} kW`}</td>
                <td className="text-right tabular">{row.observed_to_rated === null ? "—" : pct(row.observed_to_rated * 100)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
      <div className="grid gap-4 xl:grid-cols-2">
        <Demand title="Higher energy delivery" rows={data.high_demand ?? []} href={href} />
        <Demand title="Lower energy per connector" rows={data.low_demand ?? []} href={href} />
      </div>
      {data.notes.map((note) => <p key={note} className="text-xs leading-5 text-ink-soft">{note}</p>)}
    </div>
  );
}

function Demand({ title, rows, href }: { title: string; rows: EnergyReport["high_demand"]; href: (path: string) => string }) {
  return (
    <Card>
      <SectionTitle title={title} />
      <ul className="space-y-3">
        {rows.map((row) => (
          <li key={row.station_id} className="border-t border-line pt-3 first:border-0 first:pt-0">
            <Link href={href(`/stations/${row.station_id}`)} className="text-sm font-medium hover:underline">{row.station_name}</Link>
            <p className="text-xs text-ink-soft">{kwh(row.energy_kwh)} · {num(row.energy_per_connector_per_day, 1)} kWh/connector/day · busiest hour {row.peak_hour_label} at {pct(row.peak_hour_utilization_pct)}</p>
            <p className="text-xs text-ink-faint">{row.note}</p>
          </li>
        ))}
      </ul>
    </Card>
  );
}

function Sessions({ data }: { data: SessionReport }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Kpi label="Sessions" value={num(data.session_count)} />
        <Kpi label="Success rate" value={pct(data.success_rate_pct)} detail={`${num(data.completed)} completed`} />
        <Kpi label="Failure rate" value={pct(data.failure_rate_pct)} detail={`${num(data.failed)} failed · ${num(data.interrupted)} interrupted`} />
        <Kpi label="Avg duration" value={minutes(data.avg_duration_min)} detail={data.avg_energy_kwh === null ? undefined : `${num(data.avg_energy_kwh, 1)} kWh average`} />
      </div>
      <div className="grid gap-4 xl:grid-cols-2">
        <Card>
          <SectionTitle title="Demand by hour" detail="Session starts, not prorated energy." />
          <BarSeries data={data.by_hour} dataKey="session_starts" name="Session starts" />
        </Card>
        <Card>
          <SectionTitle title="Day of week" />
          <BarSeries data={(data.by_weekday ?? []).map((row) => ({ label: row.label.slice(0, 3), sessions: row.sessions }))} dataKey="sessions" name="Sessions" color="#1e4f86" />
        </Card>
      </div>
      <Card>
        <SectionTitle title="Busiest three-hour windows" />
        <ul className="space-y-2 text-sm">
          {(data.peak_windows ?? []).map((window) => (
            <li key={window.label}>{window.label}: {num(window.sessions)} sessions ({pct(window.share_pct)})</li>
          ))}
        </ul>
      </Card>
      <Card>
        <SectionTitle title="Failed and interrupted reasons" />
        {(data.failed_by_reason ?? []).length === 0 && <p className="text-sm text-ink-soft">None in this filter.</p>}
        <ul className="space-y-1 text-sm">{(data.failed_by_reason ?? []).map((row) => <li key={row.reason}>{titleCase(row.reason)} · {row.count}</li>)}</ul>
      </Card>
      <Definitions items={data.definitions} />
    </div>
  );
}

function Reliability({ data, href }: { data: ReliabilityReport; href: (path: string) => string }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Kpi label="Uptime" value={pct(data.uptime_pct)} detail={`${hours(data.downtime_hours)} connector-hours down`} />
        <Kpi label="Faults" value={num(data.fault_count)} detail={`${num(data.open_fault_count)} still open`} />
        <Kpi label="MTTR" value={data.mttr_hours === null ? "—" : hours(data.mttr_hours)} detail="Resolved faults only" />
        <Kpi label="MTBF" value={data.mtbf_hours === null ? "No failures" : hours(data.mtbf_hours)} detail="Omitted when there are no faults" />
      </div>
      <Card>
        <SectionTitle title="Estimated revenue at risk" detail={data.revenue_at_risk_assumption} />
        <p className="text-2xl font-semibold tabular">{inr(data.revenue_at_risk_inr)}</p>
        <p className="mt-2 text-sm text-ink-soft">Maintenance excluded from the uptime denominator: {hours(data.maintenance_hours_excluded)}.</p>
        {data.longest_ongoing && (
          <p className="mt-2 text-sm">Longest open outage: {data.longest_ongoing.station_id} {data.longest_ongoing.connector_id || "site"} · {titleCase(data.longest_ongoing.category)} · {num(data.longest_ongoing.duration_hours, 1)} h. {data.longest_ongoing.description}</p>
        )}
      </Card>
      <div className="grid gap-4 xl:grid-cols-2">
        <Card>
          <SectionTitle title="Fault categories" />
          <BarSeries data={(data.by_category ?? []).map((row) => ({ label: titleCase(row.category).split(" ")[0], count: row.count }))} dataKey="count" name="Faults" color="#9d3140" />
        </Card>
        <Card className="overflow-x-auto">
          <SectionTitle title="Stations by downtime" />
          <table className="w-full text-left text-sm">
            <tbody>
              {(data.by_station ?? []).slice(0, 8).map((row) => (
                <tr key={row.station_id} className="border-t border-line">
                  <td className="py-2"><Link className="font-medium hover:underline" href={href(`/stations/${row.station_id}`)}>{row.station_name}</Link><div className="text-xs text-ink-faint">{row.fault_count} faults · uptime {pct(row.uptime_pct)}</div></td>
                  <td className="text-right tabular">{hours(row.downtime_hours)}<div className="text-xs text-ink-faint">{inr(row.revenue_at_risk_inr)} at risk</div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>
      <Definitions items={data.definitions} />
    </div>
  );
}
