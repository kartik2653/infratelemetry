"use client";

import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { apiGet } from "./api";
import type { Envelope, Health } from "./types";

const ClockContext = createContext<Health | null>(null);

export function ClockProvider({ children }: { children: React.ReactNode }) {
  const [health, setHealth] = useState<Health | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let cancel = false;
    apiGet<Health>("/api/v1/health")
      .then((res) => {
        if (!cancel) setHealth(res.data);
      })
      .catch((err: Error) => {
        if (!cancel) setError(err.message);
      });
    return () => {
      cancel = true;
    };
  }, [attempt]);

  if (error) {
    return (
      <div className="grid min-h-screen place-items-center bg-paper p-6">
        <div className="max-w-md rounded-xl border border-line bg-white p-6 shadow-card">
          <p className="text-sm font-medium text-rose">Cannot reach the charging API</p>
          <p className="mt-2 text-sm text-ink-soft">{error}</p>
          <p className="mt-2 text-sm text-ink-soft">Start the Go backend, then retry. The default address is http://localhost:8080.</p>
          <button type="button" className="mt-4 rounded-md bg-pine px-3 py-2 text-sm text-white" onClick={() => { setError(null); setAttempt((n) => n + 1); }}>
            Retry
          </button>
        </div>
      </div>
    );
  }
  if (!health) {
    return <div className="grid min-h-screen place-items-center text-sm text-ink-soft">Loading simulated network…</div>;
  }
  return <ClockContext.Provider value={health}>{children}</ClockContext.Provider>;
}

export function useClock(): Health {
  const value = useContext(ClockContext);
  if (!value) throw new Error("Clock is not ready");
  return value;
}

export type RangeKey = "7d" | "30d" | "90d";

export function useFilters() {
  const clock = useClock();
  const sp = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const raw = sp.get("range");
  const range: RangeKey = raw === "7d" || raw === "90d" ? raw : "30d";
  const chargerType = sp.get("charger_type") === "AC" || sp.get("charger_type") === "DC" ? sp.get("charger_type")! : "";
  const stationId = sp.get("station_id") || "";

  const { from, to } = useMemo(() => {
    const asOf = new Date(clock.as_of).getTime();
    const dataStart = new Date(clock.data_start).getTime();
    const days = range === "7d" ? 7 : range === "30d" ? 30 : 0;
    let fromMs = days === 0 ? dataStart : asOf - days * 24 * 60 * 60 * 1000;
    if (fromMs < dataStart) fromMs = dataStart;
    return { from: new Date(fromMs).toISOString(), to: new Date(asOf).toISOString() };
  }, [clock.as_of, clock.data_start, range]);

  function update(next: Record<string, string | null>) {
    const params = new URLSearchParams(sp.toString());
    for (const [key, value] of Object.entries(next)) {
      if (value === null || value === "") params.delete(key);
      else params.set(key, value);
    }
    const q = params.toString();
    router.replace(q ? `${pathname}?${q}` : pathname, { scroll: false });
  }

  function href(path: string, extra?: Record<string, string | null>) {
    const params = new URLSearchParams();
    params.set("range", range);
    if (chargerType) params.set("charger_type", chargerType);
    if (extra) {
      for (const [key, value] of Object.entries(extra)) {
        if (value) params.set(key, value);
        else params.delete(key);
      }
    }
    return `${path}?${params.toString()}`;
  }

  return { clock, range, chargerType, stationId, from, to, update, href, sp };
}

export type { Envelope };
