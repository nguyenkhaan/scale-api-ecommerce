import { API_BASE_URL } from '@/api/base-url'
import { createJsonHttp, type JsonHttpOptions } from '@/lib/json-http'

import type { PaymentSession, ScheduledPaymentEvent, SimulatorEventRequest, SimulatorSessionPage } from './payment.type'

export const SIMULATOR_SESSION_LIMIT = 50

/** Payment API client; the transport can be stubbed through `options.fetcher`. */
export function createPaymentClient(options: JsonHttpOptions) {
  const http = createJsonHttp(options)

  return {
    /** 201 when created, 200 when the order already has a session: both return the same session. */
    createPaymentSession: (orderId: string) =>
      http.json<PaymentSession>(`/orders/${encodeURIComponent(orderId)}/payment-session`, { method: 'POST' }),
    listSimulatorSessions: (limit = SIMULATOR_SESSION_LIMIT) =>
      http.json<SimulatorSessionPage>(`/payment-simulator/sessions?status=pending&limit=${limit}`, { method: 'GET' }),
    scheduleSimulatorEvent: (body: SimulatorEventRequest) =>
      http.json<ScheduledPaymentEvent>('/payment-simulator/events', { method: 'POST', body }),
  }
}

export const paymentClient = createPaymentClient({ baseURL: API_BASE_URL })
