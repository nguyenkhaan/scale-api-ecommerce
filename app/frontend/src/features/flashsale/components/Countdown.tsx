import { useEffect, useState } from 'react'
import type { FlashSaleStatus } from '../flashsale.service'
import { countdownTarget, formatCountdown } from '../flashsale.logic'

type Props = { status: FlashSaleStatus; startsAt: string; endsAt: string }

/** Display-only ticking countdown; no business decision depends on the browser clock. */
export function Countdown({ status, startsAt, endsAt }: Props) {
  const [now, setNow] = useState(() => Date.now())
  const target = countdownTarget(status, startsAt, endsAt)

  const targetAt = target?.at

  useEffect(() => {
    if (!targetAt) return
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [targetAt])

  if (!target) return null
  const remaining = formatCountdown(target.at, now)
  return (
    <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-[var(--ss-color-muted)]">
      <span>{target.label}:</span>
      <span role="timer" aria-live="off" aria-atomic="true" className="font-mono font-semibold tabular-nums text-[var(--ss-color-text)]">
        {remaining ?? 'Đang chờ hệ thống cập nhật trạng thái chương trình.'}
      </span>
      <span className="text-[11px]">Thời gian chính thức theo hệ thống</span>
    </div>
  )
}
