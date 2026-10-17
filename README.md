# InfraTelemetry

Simulated EV charging infrastructure intelligence for a fictional operator, **NammaCharge Grid**, and its Bengaluru network. The prototype calculates utilization, gross margin, reliability, session behavior, alerts, and rule-based recommendations from an in-memory dataset. It does not connect to chargers, vendor APIs, OCPP, or a database.

Every screen reads the Go API. Chart series, KPI values, and recommendations are not hardcoded in the frontend.

## Demo data

Stations, sessions, faults, tariffs, and electricity costs are generated in memory when the backend starts. The random seed is fixed (`20261017` by default), so restarts reproduce the same network. The UI marks the data as simulated. Nothing here is a live charging station.

The demo clock is **17 Oct 2026, 18:30 IST** unless `DATA_AS_OF` is set. History covers about 90 days before that instant.

## Prerequisites

- Go 1.23 or newer
- Node.js 20 or newer

## Run locally

From two terminals.

```powershell
cd backend
go run ./cmd/server
```

```powershell
cd frontend
npm install
npm run dev
```

Open http://localhost:3000. The page calls `/api` on that same origin, and Next forwards it to the Go server on port 8080. No env file is required for that.

### Checks

```powershell
cd backend
go test ./...
```

```powershell
cd frontend
npm run build
```

## Configuration

The process does not load `.env` files automatically. Set variables in the shell, or rely on the defaults.

| Variable | Default | Meaning |
| --- | --- | --- |
| `ADDR` | `:8080` | Backend listen address |
| `PORT` | | Used when `ADDR` is empty. `8080` becomes `:8080` |
| `DATA_SEED` | `20261017` | Generator seed |
| `DATA_DAYS` | `90` | Calendar days of history. Allowed range: 7–180 |
| `DATA_AS_OF` | `2026-10-17T18:30:00+05:30` | RFC3339 demo clock |
| `NEXT_PUBLIC_API_BASE_URL` | empty (same site) | Browser API origin. Set this only when the API is on another host. The value is fixed when the frontend is built. |
| `BACKEND_URL` | `http://127.0.0.1:8080` | Where Next proxies `/api` when the browser uses the same site. Fixed when the frontend is built. |

Examples live in `backend/.env.example` and `frontend/.env.example`.

## Sharing on your domain

The product name in the app is **InfraTelemetry**. The browser calls `/api` on the same host as the page, so one domain can serve both.

This workspace has no hosting account and cannot register a domain or change DNS. To publish it as `infratelemetry.com` (or `www.infratelemetry.com`):

1. Build the image from the repo root: `docker build -t infratelemetry .`
2. Run it on a machine with a public address: `docker run --rm -p 3000:3000 infratelemetry`
3. At the registrar for that domain, point the name at the machine and terminate TLS there. Forward HTTPS to container port 3000.

The container starts the Go API on port 8080 inside the container and the website on port 3000. Alert acknowledgements reset when the container stops.

## What the numbers mean

Calculations live in `backend/internal/analytics`. The same definitions are returned in API payloads and shown under **How these numbers are calculated**.

**Utilization** = charging time ÷ available operating time × 100. Available time is the selected window minus scheduled maintenance and unplanned fault downtime. Plug-in time after charging stops does not count. Failed attempts with no energy do not count. Utilization is not uptime.

**Uptime** = time outside unplanned faults ÷ (window − scheduled maintenance) × 100. A window that is entirely planned maintenance has no uptime figure.

**Gross margin** = charging revenue − electricity procurement cost. Peak tariff and peak procurement apply at 08:00–11:00 and 18:00–21:00 IST. This is not net profit: rent, maintenance, staffing, payment fees, taxes, depreciation, and demand charges are excluded.

**Revenue at risk** = unplanned downtime connector-hours × (station revenue ÷ available connector-hours) in the same window. It assumes demand during the outage would have matched observed utilization. It is an estimate, not recorded lost revenue.

**MTTR** uses the full duration of resolved faults that overlap the window. Open faults are excluded.

**MTBF** = connector operating hours ÷ fault incidents overlapping the window. It is omitted when there are no faults. It describes this synthetic history and is not a forecast.

**Session KPIs** (counts, energy, revenue, cost, average duration) include sessions that **start** inside the window. Utilization and uptime use the minutes that fall inside each bucket. Hourly energy splits a session across the hours it was actually charging.

**Connector status** (available, occupied, unavailable, faulted) is the state at the demo clock. It does not replay the selected dates. Hardware faults show as faulted. Communication and payment faults show as unavailable.

**Observed power** is session energy ÷ charging hours, plus the peak reading stored on the session. Hourly telemetry samples are derived from those sessions. The product does not infer grid import, transformer loading, or electrical efficiency.

## Screens

1. **Overview** — network KPIs, status mix, energy/revenue and utilization/uptime trends, top and bottom stations, critical alerts, highest-priority recommendations.
2. **Stations** — search, sort, charger type, date range, pagination. A row opens the station.
3. **Station detail** — connectors, tariffs, trends, sessions, faults, station recommendations.
4. **Analytics** — utilization, revenue, energy, sessions, and reliability. Date, charger type, and station filters apply to the active tab.
5. **Alerts** — filter by severity, status, category, and station. Acknowledge, resolve, and reopen. Status lasts until the backend process stops.
6. **Insights** — ranked recommendations for the selected window, with evidence, why it matters, a next step, and an impact statement only where the arithmetic is defensible.

Date presets are 7, 30, and 90 days ending at the demo clock. The 90-day preset is the full generated history.

## API

Responses use this envelope:

```json
{ "demo": true, "data": {}, "meta": { "as_of": "", "from": "", "to": "", "disclaimer": "" } }
```

Errors:

```json
{ "demo": true, "error": { "code": "INVALID_TIME_RANGE", "message": "to must be after from" } }
```

`from` is inclusive and `to` is exclusive. Both accept RFC3339. A date-only `from` is midnight IST that day. A date-only `to` is midnight IST of the next day. If both are omitted, the range is the full generated history. Ranges longer than 366 days are rejected.

| Method | Path | Filters |
| --- | --- | --- |
| GET | `/api/v1/health` | none |
| GET | `/api/v1/dashboard/summary` | `from`, `to`, `charger_type` |
| GET | `/api/v1/dashboard/trends` | `from`, `to`, `charger_type` |
| GET | `/api/v1/stations` | `from`, `to`, `charger_type`, `station_id`, `q`, `sort`, `order`, `page`, `limit` |
| GET | `/api/v1/stations/{id}` | `from`, `to`, `charger_type` |
| GET | `/api/v1/stations/{id}/sessions` | `from`, `to`, `status`, `page`, `limit` |
| GET | `/api/v1/stations/{id}/faults` | `from`, `to`, `page`, `limit` |
| GET | `/api/v1/analytics/utilization` | `from`, `to`, `station_id`, `charger_type` |
| GET | `/api/v1/analytics/revenue` | `from`, `to`, `station_id`, `charger_type` |
| GET | `/api/v1/analytics/energy` | `from`, `to`, `station_id`, `charger_type` |
| GET | `/api/v1/analytics/reliability` | `from`, `to`, `station_id`, `charger_type` |
| GET | `/api/v1/analytics/sessions` | `from`, `to`, `station_id`, `charger_type` |
| GET | `/api/v1/sessions` | `from`, `to`, `station_id`, `charger_type`, `status`, `page`, `limit` |
| GET | `/api/v1/alerts` | `severity`, `status`, `category`, `station_id` |
| PATCH | `/api/v1/alerts/{id}` | JSON body `{ "status": "open" \| "acknowledged" \| "resolved" }` |
| GET | `/api/v1/insights` | `from`, `to`, `station_id`, `charger_type` |

`charger_type` is `AC` or `DC`. Session `status` is `completed`, `failed`, `interrupted`, or `in_progress`. Station `sort` is `name`, `utilization`, `uptime`, `energy`, `revenue`, `margin`, `sessions`, or `downtime`. `order` is `asc` or `desc`. `page` starts at 1. `limit` is 1–100, default 25 (the station explorer asks for 10).

Alert `severity`: `critical`, `warning`, `info`. Alert `status`: `open`, `acknowledged`, `resolved`.

```powershell
curl.exe http://localhost:8080/api/v1/health
curl.exe "http://localhost:8080/api/v1/dashboard/summary?from=2026-09-17T18:30:00%2B05:30&to=2026-10-17T18:30:00%2B05:30"
curl.exe "http://localhost:8080/api/v1/stations?sort=utilization&order=asc&limit=5"
curl.exe -X PATCH http://localhost:8080/api/v1/alerts/ALT-OFF-FLT-001 -H "Content-Type: application/json" -d "{\"status\":\"acknowledged\"}"
```

Alert IDs come from `GET /api/v1/alerts`. The sample ID above is illustrative.

Unknown stations return 404. Invalid ranges, charger types, sorts, and alert statuses return 400.

## Architecture

```
backend/
  cmd/server                 startup, seed, alert generation, HTTP server
  internal/domain            stations, sessions, faults, formulas
  internal/seed              deterministic network generator
  internal/store             in-memory repository; alerts are the only mutable data
  internal/analytics         utilization, money, energy, reliability, session stats
  internal/insights          rule-based recommendations
  internal/alerts            rule-based operational alerts
  internal/httpapi           routing and request validation
frontend/                    Next.js app
```

Handlers stay thin. Analytics depends on a `Source` interface (`Dataset`, stations, connectors, sessions, faults, maintenance), not on HTTP or on a database driver. Replacing the memory store with a persistent repository later does not require rewriting the calculations. This prototype does not ship that repository, and it does not add authentication.

Insights and alerts are explainable rules. They do not call a language model. Alerts are built once at startup for the trailing 7, 14, and 30 days. Insights are recomputed for the requested window so the evidence matches the filter.

The synthetic network is shaped so the rules have something to explain: a low-utilization DC site, a high-utilization thin-margin depot, a connector-fault cluster, a long supply outage, weekend-heavy demand, payment failures, evening session peaks, and vehicles left plugged in after charging. The exact percentages come from the generator, not from constants in the UI.

## Tests

`go test ./...` covers:

- interval union and subtraction
- utilization and uptime when available time is zero, sessions are zero, or the window is entirely maintenance
- gross margin, including zero revenue
- MTTR and MTBF, including ongoing faults and zero failures
- missing peak-power telemetry
- charging time clipped out of fault windows
- network totals equal to the sum of stations
- date filters dropping sessions that start outside the window
- insight text using the supplied percentages, and healthy stations not being flagged as underused
- alert generation for an open fault, payment failures, and a utilization drop
- deterministic seed, session revenue matching the tariff, telemetry energy matching the session, and no overlapping sessions on a connector
- HTTP health, invalid range, unknown station, and acknowledge-in-memory

## Out of scope

No database, OCPP, OCPI, vendor credentials, hardware, billing, authentication, machine learning, or predictive-accuracy claims.
