import { canSwitchToRegular } from '@/features/cart/cart.logic'
import type { LineProblem } from '@/features/cart/cart.hook'
import type { CartItem } from '@/features/cart/cart.type'
import { Button, PriceBlock, QuantityStepper, StatusBadge, Surface } from '@/components/common'

type Props = {
  item: CartItem
  selected: boolean
  busy: boolean
  problem?: LineProblem
  onToggle: () => void
  onQuantity: (quantity: number) => void
  onRemove: () => void
  onSwitchToRegular: () => void
}

export function PriceSourceBadge({ source }: { source: CartItem['price_source'] }) {
  return <StatusBadge domain="priceSource" value={source} />
}

export function CartLineItem({
  item,
  selected,
  busy,
  problem,
  onToggle,
  onQuantity,
  onRemove,
  onSwitchToRegular,
}: Props) {
  return (
    <li>
      <Surface as="article" className="flex items-start gap-3 p-4 sm:gap-4">
        <input
          type="checkbox"
          checked={selected}
          onChange={onToggle}
          aria-label={`Chọn ${item.product_name}`}
          className="mt-1 size-4 cursor-pointer"
        />

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="break-words font-semibold text-[var(--ss-color-text)]">{item.product_name}</p>
              <div className="mt-1 flex flex-wrap items-center gap-2">
                <span className="font-mono text-xs text-[var(--ss-color-muted)]">SKU: {item.sku_code}</span>
                <PriceSourceBadge source={item.price_source} />
              </div>
            </div>
            <PriceBlock regularPrice={item.reference_unit_price} effectivePrice={item.reference_unit_price} source="regular" compact />
          </div>

          <div className="mt-3 flex flex-wrap items-center gap-4">
            <QuantityStepper value={item.quantity} min={1} max={Number.MAX_SAFE_INTEGER} busy={busy} label={`Số lượng ${item.product_name}`} onChange={onQuantity} />
            <Button variant="ghost" disabled={busy} onClick={onRemove}>Xóa</Button>
          </div>

          {problem && (
            <div role="alert" className="mt-3 border border-red-200 bg-[var(--ss-color-danger-surface)] p-3 text-sm text-[var(--ss-color-danger)]">
              <p className="font-semibold">{problem.message}</p>
              {problem.code && canSwitchToRegular(item.price_source, problem.code) && (
                <Button variant="ghost" disabled={busy} onClick={onSwitchToRegular} className="mt-1 px-0 underline">
                  Chuyển sang giá thường
                </Button>
              )}
            </div>
          )}
        </div>
      </Surface>
    </li>
  )
}
