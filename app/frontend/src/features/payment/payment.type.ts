import type { OrderPaymentStatus, OrderStatus } from '@/features/bill/bill.type'

export type PaymentSession = {
  id: string
  order_id: string
  status: OrderPaymentStatus
  amount: number
  currency: 'VND'
  simulator_reference: string
}

/** Admin/test only: a payment session still awaiting a simulated gateway event. */
export type SimulatorSession = {
  id: string
  order_id: string
  order_number: string
  amount: number
  currency: string
  status: OrderPaymentStatus
  simulator_reference: string
  order_status: OrderStatus
  expires_at: string
}

export type SimulatorSessionPage = {
  items: SimulatorSession[]
}

export type SimulatorEventRequest = {
  payment_session_id: string
  event_type: 'succeeded' | 'failed'
  delay_ms: number
  repeat_count: number
  drop_delivery: boolean
  signature_mode: 'valid' | 'invalid'
  amount_override?: number
  currency_override?: string
}

export type ScheduledPaymentEvent = {
  event_id: string
  scheduled_at: string
}
