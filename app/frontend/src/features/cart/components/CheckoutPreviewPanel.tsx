import { useState } from 'react'

import { Alert, Button, PriceBlock, StatusBadge, Surface } from '@/components/common'
import type { PreviewChange } from '@/features/cart/cart.logic'
import type { CartItem, CheckoutPreview } from '@/features/cart/cart.type'
import { formatVND } from '@/lib/format'

type Props = {
  selectedCount: number
  items: CartItem[]
  preview: CheckoutPreview | null
  changes: PreviewChange[]
  acknowledged: boolean
  loading: boolean
  error: string | null
  onPreview: () => void
  onAcknowledge: () => void
}

export function CheckoutPreviewPanel(props: Props) {
  const { selectedCount, items, preview, changes, acknowledged, loading, error } = props
  const [orderNote, setOrderNote] = useState(false)
  const lineName = (skuId: string) => items.find((item) => item.sku_id === skuId)?.product_name ?? 'Sản phẩm'
  const changeFor = (skuId: string) => changes.find((change) => change.sku_id === skuId)
  const canOrder = preview !== null && acknowledged

  return (
    <Surface as="section" className="sticky top-6 p-4 sm:p-5">
      <h2 className="text-lg font-semibold text-[var(--ss-color-text)]">Xem trước đơn hàng</h2>
      <p className="mt-1 text-xs text-[var(--ss-color-muted)]">
        Tính giá ưu đãi, suất và giới hạn mua theo thời điểm hiện tại. Xem trước không giữ tồn kho.
      </p>

      {error && <div className="mt-4"><Alert tone="danger">{error}</Alert></div>}

      <div className="mt-4 border-t border-[var(--ss-color-border)] pt-4">
        <div className="flex justify-between text-sm text-[var(--ss-color-muted)]">
          <span>Mặt hàng đã chọn</span>
          <span className="font-semibold">{selectedCount}</span>
        </div>

        {preview && (
          <div className="mt-4 space-y-3">
            <h3 className="text-xs font-semibold text-[var(--ss-color-muted)]">Chi tiết dòng tính giá</h3>
            <ul className="space-y-2">
              {preview.lines.map((line) => {
                const change = changeFor(line.sku_id)
                return (
                  <li
                    key={line.cart_item_id}
                    className={`text-xs ${change ? 'border border-amber-300 bg-amber-50 p-2' : ''}`}
                  >
                    <div className="flex justify-between gap-2 font-medium text-[var(--ss-color-text)]">
                      <span>
                        {lineName(line.sku_id)} × {line.quantity}
                      </span>
                      <span className="font-semibold">{formatVND(line.subtotal)}</span>
                    </div>
                    <div className="mt-1 flex flex-wrap items-center gap-2 text-[11px] text-[var(--ss-color-muted)]">
                      <PriceBlock
                        regularPrice={items.find((item) => item.sku_id === line.sku_id)?.reference_unit_price ?? line.unit_price}
                        effectivePrice={line.unit_price}
                        source={line.price_source}
                        compact
                      />
                      <StatusBadge domain="priceSource" value={line.price_source} />
                    </div>
                    {change && (
                      <p className="mt-2 text-[11px] font-bold text-amber-900">
                        Giá đã đổi: {formatVND(change.oldUnitPrice)} ({change.oldSource === 'flash_sale' ? 'Flash Sale' : 'Giá thường'}) →{' '}
                        {formatVND(change.newUnitPrice)} ({change.newSource === 'flash_sale' ? 'Flash Sale' : 'Giá thường'})
                      </p>
                    )}
                  </li>
                )
              })}
            </ul>

            <div className="flex justify-between border-t border-[var(--ss-color-border)] pt-3 text-base font-bold text-[var(--ss-color-text)]">
              <span>Tổng cộng ({preview.currency})</span>
              <span className="text-[var(--ss-color-primary-strong)]">{formatVND(preview.total_amount)}</span>
            </div>
            <p className="text-[11px] italic text-[var(--ss-color-muted)]">Xem trước lúc {new Date(preview.previewed_at).toLocaleTimeString('vi-VN')}</p>

            {changes.length > 0 && !acknowledged && (
              <Alert tone="warning" className="text-xs">
                <p className="font-semibold">Giá của một số dòng đã thay đổi so với lần xem trước trước đó.</p>
                <Button variant="secondary" onClick={props.onAcknowledge} className="mt-2">
                  Tôi đã xem giá mới
                </Button>
              </Alert>
            )}
            {changes.length > 0 && acknowledged && (
              <p role="status" className="text-xs font-semibold text-[var(--ss-color-success-strong)]">
                Bạn đã xác nhận giá mới.
              </p>
            )}
          </div>
        )}
      </div>

      <div className="mt-6 space-y-2">
        <Button loading={loading} disabled={selectedCount === 0} onClick={props.onPreview} className="w-full">
          Xem trước thanh toán
        </Button>
        <Button variant="secondary" disabled={!canOrder} onClick={() => setOrderNote(true)} className="w-full">
          Đặt hàng (Phase 04)
        </Button>
        {orderNote && canOrder && (
          <p role="status" className="text-center text-xs text-[var(--ss-color-muted)]">
            Bước đặt hàng sẽ được bổ sung ở Phase 04. Giỏ hàng của bạn vẫn được giữ nguyên.
          </p>
        )}
      </div>
    </Surface>
  )
}
