import type { Dispatch, SetStateAction } from 'react'

import { Button, Surface } from '@/components/common'

import { CheckoutOutcomeView } from './CheckoutOutcomeView'
import { RecipientForm } from './RecipientForm'
import type { CheckoutOutcome, RecipientErrors, RecipientInput } from '../checkout.type'

export type PlaceOrderState = {
  recipient: RecipientInput
  setRecipient: Dispatch<SetStateAction<RecipientInput>>
  errors: RecipientErrors
  submitting: boolean
  outcome: CheckoutOutcome | null
  retryCountdown: number | null
  submit: () => Promise<void>
  retry: () => Promise<void>
  checkResult: () => Promise<void>
}

type Props = { state: PlaceOrderState }

export function PlaceOrderSection({ state }: Props) {
  return (
    <Surface as="section" className="mt-4 p-4 sm:p-5" aria-label="Đặt hàng">
      <h3 className="text-base font-semibold text-[var(--ss-color-text)]">Thông tin nhận hàng</h3>
      <div className="mt-4">
        <RecipientForm
          value={state.recipient}
          errors={state.errors}
          disabled={state.submitting}
          onChange={(field, value) => state.setRecipient((current) => ({ ...current, [field]: value }))}
        />
      </div>

      <dl className="mt-4 space-y-2 border-t border-[var(--ss-color-border)] pt-4 text-sm text-[var(--ss-color-muted)]">
        <div className="flex justify-between gap-3">
          <dt>Giữ hàng sau khi đặt</dt>
          <dd className="font-medium text-[var(--ss-color-text)]">5 phút</dd>
        </div>
      </dl>

      {state.outcome && (
        <div className="mt-4">
          <CheckoutOutcomeView outcome={state.outcome} onRetry={() => void state.retry()} onCheck={() => void state.checkResult()} retryCountdown={state.retryCountdown} />
        </div>
      )}

      {!state.outcome || state.outcome.kind === 'price_changed' || state.outcome.kind === 'cart_changed' ? (
        <Button className="mt-4 w-full" disabled={state.submitting} onClick={() => void state.submit()}>
          {state.submitting ? 'Đang đặt hàng…' : 'Đặt hàng'}
        </Button>
      ) : null}
      {state.submitting && <p role="status" className="mt-2 text-center text-xs text-[var(--ss-color-muted)]">Đang gửi yêu cầu. Vui lòng không đóng trang.</p>}
    </Surface>
  )
}
