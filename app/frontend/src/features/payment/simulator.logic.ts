import { isApiError } from '@/lib/api-error'
import { describeLoadError } from '@/lib/load-error'
import { violationMessage } from '@/lib/violation-messages'

import type { SimulatorEventRequest } from './payment.type'

/** Raw form state: numeric fields stay strings so blank means "use the server default / omit". */
export type SimulatorFormValues = {
  eventType: SimulatorEventRequest['event_type']
  delayMs: string
  repeatCount: string
  dropDelivery: boolean
  invalidSignature: boolean
  amountOverride: string
  currencyOverride: string
}

export const defaultSimulatorForm: SimulatorFormValues = {
  eventType: 'succeeded',
  delayMs: '',
  repeatCount: '',
  dropDelivery: false,
  invalidSignature: false,
  amountOverride: '',
  currencyOverride: '',
}

export type SimulatorFormErrors = Partial<Record<'delayMs' | 'repeatCount' | 'amountOverride', string>>

export type SimulatorBuildResult =
  | { ok: true; body: SimulatorEventRequest }
  | { ok: false; errors: SimulatorFormErrors }

function parseWholeNumber(raw: string, min: number): number | null {
  const text = raw.trim()
  if (!/^\d+$/.test(text)) return null
  const value = Number(text)
  return Number.isSafeInteger(value) && value >= min ? value : null
}

/** Maps the form to the request body. Blank overrides are omitted; blank delay/repeat use the defaults (0 ms, 1 delivery). */
export function buildSimulatorRequest(paymentSessionId: string, values: SimulatorFormValues): SimulatorBuildResult {
  const errors: SimulatorFormErrors = {}
  let delayMs = 0
  let repeatCount = 1
  let amountOverride: number | undefined

  if (values.delayMs.trim() !== '') {
    const parsed = parseWholeNumber(values.delayMs, 0)
    if (parsed === null) errors.delayMs = 'Độ trễ phải là số nguyên không âm (ms).'
    else delayMs = parsed
  }
  if (values.repeatCount.trim() !== '') {
    const parsed = parseWholeNumber(values.repeatCount, 1)
    if (parsed === null) errors.repeatCount = 'Số lần gửi lặp phải là số nguyên từ 1 trở lên.'
    else repeatCount = parsed
  }
  if (values.amountOverride.trim() !== '') {
    const parsed = parseWholeNumber(values.amountOverride, 1)
    if (parsed === null) errors.amountOverride = 'Số tiền ghi đè phải là số nguyên từ 1 trở lên.'
    else amountOverride = parsed
  }
  if (Object.keys(errors).length > 0) return { ok: false, errors }

  const body: SimulatorEventRequest = {
    payment_session_id: paymentSessionId,
    event_type: values.eventType,
    delay_ms: delayMs,
    repeat_count: repeatCount,
    drop_delivery: values.dropDelivery,
    signature_mode: values.invalidSignature ? 'invalid' : 'valid',
  }
  if (amountOverride !== undefined) body.amount_override = amountOverride
  const currency = values.currencyOverride.trim()
  if (currency !== '') body.currency_override = currency
  return { ok: true, body }
}

/** Vietnamese copy for a failed simulator call; 404 means the simulator is disabled on the server. */
export function describeSimulatorError(err: unknown): string {
  if (isApiError(err)) {
    if (err.status === 404) return 'Công cụ mô phỏng thanh toán đang tắt trên môi trường này.'
    if (err.status === 403) return 'Chỉ quản trị viên mới được dùng công cụ mô phỏng thanh toán.'
    if (err.status === 400 || err.status === 422) {
      const messages = err.violations.map(violationMessage)
      return messages.length > 0 ? messages.join(' ') : 'Thông số sự kiện không hợp lệ. Vui lòng kiểm tra lại.'
    }
  }
  return describeLoadError(err, 'công cụ mô phỏng')
}
