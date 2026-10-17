"use client";

import Link from "next/link";
import { useState } from "react";
import { StationSelect } from "@/components/filters";
import { ErrorState, Empty, Loading } from "@/components/states";
import { Badge, Card } from "@/components/ui";
import { apiPatch, qs } from "@/lib/api";
import { useFilters } from "@/lib/filters";
import { titleCase, when } from "@/lib/format";
import { useGet } from "@/lib/use-get";
import type { Alert } from "@/lib/types";

const categories = [
  "charger_offline",
  "repeated_connector_failure",
  "high_downtime",
  "utilization_drop",
  "abnormal_session_duration",
  "energy_anomaly",
  "revenue_at_risk",
  "repeated_payment_failure",
];

export default function AlertsPage() {
  const { clock, href, sp, update } = useFilters();
  const severity = sp.get("severity") || "";
  const status = sp.get("status") || "";
  const category = sp.get("category") || "";
  const stationId = sp.get("station_id") || "";
  const res = useGet<{ alerts: Alert[] }>(`/api/v1/alerts${qs({ severity, status, category, station_id: stationId })}`);
  const [pending, setPending] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  async function setStatus(id: string, next: string) {
    setPending(id);
    setActionError(null);
    try {
      await apiPatch(`/api/v1/alerts/${id}`, { status: next });
      res.reload();
      window.dispatchEvent(new Event("alerts-changed"));
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Could not update the alert");
    } finally {
      setPending(null);
    }
  }

  return (
    <div className="space-y-4">
      <header>
        <h1 className="text-xl font-semibold">Alerts</h1>
        <p className="mt-1 max-w-3xl text-sm text-ink-soft">Generated at backend startup from the trailing 7, 14, and 30 days. Acknowledge and resolve are stored in memory until the process restarts. The date control on other pages does not change this list.</p>
      </header>
      <div className="flex flex-wrap gap-2">
        <Select label="Severity" value={severity} onChange={(value) => update({ severity: value || null })} options={["critical", "warning", "info"]} />
        <Select label="Status" value={status} onChange={(value) => update({ status: value || null })} options={["open", "acknowledged", "resolved"]} />
        <Select label="Category" value={category} onChange={(value) => update({ category: value || null })} options={categories} />
        <StationSelect stations={clock.stations} value={stationId} onChange={(id) => update({ station_id: id || null })} />
      </div>
      {actionError && <ErrorState message={actionError} />}
      {res.loading && <Loading />}
      {res.error && <ErrorState message={res.error} onRetry={res.reload} />}
      {!res.loading && res.data && res.data.data.alerts.length === 0 && <Empty>No alerts match these filters.</Empty>}
      <div className="space-y-3">
        {res.data?.data.alerts.map((alert) => (
          <Card key={alert.id}>
            <div className="flex flex-wrap items-center gap-2">
              <Badge value={alert.severity} />
              <Badge value={alert.status} />
              <span className="text-xs text-ink-faint">{titleCase(alert.category)} · {when(alert.timestamp)}</span>
            </div>
            <h2 className="mt-2 text-sm font-semibold">{alert.title}</h2>
            <p className="mt-1 text-sm leading-6 text-ink-soft">{alert.description}</p>
            <p className="mt-2 text-sm"><span className="font-medium">{titleCase(alert.metric)}: </span>{alert.observed_value}</p>
            <p className="mt-1 text-sm"><span className="font-medium">Next step. </span>{alert.recommended_action}</p>
            <div className="mt-3 flex flex-wrap items-center gap-2">
              <Link href={href(`/stations/${alert.station_id}`)} className="text-sm font-medium text-pine">{alert.station_name}{alert.connector_id ? ` · ${alert.connector_id}` : ""}</Link>
              {alert.status !== "acknowledged" && <button type="button" disabled={pending === alert.id} onClick={() => setStatus(alert.id, "acknowledged")} className="rounded-md border border-line px-2 py-1 text-sm">Acknowledge</button>}
              {alert.status !== "resolved" && <button type="button" disabled={pending === alert.id} onClick={() => setStatus(alert.id, "resolved")} className="rounded-md bg-pine px-2 py-1 text-sm text-white">Resolve</button>}
              {alert.status !== "open" && <button type="button" disabled={pending === alert.id} onClick={() => setStatus(alert.id, "open")} className="rounded-md border border-line px-2 py-1 text-sm">Reopen</button>}
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}

function Select({ label, value, onChange, options }: { label: string; value: string; onChange: (value: string) => void; options: string[] }) {
  return (
    <label className="flex items-center gap-2 text-sm text-ink-soft">
      {label}
      <select value={value} onChange={(event) => onChange(event.target.value)} className="rounded-md border border-line bg-white px-2 py-1.5 text-sm text-ink">
        <option value="">All</option>
        {options.map((option) => <option key={option} value={option}>{titleCase(option)}</option>)}
      </select>
    </label>
  );
}
