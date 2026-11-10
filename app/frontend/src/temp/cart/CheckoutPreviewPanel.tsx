import { useState } from 'react'
import type { CartItem, CheckoutPreview } from '../lib/cart-api'
import { formatVND } from '../lib/format'
import type { PreviewChange } from './preview-changes'

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

const sourceLabel = (source: string) => (source === 'flash_sale' ? 'Flash Sale' : 'Giá thường')

export function CheckoutPreviewPanel(props: Props) {
  const { selectedCount, items, preview, changes, acknowledged, loading, error } = props
  const [orderNote, setOrderNote] = useState(false)
  const lineName = (skuId: string) => items.find((i) => i.sku_id === skuId)?.product_name ?? 'Sản phẩm'
  const changeFor = (skuId: string) => changes.find((c) => c.sku_id === skuId)
  const canOrder = preview !== null && acknowledged

  return (
    <div className="sticky top-6 rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-bold text-slate-900">Xem trước đơn hàng</h2>
      <p className="mt-1 text-xs text-slate-500">
        Tính giá ưu đãi, suất và giới hạn mua theo thời điểm hiện tại. Xem trước không giữ tồn kho.
      </p>

      {error && (
        <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-3 text-xs text-rose-800">
          <p className="font-semibold">{error}</p>
        </div>
      )}

      <div className="mt-4 border-t border-slate-200 pt-4">
        <div className="flex justify-between text-sm text-slate-600">
          <span>Mặt hàng đã chọn</span>
          <span className="font-semibold">{selectedCount}</span>
        </div>

        {preview && (
          <div className="mt-4 space-y-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500">Chi tiết dòng tính giá</h3>
            <ul className="space-y-2">
              {preview.lines.map((line) => {
                const change = changeFor(line.sku_id)
                return (
                  <li
                    key={line.cart_item_id}
                    className={`text-xs ${change ? 'rounded border border-amber-400 bg-amber-50 p-2' : ''}`}
                  >
                    <div className="flex justify-between gap-2 font-medium text-slate-700">
                      <span>
                        {lineName(line.sku_id)} × {line.quantity}
                      </span>
                      <span>{formatVND(line.subtotal)}</span>
                    </div>
                    <div className="text-[11px] text-slate-500">
                      Đơn giá: {formatVND(line.unit_price)} ({sourceLabel(line.price_source)})
                    </div>
                    {change && (
                      <p className="mt-1 text-[11px] font-bold text-amber-900">
                        Giá đã đổi: {formatVND(change.oldUnitPrice)} ({sourceLabel(change.oldSource)}) →{' '}
                        {formatVND(change.newUnitPrice)} ({sourceLabel(change.newSource)})
                      </p>
                    )}
                  </li>
                )
              })}
            </ul>

            <div className="flex justify-between border-t border-slate-200 pt-3 text-base font-bold text-slate-900">
              <span>Tổng cộng ({preview.currency})</span>
              <span className="text-rose-600">{formatVND(preview.total_amount)}</span>
            </div>
            <p className="text-[11px] italic text-slate-500">
              Xem trước lúc {new Date(preview.previewed_at).toLocaleTimeString('vi-VN')}
            </p>

            {changes.length > 0 && !acknowledged && (
              <div role="alert" className="rounded border border-amber-400 bg-amber-50 p-3 text-xs text-amber-900">
                <p className="font-semibold">Giá của một số dòng đã thay đổi so với lần xem trước trước đó.</p>
                <button
                  type="button"
                  onClick={props.onAcknowledge}
                  className="mt-2 min-h-9 border border-amber-700 bg-white px-3 font-bold text-amber-900 hover:bg-amber-100 focus-visible:outline-2 focus-visible:outline-amber-500"
                >
                  Tôi đã xem giá mới
                </button>
              </div>
            )}
            {changes.length > 0 && acknowledged && (
              <p role="status" className="text-xs font-semibold text-emerald-700">
                ✓ Bạn đã xác nhận giá mới.
              </p>
            )}
          </div>
        )}
      </div>

      <div className="mt-6 space-y-2">
        <button
          type="button"
          disabled={loading || selectedCount === 0}
          onClick={props.onPreview}
          className="min-h-11 w-full rounded bg-amber-600 px-4 py-2.5 text-sm font-bold text-white shadow-sm hover:bg-amber-500 focus-visible:outline-2 focus-visible:outline-amber-600 disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          {loading ? 'Đang tính toán giá...' : 'Xem trước thanh toán'}
        </button>
        <button
          type="button"
          disabled={!canOrder}
          onClick={() => setOrderNote(true)}
          className="min-h-11 w-full rounded border border-slate-900 px-4 py-2.5 text-sm font-bold text-slate-900 hover:bg-stone-100 focus-visible:outline-2 focus-visible:outline-amber-600 disabled:cursor-not-allowed disabled:border-slate-300 disabled:text-slate-400"
        >
          Đặt hàng (Phase 04)
        </button>
        {orderNote && canOrder && (
          <p role="status" className="text-center text-xs text-slate-600">
            Bước đặt hàng sẽ được bổ sung ở Phase 04. Giỏ hàng của bạn vẫn được giữ nguyên.
          </p>
        )}
      </div>
    </div>
  )
}
