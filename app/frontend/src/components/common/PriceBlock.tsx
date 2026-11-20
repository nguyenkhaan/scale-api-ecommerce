import { formatDateTime, formatVND } from '@/lib/format'
import { StatusBadge } from './StatusBadge'

export type PriceBlockProps = {
  regularPrice: number
  effectivePrice: number
  source: 'regular' | 'flash_sale'
  endsAt?: string
  compact?: boolean
}

export function PriceBlock({ regularPrice, effectivePrice, source, endsAt, compact = false }: PriceBlockProps) {
  if (source === 'regular') {
    return (
      <div data-price-source={source} className="flex flex-col gap-1">
        <span className="text-xs text-[var(--ss-color-muted)]">Giá thường</span>
        <strong className={`${compact ? 'text-sm' : 'text-base'} font-semibold text-[var(--ss-color-primary-strong)]`}>{formatVND(effectivePrice)}</strong>
      </div>
    )
  }

  return (
    <div data-price-source={source} className="flex flex-col gap-1">
      <span className="text-xs text-[var(--ss-color-muted)]">Có ưu đãi đang chạy</span>
      <div className="flex flex-wrap items-baseline gap-2">
        <strong className={`${compact ? 'text-base' : 'text-xl'} font-bold text-[var(--ss-color-primary)]`}>
          <span className="sr-only">Giá ưu đãi </span>
          {formatVND(effectivePrice)}
        </strong>
        <del className="text-xs text-[var(--ss-color-muted)]">
          <span className="sr-only">Giá thường </span>
          {formatVND(regularPrice)}
        </del>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <StatusBadge domain="priceSource" value="flash_sale" />
        {endsAt && <span className="text-xs text-[var(--ss-color-muted)]">Kết thúc {formatDateTime(endsAt)}</span>}
      </div>
    </div>
  )
}
