"use client";

import { Area, AreaChart, Bar, BarChart, CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { dayLabel, inr, num, pct } from "@/lib/format";

type Row = Record<string, string | number | null>;

function Tip({ active, payload, label }: { active?: boolean; payload?: { name?: string; value?: number | string; color?: string; dataKey?: string }[]; label?: string }) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-lg border border-line bg-white px-3 py-2 text-xs shadow-card">
      <p className="mb-1 font-medium">{label}</p>
      {payload.map((item) => (
        <p key={String(item.dataKey)} className="flex justify-between gap-4">
          <span style={{ color: item.color }}>{item.name}</span>
          <span className="tabular">{formatTip(String(item.dataKey), item.value)}</span>
        </p>
      ))}
    </div>
  );
}

function formatTip(key: string, value: number | string | undefined) {
  if (value === undefined || value === null || value === "") return "—";
  if (typeof value !== "number") return String(value);
  if (key.includes("pct")) return pct(value);
  if (key.includes("inr") || key.includes("revenue") || key.includes("margin") || key.includes("cost")) return inr(value);
  if (key.includes("kwh") || key.includes("energy")) return `${num(value)} kWh`;
  return num(value, 1);
}

export function TrendChart({
  data,
  lines,
  height = 260,
}: {
  data: Row[];
  lines: { key: string; name: string; color: string }[];
  height?: number;
}) {
  return (
    <div style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
          <CartesianGrid stroke="#e2e6e1" vertical={false} />
          <XAxis dataKey="label" tick={{ fontSize: 11, fill: "#5d6b64" }} axisLine={false} tickLine={false} minTickGap={24} />
          <YAxis tick={{ fontSize: 11, fill: "#5d6b64" }} axisLine={false} tickLine={false} width={48} />
          <Tooltip content={<Tip />} />
          {lines.map((line) => (
            <Line key={line.key} type="monotone" dataKey={line.key} name={line.name} stroke={line.color} strokeWidth={2} dot={false} connectNulls={false} />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

export function AreaTrend({ data, height = 260 }: { data: { label: string; energy_kwh: number; revenue_inr: number }[]; height?: number }) {
  return (
    <div style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
          <CartesianGrid stroke="#e2e6e1" vertical={false} />
          <XAxis dataKey="label" tick={{ fontSize: 11, fill: "#5d6b64" }} axisLine={false} tickLine={false} minTickGap={24} />
          <YAxis yAxisId="energy" tick={{ fontSize: 11, fill: "#5d6b64" }} axisLine={false} tickLine={false} width={48} />
          <YAxis yAxisId="revenue" orientation="right" tick={{ fontSize: 11, fill: "#5d6b64" }} axisLine={false} tickLine={false} width={56} />
          <Tooltip content={<Tip />} />
          <Area yAxisId="energy" type="monotone" dataKey="energy_kwh" name="Energy" stroke="#1c6b45" fill="#1c6b45" fillOpacity={0.15} />
          <Area yAxisId="revenue" type="monotone" dataKey="revenue_inr" name="Revenue" stroke="#1e4f86" fill="#1e4f86" fillOpacity={0.08} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}

export function BarSeries({ data, dataKey, name, color = "#1c6b45", height = 240 }: { data: object[]; dataKey: string; name: string; color?: string; height?: number }) {
  return (
    <div style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
          <CartesianGrid stroke="#e2e6e1" vertical={false} />
          <XAxis dataKey="label" tick={{ fontSize: 11, fill: "#5d6b64" }} axisLine={false} tickLine={false} />
          <YAxis tick={{ fontSize: 11, fill: "#5d6b64" }} axisLine={false} tickLine={false} width={48} />
          <Tooltip content={<Tip />} />
          <Bar dataKey={dataKey} name={name} fill={color} radius={[3, 3, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

export function dailyLabels<T extends { date: string }>(rows: T[]) {
  return rows.map((row) => ({ ...row, label: dayLabel(row.date) }));
}
