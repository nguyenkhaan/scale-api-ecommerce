import { useEffect, useState } from 'react'

import { healthClient } from './health.service'

export type HealthState =
  | { type: 'loading' }
  | { type: 'success'; requestID: string }
  | { type: 'error' }

export function useHealthReadiness() {
  const [attempt, setAttempt] = useState(0)
  const [state, setState] = useState<HealthState>({ type: 'loading' })

  useEffect(() => {
    const controller = new AbortController()
    healthClient.getReadiness(controller.signal).then(
      (result) => setState({ type: 'success', requestID: result.requestID }),
      () => {
        if (!controller.signal.aborted) setState({ type: 'error' })
      },
    )
    return () => controller.abort()
  }, [attempt])

  return {
    state,
    retry: () => {
      setState({ type: 'loading' })
      setAttempt((current) => current + 1)
    },
  }
}
