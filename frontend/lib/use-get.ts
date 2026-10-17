"use client";

import { useEffect, useState } from "react";
import { apiGet } from "./api";
import type { Envelope } from "./types";

export function useGet<T>(path: string | null) {
  const [data, setData] = useState<Envelope<T> | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (!path) return;
    let cancel = false;
    setLoading(true);
    setError(null);
    apiGet<T>(path)
      .then((res) => {
        if (!cancel) setData(res);
      })
      .catch((err: Error) => {
        if (!cancel) setError(err.message);
      })
      .finally(() => {
        if (!cancel) setLoading(false);
      });
    return () => {
      cancel = true;
    };
  }, [path, tick]);

  return { data, error, loading: path === null || loading, reload: () => setTick((n) => n + 1) };
}
