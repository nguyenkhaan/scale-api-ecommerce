import { useEffect } from 'react'

import { formatCountdown } from '../bill.logic'
import { useNow } from '../bill.hook'

type Props = {
  expiresAt: string
  /** Called once when the local clock passes the deadline, so the caller can refetch the official state. */
  onExpire?: () => void
}

/** Display-only countdown; the server decides whether the order is still payable. */
export function OrderCountdown({ expiresAt, onExpire }: Props) {
  const now = useNow()
  const text = formatCountdown(expiresAt, now)
  const expired = text === null

  useEffect(() => {
    if (expired) onExpire?.()
    // Fire on the transition to expired only, not whenever the callback identity changes.
  }, [expired]) // eslint-disable-line react-hooks/exhaustive-deps

  if (expired) return <span className="text-sm font-semibold text-[var(--ss-color-warning-strong)]">Đã quá hạn</span>
  return (
    <span className="text-sm font-semibold text-[var(--ss-color-primary-strong)]" data-countdown>
      Còn {text} để thanh toán
    </span>
  )
}
