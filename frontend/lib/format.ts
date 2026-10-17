export function inr(value: number | null | undefined, digits = 0): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "—";
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: "INR",
    maximumFractionDigits: digits,
    minimumFractionDigits: digits,
  }).format(value);
}

export function num(value: number | null | undefined, digits = 0): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "—";
  return new Intl.NumberFormat("en-IN", { maximumFractionDigits: digits, minimumFractionDigits: digits }).format(value);
}

export function pct(value: number | null | undefined): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "—";
  return `${value.toFixed(1)}%`;
}

export function kwh(value: number | null | undefined): string {
  return value === null || value === undefined ? "—" : `${num(value)} kWh`;
}

export function kw(value: number | null | undefined): string {
  return value === null || value === undefined ? "—" : `${num(value, value % 1 === 0 ? 0 : 1)} kW`;
}

export function hours(value: number | null | undefined): string {
  if (value === null || value === undefined) return "—";
  if (value < 48) return `${num(value, 1)} h`;
  return `${num(value)} h`;
}

export function minutes(value: number | null | undefined): string {
  if (value === null || value === undefined) return "—";
  const total = Math.round(value);
  if (total < 60) return `${total} min`;
  const h = Math.floor(total / 60);
  const m = total % 60;
  return m === 0 ? `${h}h` : `${h}h ${m}m`;
}

export function when(value: string | null | undefined): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return new Intl.DateTimeFormat("en-IN", {
    timeZone: "Asia/Kolkata",
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(d);
}

export function dayLabel(value: string): string {
  const d = new Date(`${value}T00:00:00+05:30`);
  return new Intl.DateTimeFormat("en-IN", { timeZone: "Asia/Kolkata", day: "2-digit", month: "short" }).format(d);
}

export function clockLabel(value: string): string {
  const d = new Date(value);
  return new Intl.DateTimeFormat("en-IN", {
    timeZone: "Asia/Kolkata",
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(d);
}

export function titleCase(value: string): string {
  return value.replaceAll("_", " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
