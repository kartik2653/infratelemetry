"use client";

import Link from "next/link";
import { AreaTrend, TrendChart, dailyLabels } from "@/components/charts";
import { RangeFilter } from "@/components/filters";
import { InsightCard } from "@/components/insight-card";
import { ErrorState, Loading } from "@/components/states";
import { Badge, Card, Definitions, Kpi, SectionTitle, StatusMix } from "@/components/ui";
import { qs } from "@/lib/api";
import { useFilters } from "@/lib/filters";
import { hours, inr, kwh, minutes, num, pct, when } from "@/lib/format";
import { useGet } from "@/lib/use-get";
import type { Dashboard, StationRow, Trends } from "@/lib/types";

export default function OverviewPage() {
  const { range, chargerType, from, to, update, href } = useFilters();
  const summary = useGet<Dashboard>(`/api/v1/dashboard/summary${qs({ from, to, charger_type: chargerType })}`);
  const trends = useGet<Trends>(`/api/v1/dashboard/trends${qs({ from, to, charger_type: chargerType })}`);
  const data = summary.data?.data;
  const kpis = data?.summary.kpis;

  return (
    <div className="space-y-5">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">Network overview</h1>
          <p className="mt-1 text-sm text-ink-soft">Performance, reliability, and gross margin for the simulated Bengaluru network. Connector status is the state at the demo clock.</p>
        </div>
        <RangeFilter range={range} chargerType={chargerType} onRange={(value) => update({ range: value })} onCharger={(value) => update({ charger_type: value || null })} />
      </header>

      {summary.loading && <Loading />}
      {summary.error && <ErrorState message={summary.error} onRetry={summary.reload} />}
      {data && kpis && (
        <>
          <Card>
            <SectionTitle title="Connectors right now" detail="Current status at the demo clock, limited to the charger-type filter. This is not utilization." />
            <StatusMix counts={data.summary.connectors_by_status} />
          </Card>
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <Kpi label="Stations" value={num(kpis.station_count)} detail={`${num(kpis.connector_count)} connectors`} href={href("/stations")} />
            <Kpi label="Uptime" value={pct(kpis.uptime_pct)} detail={`${hours(kpis.downtime_hours)} unplanned downtime`} href={href("/analytics", { tab: "reliability" })} />
            <Kpi label="Utilization" value={pct(kpis.utilization_pct)} detail={`${hours(kpis.charging_hours)} charging / ${hours(kpis.available_hours)} available`} href={href("/analytics", { tab: "utilization" })} />
            <Kpi label="Sessions" value={num(kpis.session_count)} detail={`${pct(data.summary.session_stats.success_rate_pct)} completed`} href={href("/analytics", { tab: "sessions" })} />
            <Kpi label="Energy delivered" value={kwh(kpis.energy_kwh)} detail={`${num(kpis.avg_energy_per_session_kwh, 1)} kWh average session`} href={href("/analytics", { tab: "energy" })} />
            <Kpi label="Charging revenue" value={inr(kpis.revenue_inr)} detail="Tariff × energy" href={href("/analytics", { tab: "revenue" })} />
            <Kpi label="Electricity cost" value={inr(kpis.electricity_cost_inr)} detail="Procurement only" href={href("/analytics", { tab: "revenue" })} />
            <Kpi label="Gross margin" value={inr(kpis.gross_margin_inr)} detail={`${pct(kpis.gross_margin_pct)} · not net profit`} href={href("/analytics", { tab: "revenue" })} />
          </div>
          <div className="grid gap-3 sm:grid-cols-3">
            <Kpi label="Avg session duration" value={minutes(kpis.avg_session_duration_min)} detail="Plug time, including idle tail" />
            <Kpi label="Avg energy / session" value={kpis.avg_energy_per_session_kwh === null ? "—" : `${num(kpis.avg_energy_per_session_kwh, 1)} kWh`} />
            <Kpi label="Failed + interrupted" value={num(data.summary.session_stats.failed + data.summary.session_stats.interrupted)} detail={`${pct(data.summary.session_stats.failure_rate_pct)} failed`} />
          </div>
        </>
      )}

      {trends.loading && <Loading label="Loading trends…" />}
      {trends.error && <ErrorState message={trends.error} onRetry={trends.reload} />}
      {trends.data && (
        <div className="grid gap-4 xl:grid-cols-2">
          <Card>
            <SectionTitle title="Energy and revenue" detail="Whole sessions are attributed to the day they started. Revenue uses the right axis." />
            <AreaTrend data={dailyLabels(trends.data.data.daily)} />
          </Card>
          <Card>
            <SectionTitle title="Utilization and uptime" detail="Charging time over available time, and uptime after planned maintenance is excluded." />
            <TrendChart
              data={dailyLabels(trends.data.data.daily)}
              lines={[
                { key: "utilization_pct", name: "Utilization", color: "#1c6b45" },
                { key: "uptime_pct", name: "Uptime", color: "#1e4f86" },
              ]}
            />
          </Card>
        </div>
      )}

      {data && (
        <div className="grid gap-4 xl:grid-cols-2">
          <StationList title="Highest revenue" rows={data.summary.top_stations} hrefFor={(id) => href(`/stations/${id}`)} />
          <StationList title="Lowest utilization" rows={data.summary.bottom_stations} hrefFor={(id) => href(`/stations/${id}`)} metric="utilization" />
        </div>
      )}

      {data && (
        <div className="grid gap-4 xl:grid-cols-2">
          <Card>
            <SectionTitle title="Critical alerts" detail="Open hardware and long-running faults from the in-memory alert list. Acknowledged items stay until you resolve them or restart the backend." />
            {data.critical_alerts.length === 0 && <p className="text-sm text-ink-soft">No open critical alerts.</p>}
            <ul className="space-y-3">
              {data.critical_alerts.map((alert) => (
                <li key={alert.id} className="border-t border-line pt-3 first:border-t-0 first:pt-0">
                  <div className="flex items-center gap-2"><Badge value={alert.severity} /><span className="text-xs text-ink-faint">{when(alert.timestamp)}</span></div>
                  <Link href={href(`/stations/${alert.station_id}`)} className="mt-1 block text-sm font-medium hover:underline">{alert.title}</Link>
                  <p className="text-xs text-ink-soft">{alert.observed_value}</p>
                </li>
              ))}
            </ul>
          </Card>
          <div className="space-y-3">
            <SectionTitle title="Highest-priority recommendations" detail="Generated from this date range by rules in the backend, not a language model." />
            {data.insights.length === 0 && <p className="text-sm text-ink-soft">No recommendations for this filter.</p>}
            {data.insights.map((insight) => (
              <InsightCard key={insight.id} insight={insight} href={insight.station_ids[0] ? href(`/stations/${insight.station_ids[0]}`) : href("/insights")} />
            ))}
            <Link href={href("/insights")} className="inline-block text-sm font-medium text-pine">All recommendations</Link>
          </div>
        </div>
      )}

      {data && <Definitions items={data.summary.definitions} />}
      {data && (
        <Card>
          <SectionTitle title="What gross margin includes" />
          <ul className="list-disc space-y-1 pl-5 text-sm text-ink-soft">
            {data.summary.included_cost_components.map((item) => <li key={item}>{item}</li>)}
          </ul>
          <p className="mt-3 text-sm font-medium">Excluded, so this is not net profit</p>
          <ul className="mt-1 list-disc space-y-1 pl-5 text-sm text-ink-soft">
            {data.summary.excluded_cost_components.map((item) => <li key={item}>{item}</li>)}
          </ul>
        </Card>
      )}
    </div>
  );
}

function StationList({ title, rows, hrefFor, metric = "revenue" }: { title: string; rows: StationRow[]; hrefFor: (id: string) => string; metric?: "revenue" | "utilization" }) {
  return (
    <Card>
      <SectionTitle title={title} />
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <caption className="sr-only">{title}</caption>
          <thead className="text-xs uppercase tracking-wide text-ink-faint">
            <tr>
              <th className="py-2 font-medium">Station</th>
              <th className="py-2 font-medium">Status</th>
              <th className="py-2 text-right font-medium">{metric === "revenue" ? "Revenue" : "Utilization"}</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id} className="border-t border-line">
                <td className="py-2">
                  <Link href={hrefFor(row.id)} className="font-medium hover:underline">{row.name}</Link>
                  <div className="text-xs text-ink-faint">{row.id} · {row.area}</div>
                </td>
                <td className="py-2"><Badge value={row.status} /></td>
                <td className="py-2 text-right tabular">{metric === "revenue" ? inr(row.revenue_inr) : pct(row.utilization_pct)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
