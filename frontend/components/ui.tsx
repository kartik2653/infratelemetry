import Link from "next/link";

export function Card({ children, className = "" }: { children: React.ReactNode; className?: string }) {
  return <section className={`rounded-xl border border-line bg-white p-4 shadow-card ${className}`}>{children}</section>;
}

export function SectionTitle({ title, detail }: { title: string; detail?: string }) {
  return (
    <div className="mb-3">
      <h2 className="text-sm font-semibold text-ink">{title}</h2>
      {detail && <p className="mt-1 text-xs leading-5 text-ink-soft">{detail}</p>}
    </div>
  );
}

export function Kpi({
  label,
  value,
  detail,
  href,
}: {
  label: string;
  value: string;
  detail?: string;
  href?: string;
}) {
  const body = (
    <div className="flex h-full flex-col justify-between rounded-xl border border-line bg-white px-4 py-3 shadow-card">
      <p className="text-[11px] font-medium uppercase tracking-wide text-ink-faint">{label}</p>
      <p className="mt-2 text-2xl font-semibold tabular tracking-tight text-ink">{value}</p>
      {detail && <p className="mt-1 text-xs text-ink-soft">{detail}</p>}
    </div>
  );
  if (!href) return body;
  return (
    <Link href={href} className="block h-full rounded-xl focus-visible:outline focus-visible:outline-2 focus-visible:outline-pine">
      {body}
    </Link>
  );
}

const statusTone: Record<string, string> = {
  available: "bg-pine-soft text-pine-dark",
  occupied: "bg-blue-50 text-sea",
  unavailable: "bg-amber-50 text-clay",
  faulted: "bg-rose-50 text-rose",
  completed: "bg-pine-soft text-pine-dark",
  failed: "bg-rose-50 text-rose",
  interrupted: "bg-amber-50 text-clay",
  in_progress: "bg-blue-50 text-sea",
  critical: "bg-rose-50 text-rose",
  warning: "bg-amber-50 text-clay",
  info: "bg-blue-50 text-sea",
  open: "bg-rose-50 text-rose",
  acknowledged: "bg-amber-50 text-clay",
  resolved: "bg-pine-soft text-pine-dark",
  high: "bg-rose-50 text-rose",
  medium: "bg-amber-50 text-clay",
  low: "bg-blue-50 text-sea",
};

export function Badge({ value }: { value: string }) {
  const tone = statusTone[value] || "bg-paper text-ink-soft";
  return <span className={`inline-flex rounded-full px-2 py-0.5 text-[11px] font-medium capitalize ${tone}`}>{value.replaceAll("_", " ")}</span>;
}

export function Definitions({ items }: { items?: Record<string, string> }) {
  if (!items || Object.keys(items).length === 0) return null;
  return (
    <details className="rounded-xl border border-line bg-white p-4 shadow-card">
      <summary className="cursor-pointer text-sm font-medium">How these numbers are calculated</summary>
      <div className="mt-3 space-y-3">
        {Object.entries(items).map(([key, text]) => (
          <p key={key} className="text-sm leading-6 text-ink-soft">
            <span className="font-medium capitalize text-ink">{key.replaceAll("_", " ")}. </span>
            {text}
          </p>
        ))}
      </div>
    </details>
  );
}

export function StatusMix({ counts }: { counts: Record<string, number> }) {
  const order = ["available", "occupied", "unavailable", "faulted"];
  const total = order.reduce((sum, key) => sum + (counts[key] || 0), 0) || 1;
  const color: Record<string, string> = {
    available: "bg-pine",
    occupied: "bg-sea",
    unavailable: "bg-clay",
    faulted: "bg-rose",
  };
  return (
    <div>
      <div className="flex h-2 overflow-hidden rounded-full bg-paper">
        {order.map((key) => (
          <div key={key} className={color[key]} style={{ width: `${((counts[key] || 0) / total) * 100}%` }} />
        ))}
      </div>
      <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-ink-soft">
        {order.map((key) => (
          <span key={key} className="capitalize">
            <span className={`mr-1 inline-block h-2 w-2 rounded-full ${color[key]}`} />
            {key} {counts[key] || 0}
          </span>
        ))}
      </div>
    </div>
  );
}
