import { useEffect, useState } from 'react'
import { healthClient } from '../lib/health-api'

type HealthState =
  | { type: 'loading' }
  | { type: 'success'; requestID: string }
  | { type: 'error' }

type HealthStatusViewProps = {
  state: HealthState
  onRetry: () => void
}

export function HealthPage() {
  const [attempt, setAttempt] = useState(0)
  const [state, setState] = useState<HealthState>({ type: 'loading' })

  useEffect(() => {
    const controller = new AbortController()
    healthClient.getReadiness(controller.signal).then(
      (result) => setState({ type: 'success', requestID: result.requestID }),
      () => {
        if (!controller.signal.aborted) {
          setState({ type: 'error' })
        }
      },
    )
    return () => controller.abort()
  }, [attempt])

  const retry = () => {
    setState({ type: 'loading' })
    setAttempt((current) => current + 1)
  }

  return (
    <section aria-labelledby="health-title" className="mx-auto w-full max-w-3xl">
      <p className="text-sm font-semibold uppercase tracking-widest text-amber-700">
        Environment check
      </p>
      <h1 id="health-title" className="mt-2 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">
        Backend connection
      </h1>
      <p className="mt-3 max-w-2xl text-base leading-7 text-slate-600">
        This page calls the readiness endpoint through the configured API base URL or the Vite development proxy.
      </p>
      <div className="mt-8">
        <HealthStatusView state={state} onRetry={retry} />
      </div>
    </section>
  )
}

export function HealthStatusView({ state, onRetry }: HealthStatusViewProps) {
  if (state.type === 'loading') {
    return (
      <div aria-busy="true" role="status" className="border-l-4 border-slate-400 bg-white p-5">
        <p className="font-semibold text-slate-950">Checking backend readiness</p>
        <p className="mt-1 text-sm text-slate-600">Waiting for PostgreSQL and the API to respond.</p>
      </div>
    )
  }

  if (state.type === 'success') {
    return (
      <div role="status" className="border-l-4 border-emerald-600 bg-white p-5">
        <p className="font-semibold text-emerald-800">Backend connected</p>
        <p className="mt-1 text-sm text-slate-600">PostgreSQL readiness check passed.</p>
        <p className="mt-4 break-all font-mono text-xs text-slate-500">
          Request ID: {state.requestID}
        </p>
      </div>
    )
  }

  return (
    <div role="alert" className="border-l-4 border-rose-600 bg-white p-5">
      <p className="font-semibold text-rose-800">Backend unavailable</p>
      <p className="mt-1 text-sm text-slate-600">
        Start the API and confirm the database connection, then retry.
      </p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-5 min-h-11 border border-slate-900 bg-slate-900 px-4 py-2 text-sm font-semibold text-white outline-offset-4 hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-amber-500"
      >
        Try again
      </button>
    </div>
  )
}
