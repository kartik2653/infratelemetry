"use client";

import type { RangeKey } from "@/lib/filters";
import type { StationIndex } from "@/lib/types";

export function RangeFilter({
  range,
  chargerType,
  onRange,
  onCharger,
  showCharger = true,
}: {
  range: RangeKey;
  chargerType: string;
  onRange: (range: RangeKey) => void;
  onCharger?: (value: string) => void;
  showCharger?: boolean;
}) {
  const ranges: RangeKey[] = ["7d", "30d", "90d"];
  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="flex rounded-lg border border-line bg-white p-0.5">
        {ranges.map((item) => (
          <button
            key={item}
            type="button"
            aria-pressed={range === item}
            onClick={() => onRange(item)}
            className={`rounded-md px-3 py-1.5 text-sm ${range === item ? "bg-pine text-white" : "text-ink-soft"}`}
          >
            {item === "90d" ? "90 days" : item === "30d" ? "30 days" : "7 days"}
          </button>
        ))}
      </div>
      {showCharger && onCharger && (
        <label className="flex items-center gap-2 text-sm text-ink-soft">
          Charger
          <select
            value={chargerType}
            onChange={(event) => onCharger(event.target.value)}
            className="rounded-md border border-line bg-white px-2 py-1.5 text-sm text-ink"
          >
            <option value="">All types</option>
            <option value="AC">AC</option>
            <option value="DC">DC</option>
          </select>
        </label>
      )}
    </div>
  );
}

export function StationSelect({
  stations,
  value,
  onChange,
}: {
  stations: StationIndex[];
  value: string;
  onChange: (id: string) => void;
}) {
  return (
    <label className="flex items-center gap-2 text-sm text-ink-soft">
      Station
      <select value={value} onChange={(event) => onChange(event.target.value)} className="max-w-[240px] rounded-md border border-line bg-white px-2 py-1.5 text-sm text-ink">
        <option value="">All stations</option>
        {stations.map((station) => (
          <option key={station.id} value={station.id}>
            {station.id} · {station.name}
          </option>
        ))}
      </select>
    </label>
  );
}
