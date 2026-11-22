import type { FormEvent } from 'react'

import { Button, FormField } from '@/components/common'

import type { SimulatorFormErrors, SimulatorFormValues } from '../simulator.logic'

const inputClass =
  'block min-h-11 w-full border border-[var(--ss-color-input-border)] bg-white px-3 py-2 text-base text-[var(--ss-color-text)]'

type Props = {
  idPrefix: string
  values: SimulatorFormValues
  errors: SimulatorFormErrors
  submitting: boolean
  onChange: (patch: Partial<SimulatorFormValues>) => void
  onSubmit: () => void
}

export function SimulatorEventForm({ idPrefix, values, errors, submitting, onChange, onSubmit }: Props) {
  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    onSubmit()
  }
  const id = (name: string) => `${idPrefix}-${name}`

  return (
    <form onSubmit={handleSubmit} noValidate className="space-y-4">
      <FormField id={id('event-type')} label="Kết quả" errorMessages={[]}>
        {(control) => (
          <select
            {...control}
            id={id('event-type')}
            value={values.eventType}
            onChange={(event) => onChange({ eventType: event.target.value === 'failed' ? 'failed' : 'succeeded' })}
            className={inputClass}
          >
            <option value="succeeded">Thành công</option>
            <option value="failed">Thất bại</option>
          </select>
        )}
      </FormField>

      <div className="grid gap-4 sm:grid-cols-2">
        <FormField id={id('delay')} label="Độ trễ (ms)" helper="Để trống = gửi ngay." errorMessages={errors.delayMs ? [errors.delayMs] : []}>
          {(control) => (
            <input {...control} id={id('delay')} inputMode="numeric" value={values.delayMs} onChange={(event) => onChange({ delayMs: event.target.value })} className={inputClass} />
          )}
        </FormField>
        <FormField id={id('repeat')} label="Số lần gửi lặp" helper="Để trống = 1 lần." errorMessages={errors.repeatCount ? [errors.repeatCount] : []}>
          {(control) => (
            <input {...control} id={id('repeat')} inputMode="numeric" value={values.repeatCount} onChange={(event) => onChange({ repeatCount: event.target.value })} className={inputClass} />
          )}
        </FormField>
      </div>

      <label className="flex min-h-11 items-center gap-2 text-sm text-[var(--ss-color-text)]">
        <input type="checkbox" className="size-4" checked={values.dropDelivery} onChange={(event) => onChange({ dropDelivery: event.target.checked })} />
        Không gửi (mô phỏng mất callback)
      </label>
      <label className="flex min-h-11 items-center gap-2 text-sm text-[var(--ss-color-text)]">
        <input type="checkbox" className="size-4" checked={values.invalidSignature} onChange={(event) => onChange({ invalidSignature: event.target.checked })} />
        Chữ ký sai
      </label>

      <details className="border border-[var(--ss-color-border)] p-3">
        <summary className="min-h-11 cursor-pointer py-2 text-sm font-semibold text-[var(--ss-color-text)]">Nâng cao</summary>
        <div className="mt-2 grid gap-4 sm:grid-cols-2">
          <FormField id={id('amount')} label="Ghi đè số tiền" helper="Chỉ để thử trường hợp sai số tiền." errorMessages={errors.amountOverride ? [errors.amountOverride] : []}>
            {(control) => (
              <input {...control} id={id('amount')} inputMode="numeric" value={values.amountOverride} onChange={(event) => onChange({ amountOverride: event.target.value })} className={inputClass} />
            )}
          </FormField>
          <FormField id={id('currency')} label="Ghi đè tiền tệ" helper="Chỉ để thử trường hợp sai tiền tệ." errorMessages={[]}>
            {(control) => (
              <input {...control} id={id('currency')} value={values.currencyOverride} onChange={(event) => onChange({ currencyOverride: event.target.value })} className={inputClass} />
            )}
          </FormField>
        </div>
      </details>

      <Button type="submit" loading={submitting}>Gửi sự kiện</Button>
    </form>
  )
}
