import type { CartItem } from '../lib/cart-api'
import { formatVND } from '../lib/format'
import { canSwitchToRegular } from './cart-messages'
import type { LineProblem } from './use-cart'

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

const stepButton =
  'min-h-9 min-w-9 px-2.5 py-1 text-sm font-bold text-slate-600 hover:bg-slate-100 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-40'

export function PriceSourceBadge({ source }: { source: CartItem['price_source'] }) {
  return source === 'flash_sale' ? (
    <span className="rounded bg-rose-100 px-2 py-0.5 text-xs font-bold text-rose-800">Flash Sale</span>
  ) : (
    <span className="rounded bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700">Giá thường</span>
  )
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
    <li className="flex items-start gap-4 p-4">
      <input
        type="checkbox"
        checked={selected}
        onChange={onToggle}
        aria-label={`Chọn ${item.product_name}`}
        className="mt-1 size-4 cursor-pointer rounded border-slate-300 text-amber-600 focus:ring-amber-500"
      />

      <div className="min-w-0 flex-1">
        <p className="break-words font-semibold text-slate-900">{item.product_name}</p>
        <div className="mt-1 flex flex-wrap items-center gap-2">
          <span className="font-mono text-xs text-slate-500">SKU: {item.sku_code}</span>
          <PriceSourceBadge source={item.price_source} />
        </div>
        <p className="mt-1 text-sm text-slate-600">
          Đơn giá tham khảo: <span className="font-semibold">{formatVND(item.reference_unit_price)}</span>
        </p>

        <div className="mt-3 flex flex-wrap items-center gap-4">
          <div className="flex items-center rounded border border-slate-300">
            <button
              type="button"
              aria-label={`Giảm số lượng ${item.product_name}`}
              disabled={busy || item.quantity <= 1}
              onClick={() => onQuantity(item.quantity - 1)}
              className={stepButton}
            >
              −
            </button>
            <span className="min-w-8 px-3 py-1 text-center text-sm font-semibold text-slate-900" aria-live="polite">
              {item.quantity}
            </span>
            <button
              type="button"
              aria-label={`Tăng số lượng ${item.product_name}`}
              disabled={busy}
              onClick={() => onQuantity(item.quantity + 1)}
              className={stepButton}
            >
              +
            </button>
          </div>
          <button
            type="button"
            disabled={busy}
            onClick={onRemove}
            className="min-h-9 text-xs font-semibold text-rose-700 hover:text-rose-900 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-40"
          >
            Xóa
          </button>
        </div>

        {problem && (
          <div role="alert" className="mt-3 rounded border border-rose-200 bg-rose-50 p-2 text-sm text-rose-800">
            <p className="font-semibold">{problem.message}</p>
            {problem.code && canSwitchToRegular(item.price_source, problem.code) && (
              <button
                type="button"
                disabled={busy}
                onClick={onSwitchToRegular}
                className="mt-1 min-h-9 text-xs font-bold underline hover:text-rose-950 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-40"
              >
                Chuyển sang giá thường
              </button>
            )}
          </div>
        )}
      </div>
    </li>
  )
}
