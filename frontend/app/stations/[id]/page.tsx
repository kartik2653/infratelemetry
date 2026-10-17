"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { BarSeries, TrendChart, dailyLabels } from "@/components/charts";
import { RangeFilter } from "@/components/filters";
import { InsightCard } from "@/components/insight-card";
import { ErrorState, Empty, Loading } from "@/components/states";
import { Badge, Card, Definitions, Kpi, SectionTitle } from "@/components/ui";
import { qs } from "@/lib/api";
import { useFilters } from "@/lib/filters";
import { hours, inr, kwh, minutes, num, pct, titleCase, when } from "@/lib/format";
import { useGet } from "@/lib/use-get";
import type { Session, StationDetail } from "@/lib/types";

export default function StationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const { range, chargerType, from, to, update, href } = useFilters();
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(1);
  const detail = useGet<StationDetail>(id ? `/api/v1/stations/${id}${qs({ from, to, charger_type: chargerType })}` : null);
  const sessions = useGet<{ sessions: Session[] }>(id ? `/api/v1/stations/${id}/sessions${qs({ from, to, status, page, limit: 8 })}` : null);
  const station = detail.data?.data;

  return (
    <div className="space-y-5">
      <Link href={href("/stations")} className="text-sm text-pine">Back to stations</Link>
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">{station?.station.name || id}</h1>
          <p className="mt-1 text-sm text-ink-soft">{station ? `${station.station.address} · ${station.station.id}` : "Station detail"}</p>
        </div>
        <RangeFilter range={range} chargerType={chargerType} onRange={(value) => { setPage(1); update({ range: value }); }} onCharger={(value) => { setPage(1); update({ charger_type: value || null }); }} />
      </header>
      {detail.loading && <Loading />}
      {detail.error && <ErrorState message={detail.error} onRetry={detail.reload} />}
      {station && (
        <>
          <div className="flex flex-wrap items-center gap-2">
            <Badge value={station.station.status} />
            {(station.station.charger_types || []).map((type) => <Badge key={type} value={type} />)}
            <span className="text-sm text-ink-soft">{station.station.connector_count} connectors · max {station.station.max_power_kw} kW</span>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <Kpi label="Utilization" value={pct(station.kpis.utilization_pct)} detail={`${hours(station.kpis.charging_hours)} charging`} />
            <Kpi label="Uptime" value={pct(station.kpis.uptime_pct)} detail={`${hours(station.kpis.downtime_hours)} downtime`} />
            <Kpi label="Energy" value={kwh(station.kpis.energy_kwh)} detail={`${num(station.kpis.session_count)} sessions`} />
            <Kpi label="Gross margin" value={inr(station.kpis.gross_margin_inr)} detail={`${pct(station.kpis.gross_margin_pct)} · not net profit`} />
          </div>
          <Card>
            <SectionTitle title="Tariffs used for this station" detail={`Peak window: ${station.tariffs.peak_hours}. Electricity cost is procurement, not the driver tariff.`} />
            <div className="grid gap-3 text-sm sm:grid-cols-4">
              <p>Peak tariff <span className="block font-medium tabular">{inr(station.tariffs.peak_tariff_inr_per_kwh, 2)} / kWh</span></p>
              <p>Off-peak tariff <span className="block font-medium tabular">{inr(station.tariffs.off_peak_tariff_inr_per_kwh, 2)} / kWh</span></p>
              <p>Peak energy cost <span className="block font-medium tabular">{inr(station.tariffs.peak_energy_cost_inr_per_kwh, 2)} / kWh</span></p>
              <p>Off-peak energy cost <span className="block font-medium tabular">{inr(station.tariffs.off_peak_energy_cost_inr_per_kwh, 2)} / kWh</span></p>
            </div>
          </Card>
          <div className="grid gap-4 xl:grid-cols-2">
            <Card>
              <SectionTitle title="Utilization and uptime" />
              <TrendChart data={dailyLabels(station.daily)} lines={[{ key: "utilization_pct", name: "Utilization", color: "#1c6b45" }, { key: "uptime_pct", name: "Uptime", color: "#1e4f86" }]} />
            </Card>
            <Card>
              <SectionTitle title="Session starts by hour" detail="Count of sessions that started in each clock hour, IST." />
              <BarSeries data={station.hourly.map((row) => ({ ...row }))} dataKey="session_starts" name="Sessions" />
            </Card>
          </div>
          <Card className="overflow-x-auto">
            <SectionTitle title="Connectors" />
            <table className="w-full min-w-[760px] text-left text-sm">
              <thead className="text-xs uppercase tracking-wide text-ink-faint">
                <tr>
                  <th className="py-2 font-medium">Connector</th>
                  <th className="py-2 font-medium">Type</th>
                  <th className="py-2 font-medium">Status</th>
                  <th className="py-2 text-right font-medium">Utilization</th>
                  <th className="py-2 text-right font-medium">Uptime</th>
                  <th className="py-2 text-right font-medium">Energy</th>
                  <th className="py-2 text-right font-medium">Margin</th>
                </tr>
              </thead>
              <tbody>
                {station.connectors.map((connector) => (
                  <tr key={connector.id} className="border-t border-line">
                    <td className="py-2">{connector.id}<div className="text-xs text-ink-faint">{connector.standard} · {connector.rated_power_kw} kW</div></td>
                    <td>{connector.charger_type}</td>
                    <td><Badge value={connector.status} /></td>
                    <td className="text-right tabular">{pct(connector.utilization_pct)}</td>
                    <td className="text-right tabular">{pct(connector.uptime_pct)}</td>
                    <td className="text-right tabular">{kwh(connector.energy_kwh)}</td>
                    <td className="text-right tabular">{pct(connector.gross_margin_pct)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {station.connectors.length === 0 && <p className="text-sm text-ink-soft">No connectors match this charger type.</p>}
          </Card>
          <Card>
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <SectionTitle title="Sessions" detail="Sessions that started in the selected window." />
              <select value={status} onChange={(event) => { setStatus(event.target.value); setPage(1); }} aria-label="Session status" className="rounded-md border border-line px-2 py-1.5 text-sm">
                <option value="">All statuses</option>
                <option value="completed">Completed</option>
                <option value="failed">Failed</option>
                <option value="interrupted">Interrupted</option>
                <option value="in_progress">In progress</option>
              </select>
            </div>
            {sessions.loading && <Loading label="Loading sessions…" />}
            {sessions.error && <ErrorState message={sessions.error} onRetry={sessions.reload} />}
            {sessions.data && sessions.data.data.sessions.length === 0 && <Empty>No sessions in this filter.</Empty>}
            {sessions.data && sessions.data.data.sessions.length > 0 && (
              <div className="overflow-x-auto">
                <table className="w-full min-w-[820px] text-left text-sm">
                  <thead className="text-xs uppercase tracking-wide text-ink-faint">
                    <tr>
                      <th className="py-2 font-medium">Start</th>
                      <th className="py-2 font-medium">Connector</th>
                      <th className="py-2 font-medium">Status</th>
                      <th className="py-2 text-right font-medium">Energy</th>
                      <th className="py-2 text-right font-medium">Duration</th>
                      <th className="py-2 text-right font-medium">Revenue</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sessions.data.data.sessions.map((session) => (
                      <tr key={session.id} className="border-t border-line">
                        <td className="py-2">{when(session.start)}<div className="text-xs text-ink-faint">{session.id}</div></td>
                        <td>{session.connector_id}</td>
                        <td><Badge value={session.status} />{session.termination_reason && <div className="text-xs text-ink-faint">{titleCase(session.termination_reason)}</div>}</td>
                        <td className="text-right tabular">{num(session.energy_kwh, 1)} kWh</td>
                        <td className="text-right tabular">{minutes(session.duration_min)}<div className="text-xs text-ink-faint">{minutes(session.charging_min)} charging</div></td>
                        <td className="text-right tabular">{inr(session.revenue_inr)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                <div className="mt-3 flex justify-end gap-2 text-sm">
                  <button type="button" disabled={page <= 1} onClick={() => setPage((n) => n - 1)} className="rounded-md border border-line px-2 py-1 disabled:opacity-40">Previous</button>
                  <button type="button" disabled={page >= (sessions.data.meta.total_pages || 1)} onClick={() => setPage((n) => n + 1)} className="rounded-md border border-line px-2 py-1 disabled:opacity-40">Next</button>
                </div>
              </div>
            )}
          </Card>
          <Card className="overflow-x-auto">
            <SectionTitle title="Faults in this window" />
            {station.faults.length === 0 && <p className="text-sm text-ink-soft">No faults overlap this window.</p>}
            {station.faults.length > 0 && (
              <table className="w-full min-w-[720px] text-left text-sm">
                <thead className="text-xs uppercase tracking-wide text-ink-faint">
                  <tr>
                    <th className="py-2 font-medium">When</th>
                    <th className="py-2 font-medium">Category</th>
                    <th className="py-2 font-medium">Connector</th>
                    <th className="py-2 text-right font-medium">Hours</th>
                  </tr>
                </thead>
                <tbody>
                  {station.faults.slice(0, 12).map((fault) => (
                    <tr key={fault.id} className="border-t border-line">
                      <td className="py-2">{when(fault.start)}{fault.ongoing && <div className="text-xs text-rose">Ongoing</div>}</td>
                      <td>{titleCase(fault.category)}<div className="max-w-md text-xs text-ink-faint">{fault.description}</div></td>
                      <td>{fault.connector_id || "Site"}</td>
                      <td className="text-right tabular">{num(fault.duration_hours, 1)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </Card>
          <div className="space-y-3">
            <SectionTitle title="Recommendations for this station" />
            {station.insights.length === 0 && <p className="text-sm text-ink-soft">No station-specific recommendation fired for this filter.</p>}
            {station.insights.map((insight) => <InsightCard key={insight.id} insight={insight} href={href(`/stations/${id}`)} />)}
          </div>
          <Definitions items={station.definitions} />
        </>
      )}
    </div>
  );
}
