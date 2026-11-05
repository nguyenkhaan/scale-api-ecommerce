import { useEffect, useState } from 'react'
import type { FlashSaleStatus } from '../lib/flash-sale-api'
import { countdownTarget, formatCountdown } from './flash-sale-status'

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
    <p className="mt-1 text-xs text-slate-600">
      {remaining ? (
        <>
          {target.label}: <span className="font-mono font-semibold tabular-nums">{remaining}</span>
        </>
      ) : (
        'Đang chờ hệ thống cập nhật trạng thái chương trình.'
      )}
      <span className="block text-[11px] text-slate-500">Thời gian chính thức theo hệ thống</span>
    </p>
  )
}
