import Link from "next/link";
import { Card } from "@/components/ui";

const contents = [
  { href: "#business", label: "The business" },
  { href: "#customers", label: "Who it is for" },
  { href: "#features", label: "Features in this release" },
  { href: "#metrics", label: "Decisions the numbers support" },
  { href: "#roadmap", label: "What we are adding next" },
];

const audiences = [
  {
    title: "Charging network operators",
    text: "See how a whole network is performing, which sites earn their keep, and where field effort should go first.",
  },
  {
    title: "Infrastructure owners",
    text: "Compare utilization, uptime, energy delivered, and gross margin before spending more capital on a site.",
  },
  {
    title: "Operations managers",
    text: "Watch open faults, repeated failures, downtime, and the revenue those hours put at risk.",
  },
  {
    title: "Expansion teams",
    text: "Separate a quiet site from a site that is already busy in a short evening window and may need another bay.",
  },
];

const features = [
  {
    title: "Network overview",
    href: "/",
    text: "Network KPIs, connector status at the demo clock, energy and revenue trends, utilization and uptime, the highest-revenue and lowest-utilization stations, critical alerts, and the highest-priority recommendations.",
  },
  {
    title: "Station explorer",
    href: "/stations",
    text: "Search, sort, and page through stations. Each row shows charger type, rated power, connectors, utilization, uptime, energy, revenue, gross margin, and current status. A row opens the station.",
  },
  {
    title: "Station detail",
    href: "/stations",
    text: "Tariffs, connector status, utilization and session-start patterns, sessions, fault history, and recommendations that name that station.",
  },
  {
    title: "Analytics",
    href: "/analytics",
    text: "Separate views for utilization, revenue and margin, energy throughput, session behavior, and reliability. The 7, 30, and 90 day ranges, charger type, and station filter change the active view.",
  },
  {
    title: "Alerts",
    href: "/alerts",
    text: "Offline chargers, repeated connector faults, high downtime, utilization drops, long sessions, energy anomalies, revenue at risk, and repeated payment failures. Each alert has a severity, a station, a metric, and a next step. Acknowledge, resolve, and reopen stay in memory until the service restarts.",
  },
  {
    title: "Insights",
    href: "/insights",
    text: "Ranked recommendations from rules in the Go service, not a language model. Each card states the evidence, why it matters, the next action, and an impact figure only where the arithmetic holds.",
  },
];

const roadmap = [
  {
    title: "Live charger ingestion",
    text: "Accept sessions, status, and faults from charging networks through vendor APIs, webhooks, file import, and OCPP-compatible connections. This demo generates a fixed Bengaluru network in memory instead.",
  },
  {
    title: "More than one operator",
    text: "Keep each network’s stations, tariffs, and recommendations separate, then compare them without mixing the numbers.",
  },
  {
    title: "Stored history",
    text: "Replace the in-memory store with a database so alerts, acknowledgements, and history survive a restart. The calculations already sit behind a repository boundary so the screens do not have to change.",
  },
  {
    title: "A fuller cost picture",
    text: "Add rent, maintenance, field labour, payment fees, and demand charges beside electricity procurement. Until those exist, the product reports gross margin and says what is excluded.",
  },
  {
    title: "Accounts and roles",
    text: "Sign-in for operators, owners, and operations staff, with access limited to the networks they run.",
  },
];

export default function DocsPage() {
  return (
    <div className="mx-auto max-w-3xl space-y-5">
      <header>
        <p className="text-xs font-medium uppercase tracking-wide text-ink-faint">Product documentation</p>
        <h1 className="mt-1 text-xl font-semibold">InfraTelemetry</h1>
        <p className="mt-2 text-sm leading-6 text-ink-soft">
          Vendor-independent intelligence for EV charging networks. It turns station, session, and fault history into utilization, gross margin, reliability, and a short list of actions. This release runs on simulated data for the fictional NammaCharge Bengaluru network. It is not a live charger feed.
        </p>
      </header>

      <nav aria-label="On this page" className="flex flex-wrap gap-2">
        {contents.map((item) => (
          <a key={item.href} href={item.href} className="rounded-full border border-line bg-white px-3 py-1.5 text-sm text-ink-soft hover:text-ink">
            {item.label}
          </a>
        ))}
      </nav>

      <Card>
        <h2 id="business" className="text-sm font-semibold scroll-mt-4">The business</h2>
        <div className="mt-3 space-y-3 text-sm leading-6 text-ink-soft">
          <p>
            A public or depot charger is an expensive asset. It earns only while a vehicle is actually charging, the connector is available, and the driver tariff covers the electricity bought for that session. Operators already get basic status boards from charger vendors. Those boards rarely answer whether a site should get another connector, a tariff change, or a technician.
          </p>
          <p>
            InfraTelemetry sits above the charger vendors. The business outcome is a better use of capital and field time: raise energy sold on assets that already have demand, stop adding hardware where demand is thin, and repair the few sites that create most of the downtime.
          </p>
          <p>
            Gross margin in this release is charging revenue minus electricity procurement only. Rent, maintenance, staffing, payment fees, taxes, and demand charges are listed as excluded, so a strong margin here is not a claim of net profit.
          </p>
        </div>
      </Card>

      <section>
        <h2 id="customers" className="text-sm font-semibold scroll-mt-4">Who it is for</h2>
        <div className="mt-3 grid gap-3 sm:grid-cols-2">
          {audiences.map((item) => (
            <Card key={item.title}>
              <h3 className="text-sm font-semibold">{item.title}</h3>
              <p className="mt-2 text-sm leading-6 text-ink-soft">{item.text}</p>
            </Card>
          ))}
        </div>
      </section>

      <section className="space-y-3">
        <h2 id="features" className="text-sm font-semibold scroll-mt-4">Features in this release</h2>
        <p className="text-sm leading-6 text-ink-soft">
          Every chart and KPI is calculated by the API from the in-memory dataset. The screens do not ship with hardcoded results. Filters for the last 7, 30, and 90 days apply everywhere except the alert list, which is a startup snapshot you can acknowledge and resolve.
        </p>
        {features.map((item) => (
          <Card key={item.title}>
            <div className="flex flex-wrap items-baseline justify-between gap-2">
              <h3 className="text-sm font-semibold">{item.title}</h3>
              <Link href={item.href} className="text-sm font-medium text-pine hover:underline">Open</Link>
            </div>
            <p className="mt-2 text-sm leading-6 text-ink-soft">{item.text}</p>
          </Card>
        ))}
      </section>

      <Card>
        <h2 id="metrics" className="text-sm font-semibold scroll-mt-4">Decisions the numbers support</h2>
        <ul className="mt-3 space-y-3 text-sm leading-6 text-ink-soft">
          <li><span className="font-medium text-ink">Utilization.</span> Charging time divided by time the connector was actually available. Scheduled maintenance and unplanned faults are removed from the denominator. A charger can be “up” and still barely used.</li>
          <li><span className="font-medium text-ink">Gross margin.</span> Session energy times the driver tariff, minus the same energy times the station’s procurement rate. Peak rates apply at 08:00–11:00 and 18:00–21:00 IST.</li>
          <li><span className="font-medium text-ink">Uptime, MTTR, and MTBF.</span> Reliability of the recorded history. Open faults are left out of repair time. MTBF is omitted when a window has no faults, and it is not a forecast.</li>
          <li><span className="font-medium text-ink">Revenue at risk.</span> An estimate of what the lost connector-hours would have earned at the site’s observed rate. It is not cash already missing from a ledger.</li>
          <li><span className="font-medium text-ink">Energy and power.</span> Energy delivered and the power observed while charging, compared with the nameplate. The product does not infer grid import, transformer loading, or electrical efficiency.</li>
        </ul>
        <p className="mt-3 text-sm text-ink-soft">The same definitions are shown under “How these numbers are calculated” on the overview.</p>
      </Card>

      <section className="space-y-3">
        <h2 id="roadmap" className="text-sm font-semibold scroll-mt-4">What we are adding next</h2>
        <p className="text-sm leading-6 text-ink-soft">
          These are the next product capabilities. They are not in this prototype, and the demo does not pretend they are already measuring live chargers.
        </p>
        {roadmap.map((item) => (
          <Card key={item.title}>
            <h3 className="text-sm font-semibold">{item.title}</h3>
            <p className="mt-2 text-sm leading-6 text-ink-soft">{item.text}</p>
          </Card>
        ))}
      </section>
    </div>
  );
}
