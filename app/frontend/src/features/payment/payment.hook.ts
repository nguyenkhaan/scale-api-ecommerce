import { useCallback, useEffect, useRef, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'

import type { Order } from '@/features/bill/bill.type'

import { describeSessionFailure, shouldRequestSession, type SessionFailure } from './payment.logic'
import { paymentClient } from './payment.service'
import type { PaymentSession } from './payment.type'

export type PaymentSessionState = {
  session: PaymentSession | null
  failure: SessionFailure | null
  /** Retries after a transient (non "not payable") failure. */
  retry: () => void
}

/**
 * Requests the payment session once the order is known to be payable. The endpoint is idempotent
 * (a repeat returns the same session), but the page still calls it only once per mount.
 * On ORDER_NOT_PAYABLE it asks the caller to refetch the official order. `onNotPayable` must be stable.
 */
export function usePaymentSession(orderId: string, order: Order | undefined, onNotPayable: () => void): PaymentSessionState {
  const requested = useRef(false)
  const [attempt, setAttempt] = useState(0)
  const [result, setResult] = useState<Pick<PaymentSessionState, 'session' | 'failure'>>({ session: null, failure: null })

  useEffect(() => {
    if (!shouldRequestSession(order, requested.current)) return
    requested.current = true
    paymentClient
      .createPaymentSession(orderId)
      .then((session) => setResult({ session, failure: null }))
      .catch((err: unknown) => {
        const failure = describeSessionFailure(err)
        setResult({ session: null, failure })
        if (failure.kind === 'not-payable') onNotPayable()
      })
  }, [order, orderId, onNotPayable, attempt])

  const retry = useCallback(() => {
    requested.current = false
    setResult({ session: null, failure: null })
    setAttempt((value) => value + 1)
  }, [])

  return { ...result, retry }
}

export const simulatorQueryKeys = {
  sessions: ['payment-simulator', 'sessions'] as const,
}

export function useSimulatorSessions() {
  return useQuery({
    queryKey: simulatorQueryKeys.sessions,
    queryFn: () => paymentClient.listSimulatorSessions(),
    retry: false,
  })
}

export function useScheduleSimulatorEvent() {
  return useMutation({
    mutationKey: ['payment-simulator', 'schedule'],
    mutationFn: paymentClient.scheduleSimulatorEvent,
  })
}
