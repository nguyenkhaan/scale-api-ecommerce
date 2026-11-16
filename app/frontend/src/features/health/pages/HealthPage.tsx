import { HealthStatusView } from '../components/HealthStatusView'
import { useHealthReadiness } from '../health.hook'

export function HealthPage() {
  const { state, retry } = useHealthReadiness()
  return (
    <section aria-labelledby="health-title" className="mx-auto w-full max-w-3xl">
      <p className="text-sm font-semibold uppercase tracking-widest text-amber-700">Environment check</p>
      <h1 id="health-title" className="mt-2 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">Backend connection</h1>
      <p className="mt-3 max-w-2xl text-base leading-7 text-slate-600">This page calls the readiness endpoint through the configured API base URL or the Vite development proxy.</p>
      <div className="mt-8"><HealthStatusView state={state} onRetry={retry} /></div>
    </section>
  )
}
