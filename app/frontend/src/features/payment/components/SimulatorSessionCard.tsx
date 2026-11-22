import { useState } from 'react'

import { Alert, StatusBadge, Surface } from '@/components/common'
import { formatDateTime, formatVND } from '@/lib/format'

import { buildSimulatorRequest, defaultSimulatorForm, describeSimulatorError, type SimulatorFormErrors, type SimulatorFormValues } from '../simulator.logic'
import type { ScheduledPaymentEvent, SimulatorEventRequest, SimulatorSession } from '../payment.type'
import { SimulatorEventForm } from './SimulatorEventForm'

type Props = {
  session: SimulatorSession
  schedule: (body: SimulatorEventRequest) => Promise<ScheduledPaymentEvent>
}

export function SimulatorSessionCard({ session, schedule }: Props) {
  const [values, setValues] = useState<SimulatorFormValues>(defaultSimulatorForm)
  const [errors, setErrors] = useState<SimulatorFormErrors>({})
  const [submitting, setSubmitting] = useState(false)
  const [scheduled, setScheduled] = useState<ScheduledPaymentEvent | null>(null)
  const [failure, setFailure] = useState<string | null>(null)

  const submit = async () => {
    const built = buildSimulatorRequest(session.id, values)
    if (!built.ok) {
      setErrors(built.errors)
      return
    }
    setErrors({})
    setFailure(null)
    setScheduled(null)
    setSubmitting(true)
    try {
      setScheduled(await schedule(built.body))
    } catch (err) {
      setFailure(describeSimulatorError(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <li>
      <Surface as="article" className="space-y-4 p-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="font-semibold text-[var(--ss-color-text)]">Đơn {session.order_number}</h2>
            <p className="text-sm text-[var(--ss-color-muted)]">
              {formatVND(session.amount)} · Tham chiếu {session.simulator_reference}
            </p>
            <p className="text-sm text-[var(--ss-color-muted)]">Hạn thanh toán: {formatDateTime(session.expires_at)}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <StatusBadge domain="payment" value={session.status} audience="admin" />
            <StatusBadge domain="order" value={session.order_status} audience="admin" />
          </div>
        </div>
        <SimulatorEventForm
          idPrefix={`sim-${session.id}`}
          values={values}
          errors={errors}
          submitting={submitting}
          onChange={(patch) => setValues((current) => ({ ...current, ...patch }))}
          onSubmit={() => void submit()}
        />
        {scheduled && (
          <Alert tone="success">
            Sự kiện đã lên lịch. Mã sự kiện: <span className="break-all font-mono">{scheduled.event_id}</span> · Gửi lúc {formatDateTime(scheduled.scheduled_at)}
          </Alert>
        )}
        {failure && <Alert tone="danger">{failure}</Alert>}
      </Surface>
    </li>
  )
}
