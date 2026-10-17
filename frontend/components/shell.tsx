"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { Bell, LayoutDashboard, Lightbulb, LineChart, MapPinned } from "lucide-react";
import { apiGet } from "@/lib/api";
import { ClockProvider, useClock } from "@/lib/filters";
import { clockLabel } from "@/lib/format";
import type { Alert } from "@/lib/types";

const links = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/stations", label: "Stations", icon: MapPinned },
  { href: "/analytics", label: "Analytics", icon: LineChart },
  { href: "/alerts", label: "Alerts", icon: Bell },
  { href: "/insights", label: "Insights", icon: Lightbulb },
];

function Nav({ openCount }: { openCount: number }) {
  const pathname = usePathname();
  return (
    <nav className="grid grid-cols-3 gap-1 px-2 py-2 md:flex md:flex-col md:px-3">
      {links.map((item) => {
        const active = item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
        const Icon = item.icon;
        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={active ? "page" : undefined}
            className={`flex items-center gap-2 rounded-lg px-2 py-2 text-sm md:px-3 ${active ? "bg-white/10 text-white" : "text-white/70 hover:bg-white/5 hover:text-white"}`}
          >
            <Icon size={16} />
            <span>{item.label}</span>
            {item.href === "/alerts" && openCount > 0 && (
              <span className="ml-auto rounded-full bg-rose px-1.5 text-[11px] text-white">{openCount}</span>
            )}
          </Link>
        );
      })}
    </nav>
  );
}

function Frame({ children }: { children: React.ReactNode }) {
  const clock = useClock();
  const pathname = usePathname();
  const [openCount, setOpenCount] = useState(0);
  useEffect(() => {
    const load = () => {
      apiGet<{ alerts: Alert[] }>("/api/v1/alerts?status=open")
        .then((res) => setOpenCount(res.data.alerts.length))
        .catch(() => setOpenCount(0));
    };
    load();
    window.addEventListener("alerts-changed", load);
    return () => window.removeEventListener("alerts-changed", load);
  }, [pathname]);
  return (
    <div className="min-h-screen md:grid md:grid-cols-[232px_1fr]">
      <aside className="bg-side text-white md:min-h-screen">
        <div className="flex items-center gap-3 px-4 py-4">
          <div className="grid h-9 w-9 place-items-center rounded-lg bg-pine text-sm font-semibold">IN</div>
          <div>
            <p className="text-sm font-semibold">InfraTelemetry</p>
            <p className="text-[11px] text-white/60">{clock.operator_name}</p>
          </div>
        </div>
        <Nav openCount={openCount} />
      </aside>
      <div className="min-w-0">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line bg-white px-4 py-3 md:px-6">
          <div>
            <p className="text-sm font-medium">{clock.network_name}</p>
            <p className="text-xs text-ink-soft">Demo clock {clockLabel(clock.as_of)} IST · seed {clock.seed}</p>
          </div>
          <span className="rounded-full border border-clay/30 bg-amber-50 px-3 py-1 text-xs font-medium text-clay">Simulated data</span>
        </div>
        <div className="border-b border-amber-200 bg-amber-50 px-4 py-2 text-xs leading-5 text-clay md:px-6">{clock.disclaimer}</div>
        <main className="px-4 py-5 md:px-6">{children}</main>
      </div>
    </div>
  );
}

export function Shell({ children }: { children: React.ReactNode }) {
  return (
    <ClockProvider>
      <Suspense fallback={<div className="p-6 text-sm text-ink-soft">Loading view…</div>}>
        <Frame>{children}</Frame>
      </Suspense>
    </ClockProvider>
  );
}
