import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'

import { cartQueryKeys } from '@/features/cart/cart.hook'
import type { CheckoutPreview } from '@/features/cart/cart.type'
import type { LineError } from '@/lib/api-error'

import { checkoutClient } from './checkout.service'
import {
  buildPlaceOrderRequest,
  classifyPlaceOrderFailure,
  clearIntent,
  interpretAttempt,
  resolveIntent,
  validateRecipient,
} from './checkout.logic'
import type { CheckoutOutcome, PlaceOrderRequest, RecipientErrors, RecipientInput } from './checkout.type'
import type { PlaceOrderState } from './components/PlaceOrderSection'

type Args = {
  preview: CheckoutPreview | null
  acknowledged: boolean
  onPriceChanged: () => void
  onCartChanged: () => void
  onRejected?: (errors: LineError[]) => void
}

type ActiveAttempt = { key: string; request: PlaceOrderRequest }

const fallbackStorage = new Map<string, string>()

export function usePlaceOrder({ preview, acknowledged, onPriceChanged, onCartChanged, onRejected }: Args): PlaceOrderState {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [recipient, setRecipient] = useState<RecipientInput>({ name: '', phone: '', address: '' })
  const [errors, setErrors] = useState<RecipientErrors>({})
  const [submitting, setSubmitting] = useState(false)
  const [outcome, setOutcome] = useState<CheckoutOutcome | null>(null)
  const [retryCountdown, setRetryCountdown] = useState<number | null>(null)
  const activeAttempt = useRef<ActiveAttempt | null>(null)
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const storage = useCallback((): Pick<Storage, 'getItem' | 'setItem' | 'removeItem'> => {
    try {
      if (typeof sessionStorage !== 'undefined') return sessionStorage
    } catch {
      // Private browsing can expose sessionStorage but reject access; use the in-memory fallback.
    }
    return {
      getItem: (key) => fallbackStorage.get(key) ?? null,
      setItem: (key, value) => void fallbackStorage.set(key, value),
      removeItem: (key) => void fallbackStorage.delete(key),
    }
  }, [])

  const finish = useCallback(
    async (next: CheckoutOutcome) => {
      setOutcome(next)
      setRetryCountdown(next.kind === 'overload' ? next.retryAfterSeconds : null)
      if (next.kind === 'created') {
        clearIntent(storage())
        await queryClient.invalidateQueries({ queryKey: cartQueryKeys.current })
        await queryClient.invalidateQueries({ queryKey: ['orders'] })
        navigate(`/orders/${next.order.id}`)
      } else if (next.kind === 'rejected') {
        clearIntent(storage())
        onRejected?.(next.lineErrors)
      } else if (next.kind === 'price_changed') {
        clearIntent(storage())
        onPriceChanged()
      } else if (next.kind === 'cart_changed') {
        clearIntent(storage())
        onCartChanged()
      } else if (next.kind === 'intent_conflict') {
        clearIntent(storage())
      }
    },
    [navigate, onCartChanged, onPriceChanged, onRejected, queryClient, storage],
  )

  const checkResult = useCallback(async () => {
    const attempt = activeAttempt.current
    if (!attempt) return
    setSubmitting(true)
    try {
      const result = await checkoutClient.getAttempt(attempt.key)
      await finish(interpretAttempt(result))
    } catch (error) {
      const next = classifyPlaceOrderFailure(error)
      setOutcome(next.kind === 'unknown' ? { kind: 'unknown', requestID: next.requestID } : next)
    } finally {
      setSubmitting(false)
    }
  }, [finish])

  const send = useCallback(
    async (attempt: ActiveAttempt) => {
      activeAttempt.current = attempt
      setSubmitting(true)
      setOutcome(null)
      try {
        const result = await checkoutClient.placeOrder(attempt.key, attempt.request)
        await finish({ kind: 'created', order: result.order })
      } catch (error) {
        await finish(classifyPlaceOrderFailure(error))
      } finally {
        setSubmitting(false)
      }
    },
    [finish],
  )

  const submit = useCallback(async () => {
    if (preview === null || !acknowledged || submitting) return
    const nextErrors = validateRecipient(recipient)
    setErrors(nextErrors)
    if (Object.keys(nextErrors).length > 0) return

    const request = buildPlaceOrderRequest(preview, recipient)
    const intent = resolveIntent(storage(), request)
    await send({ key: intent.key, request })
  }, [acknowledged, preview, recipient, send, storage, submitting])

  const retry = useCallback(async () => {
    if (outcome?.kind === 'price_changed') {
      setOutcome(null)
      onPriceChanged()
      return
    }
    if (outcome?.kind === 'cart_changed') {
      setOutcome(null)
      onCartChanged()
      return
    }
    if (outcome?.kind === 'overload' && (retryCountdown ?? 0) > 0) return
    if (activeAttempt.current) await send(activeAttempt.current)
  }, [onCartChanged, onPriceChanged, outcome, retryCountdown, send])

  useEffect(() => {
    if (outcome?.kind === 'overload' && retryCountdown !== null && retryCountdown > 0) {
      const id = setTimeout(() => setRetryCountdown((current) => (current === null ? null : Math.max(0, current - 1))), 1000)
      return () => clearTimeout(id)
    }
    return undefined
  }, [outcome?.kind, retryCountdown])

  useEffect(() => {
    if (outcome?.kind !== 'unknown' || activeAttempt.current === null) return undefined
    timer.current = setTimeout(() => void checkResult(), 5000)
    return () => {
      if (timer.current !== null) clearTimeout(timer.current)
    }
  }, [checkResult, outcome?.kind])

  return { recipient, setRecipient, errors, submitting, outcome, retryCountdown, submit, retry, checkResult }
}
