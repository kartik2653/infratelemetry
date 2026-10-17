export function Loading({ label = "Loading simulated network data…" }: { label?: string }) {
  return (
    <div className="rounded-xl border border-dashed border-line bg-white px-4 py-10 text-center text-sm text-ink-soft">
      {label}
    </div>
  );
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="rounded-xl border border-rose/30 bg-white px-4 py-6">
      <p className="text-sm font-medium text-rose">The view could not be loaded</p>
      <p className="mt-1 text-sm text-ink-soft">{message}</p>
      {onRetry && (
        <button type="button" onClick={onRetry} className="mt-3 rounded-md border border-line px-3 py-1.5 text-sm">
          Retry
        </button>
      )}
    </div>
  );
}

export function Empty({ children }: { children: React.ReactNode }) {
  return <div className="rounded-xl border border-dashed border-line bg-white px-4 py-10 text-center text-sm text-ink-soft">{children}</div>;
}
