import Link from "next/link";
import { Badge } from "./ui";
import type { Insight } from "@/lib/types";

export function InsightCard({ insight, href }: { insight: Insight; href?: string }) {
  return (
    <article className="rounded-xl border border-line bg-white p-4 shadow-card">
      <div className="flex flex-wrap items-center gap-2">
        <Badge value={insight.priority} />
        <span className="text-xs capitalize text-ink-faint">{insight.category.replaceAll("_", " ")}</span>
      </div>
      <h3 className="mt-2 text-sm font-semibold leading-6">{insight.title}</h3>
      <p className="mt-2 text-sm leading-6 text-ink-soft">{insight.evidence}</p>
      <p className="mt-2 text-sm leading-6"><span className="font-medium">Why it matters. </span>{insight.why_it_matters}</p>
      <p className="mt-2 text-sm leading-6"><span className="font-medium">Next step. </span>{insight.recommended_action}</p>
      {insight.estimated_impact && <p className="mt-2 text-sm leading-6 text-ink-soft">{insight.estimated_impact}</p>}
      {insight.station_ids.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-2">
          {insight.station_ids.map((id, index) => (
            <Link key={id} href={href || `/stations/${id}`} className="text-sm font-medium text-pine underline-offset-2 hover:underline">
              {insight.station_names[index] || id}
            </Link>
          ))}
        </div>
      )}
    </article>
  );
}
