import type { Envelope } from "./types";

// Empty means "this site". Next proxies /api to the Go server, so one domain
// such as infratelemetry.com can serve both the UI and the API.
export const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function apiGet<T>(path: string): Promise<Envelope<T>> {
  const res = await fetch(`${API_BASE}${path}`, { cache: "no-store" });
  const body = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, body?.error?.message || `Request failed (${res.status})`);
  }
  return body as Envelope<T>;
}

export async function apiPatch<T>(path: string, payload: unknown): Promise<Envelope<T>> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const body = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, body?.error?.message || `Request failed (${res.status})`);
  }
  return body as Envelope<T>;
}

export function qs(params: Record<string, string | number | undefined | null>): string {
  const p = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") continue;
    p.set(key, String(value));
  }
  const s = p.toString();
  return s ? `?${s}` : "";
}
