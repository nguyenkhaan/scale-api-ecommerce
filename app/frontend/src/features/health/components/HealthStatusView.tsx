import type { HealthState } from '../health.hook'

export function HealthStatusView({ state, onRetry }: { state: HealthState; onRetry: () => void }) {
  if (state.type === 'loading') return <div aria-busy="true" role="status" className="border-l-4 border-slate-400 bg-white p-5"><p className="font-semibold text-slate-950">Checking backend readiness</p><p className="mt-1 text-sm text-slate-600">Waiting for PostgreSQL and the API to respond.</p></div>
  if (state.type === 'success') return <div role="status" className="border-l-4 border-emerald-600 bg-white p-5"><p className="font-semibold text-emerald-800">Backend connected</p><p className="mt-1 text-sm text-slate-600">PostgreSQL readiness check passed.</p><p className="mt-4 break-all font-mono text-xs text-slate-500">Request ID: {state.requestID}</p></div>
  return <div role="alert" className="border-l-4 border-rose-600 bg-white p-5"><p className="font-semibold text-rose-800">Backend unavailable</p><p className="mt-1 text-sm text-slate-600">Start the API and confirm the database connection, then retry.</p><button type="button" onClick={onRetry} className="mt-5 min-h-11 border border-slate-900 bg-slate-900 px-4 py-2 text-sm font-semibold text-white outline-offset-4 hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-amber-500">Try again</button></div>
}
