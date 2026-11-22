import { StatusBadge, Surface } from '@/components/common'
import { formatVND } from '@/lib/format'

import { itemAttributesText } from '../bill.logic'
import type { OrderItem } from '../bill.type'

export function OrderItemsList({ items }: { items: OrderItem[] }) {
  return (
    <Surface as="section" aria-labelledby="order-items-heading" className="p-4">
      <h2 id="order-items-heading" className="text-lg font-semibold text-[var(--ss-color-text)]">Sản phẩm</h2>
      <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Thông tin được lưu tại thời điểm đặt đơn, không đổi theo giá hiện tại.</p>
      <ul className="mt-3 divide-y divide-[var(--ss-color-border)]">
        {items.map((item) => {
          const attributes = itemAttributesText(item.sku_attributes)
          return (
            <li key={item.id} className="flex flex-wrap items-start justify-between gap-3 py-3">
              <div className="min-w-0">
                <p className="font-medium text-[var(--ss-color-text)]">{item.product_name}</p>
                <p className="text-sm text-[var(--ss-color-muted)]">
                  Mã {item.sku_code}
                  {attributes !== '' && ` · ${attributes}`}
                </p>
                <StatusBadge domain="priceSource" value={item.price_source} className="mt-1" />
              </div>
              <div className="text-right text-sm">
                <p className="text-[var(--ss-color-muted)]">
                  {item.quantity} × {formatVND(item.unit_price)}
                </p>
                <p className="font-semibold text-[var(--ss-color-text)]">{formatVND(item.subtotal)}</p>
              </div>
            </li>
          )
        })}
      </ul>
    </Surface>
  )
}
