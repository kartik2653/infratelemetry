"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { RangeFilter } from "@/components/filters";
import { ErrorState, Empty, Loading } from "@/components/states";
import { Badge, Card } from "@/components/ui";
import { qs } from "@/lib/api";
import { useFilters } from "@/lib/filters";
import { inr, kwh, pct } from "@/lib/format";
import { useGet } from "@/lib/use-get";
import type { StationRow } from "@/lib/types";

const columns = [
  { key: "name", label: "Station" },
  { key: "utilization", label: "Utilization" },
  { key: "uptime", label: "Uptime" },
  { key: "energy", label: "Energy" },
  { key: "revenue", label: "Revenue" },
  { key: "margin", label: "Gross margin" },
  { key: "sessions", label: "Sessions" },
] as const;

export default function StationsPage() {
  const { range, chargerType, from, to, update, href, sp } = useFilters();
  const sort = sp.get("sort") || "revenue";
  const order = sp.get("order") || "desc";
  const page = Number(sp.get("page") || "1");
  const [search, setSearch] = useState(sp.get("q") || "");

  useEffect(() => {
    const handle = setTimeout(() => {
      if ((sp.get("q") || "") !== search) update({ q: search || null, page: null });
    }, 300);
    return () => clearTimeout(handle);
  }, [search]);

  const path = `/api/v1/stations${qs({ from, to, charger_type: chargerType, q: sp.get("q") || "", sort, order, page, limit: 10 })}`;
  const res = useGet<{ stations: StationRow[] }>(path);
  const rows = res.data?.data.stations || [];
  const meta = res.data?.meta;

  function toggleSort(key: string) {
    const nextOrder = sort === key && order === "desc" ? "asc" : "desc";
    update({ sort: key, order: key === "name" && sort !== key ? "asc" : nextOrder, page: null });
  }

  return (
    <div className="space-y-4">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">Station explorer</h1>
          <p className="mt-1 text-sm text-ink-soft">Metrics follow the selected dates. Status is the connector state at the demo clock.</p>
        </div>
        <RangeFilter range={range} chargerType={chargerType} onRange={(value) => update({ range: value, page: null })} onCharger={(value) => update({ charger_type: value || null, page: null })} />
      </header>
      <input
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        placeholder="Search name, area, or station ID"
        aria-label="Search stations"
        className="w-full max-w-md rounded-lg border border-line bg-white px-3 py-2 text-sm"
      />
      {res.loading && <Loading />}
      {res.error && <ErrorState message={res.error} onRetry={res.reload} />}
      {!res.loading && !res.error && rows.length === 0 && <Empty>No stations match these filters.</Empty>}
      {rows.length > 0 && (
        <Card className="overflow-x-auto">
          <table className="w-full min-w-[860px] text-left text-sm">
            <caption className="sr-only">Charging stations</caption>
            <thead className="text-xs uppercase tracking-wide text-ink-faint">
              <tr>
                {columns.map((column) => (
                  <th key={column.key} className="py-2 pr-3 font-medium">
                    <button type="button" onClick={() => toggleSort(column.key)} className="inline-flex items-center gap-1">
                      {column.label}
                      {sort === column.key ? (order === "asc" ? "↑" : "↓") : ""}
                    </button>
                  </th>
                ))}
                <th className="py-2 font-medium">Status</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.id} className="border-t border-line">
                  <td className="py-3 pr-3">
                    <Link href={href(`/stations/${row.id}`)} className="font-medium hover:underline">{row.name}</Link>
                    <div className="text-xs text-ink-faint">{row.id} · {row.area} · {(row.charger_types || []).join(" / ") || "—"} · {row.connector_count} connectors · {row.max_power_kw} kW max</div>
                  </td>
                  <td className="py-3 pr-3 tabular">{pct(row.utilization_pct)}</td>
                  <td className="py-3 pr-3 tabular">{pct(row.uptime_pct)}</td>
                  <td className="py-3 pr-3 tabular">{kwh(row.energy_kwh)}</td>
                  <td className="py-3 pr-3 tabular">{inr(row.revenue_inr)}</td>
                  <td className="py-3 pr-3 tabular">{pct(row.gross_margin_pct)}</td>
                  <td className="py-3 pr-3 tabular">{row.session_count}</td>
                  <td className="py-3"><Badge value={row.status} /></td>
                </tr>
              ))}
            </tbody>
          </table>
          {meta && (
            <div className="mt-3 flex items-center justify-between text-sm text-ink-soft">
              <span>{meta.total || 0} stations · page {meta.page || 1} of {meta.total_pages || 1}</span>
              <div className="flex gap-2">
                <button type="button" disabled={(meta.page || 1) <= 1} onClick={() => update({ page: String((meta.page || 1) - 1) })} className="rounded-md border border-line px-2 py-1 disabled:opacity-40">Previous</button>
                <button type="button" disabled={(meta.page || 1) >= (meta.total_pages || 1)} onClick={() => update({ page: String((meta.page || 1) + 1) })} className="rounded-md border border-line px-2 py-1 disabled:opacity-40">Next</button>
              </div>
            </div>
          )}
        </Card>
      )}
    </div>
  );
}
