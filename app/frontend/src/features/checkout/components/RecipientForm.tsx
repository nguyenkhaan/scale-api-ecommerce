import { FormField } from '@/components/common'

import type { RecipientErrors, RecipientInput } from '../checkout.type'

type Props = {
  value: RecipientInput
  errors: RecipientErrors
  disabled: boolean
  onChange: (field: keyof RecipientInput, value: string) => void
}

const inputClass =
  'mt-1 min-h-11 w-full rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] bg-white px-3 py-2 text-base text-[var(--ss-color-text)] focus:border-[var(--ss-color-primary)] focus:outline-none aria-[invalid=true]:border-[var(--ss-color-danger)] disabled:bg-[var(--ss-color-app-bg)]'

export function RecipientForm({ value, errors, disabled, onChange }: Props) {
  return (
    <div className="space-y-4" aria-label="Thông tin người nhận">
      <FormField id="checkout-recipient-name" label="Họ tên" errorMessages={errors.name ? [errors.name] : []}>
        {(control) => <input {...control} disabled={disabled} value={value.name} onChange={(event) => onChange('name', event.target.value)} className={inputClass} />}
      </FormField>
      <FormField id="checkout-recipient-phone" label="Số điện thoại" errorMessages={errors.phone ? [errors.phone] : []}>
        {(control) => <input {...control} disabled={disabled} inputMode="tel" value={value.phone} onChange={(event) => onChange('phone', event.target.value)} className={inputClass} />}
      </FormField>
      <FormField id="checkout-recipient-address" label="Địa chỉ" errorMessages={errors.address ? [errors.address] : []}>
        {(control) => <textarea {...control} disabled={disabled} rows={3} value={value.address} onChange={(event) => onChange('address', event.target.value)} className={`${inputClass} resize-y`} />}
      </FormField>
    </div>
  )
}
