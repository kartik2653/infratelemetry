"use client";

import { RangeFilter, StationSelect } from "@/components/filters";
import { InsightCard } from "@/components/insight-card";
import { ErrorState, Empty, Loading } from "@/components/states";
import { useFilters } from "@/lib/filters";
import { qs } from "@/lib/api";
import { useGet } from "@/lib/use-get";
import type { Insight } from "@/lib/types";

export default function InsightsPage() {
  const { range, chargerType, stationId, from, to, update, href, clock } = useFilters();
  const res = useGet<{ insights: Insight[] }>(`/api/v1/insights${qs({ from, to, charger_type: chargerType, station_id: stationId })}`);
  const insights = res.data?.data.insights || [];

  return (
    <div className="space-y-4">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">Insights and recommendations</h1>
          <p className="mt-1 max-w-3xl text-sm text-ink-soft">Each card is a rule applied to the numbers in the selected window. Evidence uses those numbers. Estimated impact is shown only where the arithmetic is defensible, and it is not a forecast of net profit or price elasticity.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <RangeFilter range={range} chargerType={chargerType} onRange={(value) => update({ range: value })} onCharger={(value) => update({ charger_type: value || null })} />
          <StationSelect stations={clock.stations} value={stationId} onChange={(id) => update({ station_id: id || null })} />
        </div>
      </header>
      {res.loading && <Loading />}
      {res.error && <ErrorState message={res.error} onRetry={res.reload} />}
      {!res.loading && !res.error && insights.length === 0 && <Empty>No recommendations fired for this filter.</Empty>}
      <div className="grid gap-3 xl:grid-cols-2">
        {insights.map((insight) => (
          <InsightCard key={insight.id} insight={insight} href={insight.station_ids[0] ? href(`/stations/${insight.station_ids[0]}`) : undefined} />
        ))}
      </div>
    </div>
  );
}
