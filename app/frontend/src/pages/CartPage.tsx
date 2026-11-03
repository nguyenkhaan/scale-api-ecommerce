import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { isApiError } from '../lib/api-error'
import {
  cartClient,
  type Cart,
  type CartItem,
  type CheckoutPreview,
} from '../lib/cart-api'

function formatVND(amount: number): string {
  return new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount)
}

export function CartPage() {
  const [cart, setCart] = useState<Cart | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [preview, setPreview] = useState<CheckoutPreview | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [updatingSkuId, setUpdatingSkuId] = useState<string | null>(null)

  const loadCart = async () => {
    setLoading(true)
    setError(null)
    setPreview(null)
    try {
      const data = await cartClient.getCart()
      setCart(data)
      // Auto-select all items by default
      setSelectedIds(new Set(data.items.map((it) => it.id)))
    } catch (err) {
      if (isApiError(err)) {
        setError(`Không thể tải giỏ hàng (${err.code}).`)
      } else {
        setError('Lỗi kết nối khi tải giỏ hàng.')
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let active = true
    cartClient
      .getCart()
      .then((data) => {
        if (!active) return
        setCart(data)
        setSelectedIds(new Set(data.items.map((it) => it.id)))
      })
      .catch((err) => {
        if (!active) return
        if (isApiError(err)) {
          setError(`Không thể tải giỏ hàng (${err.code}).`)
        } else {
          setError('Lỗi kết nối khi tải giỏ hàng.')
        }
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [])

  const toggleSelect = (itemId: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(itemId)) {
        next.delete(itemId)
      } else {
        next.add(itemId)
      }
      return next
    })
    setPreview(null)
  }

  const toggleSelectAll = () => {
    if (!cart) return
    if (selectedIds.size === cart.items.length) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(cart.items.map((it) => it.id)))
    }
    setPreview(null)
  }

  const handleUpdateQuantity = async (item: CartItem, newQty: number) => {
    if (newQty < 1) return
    setUpdatingSkuId(item.sku_id)
    setError(null)
    try {
      const updated = await cartClient.upsertCartItem(item.sku_id, {
        quantity: newQty,
        price_source: item.price_source,
        flash_sale_item_id: item.flash_sale_item_id,
        expected_version: item.version,
      })
      setCart(updated)
      setPreview(null)
    } catch (err) {
      if (isApiError(err)) {
        if (err.status === 409) {
          setError('Giỏ hàng đã bị thay đổi bởi thao tác khác. Vui lòng làm mới.')
        } else {
          setError(`Cập nhật số lượng thất bại: ${err.code}`)
        }
      } else {
        setError('Lỗi kết nối khi cập nhật giỏ hàng.')
      }
    } finally {
      setUpdatingSkuId(null)
    }
  }

  const handleRemove = async (skuId: string) => {
    setUpdatingSkuId(skuId)
    setError(null)
    try {
      const updated = await cartClient.removeCartItem(skuId)
      setCart(updated)
      setSelectedIds((prev) => {
        const next = new Set(prev)
        const removed = cart?.items.find((i) => i.sku_id === skuId)
        if (removed) next.delete(removed.id)
        return next
      })
      setPreview(null)
    } catch (err) {
      if (isApiError(err)) {
        setError(`Xóa sản phẩm thất bại (${err.code}).`)
      } else {
        setError('Lỗi kết nối khi xóa sản phẩm.')
      }
    } finally {
      setUpdatingSkuId(null)
    }
  }

  const handlePreview = async () => {
    if (!cart) return
    const selectedItems = cart.items.filter((it) => selectedIds.has(it.id))
    if (selectedItems.length === 0) return

    setPreviewLoading(true)
    setPreviewError(null)
    try {
      const result = await cartClient.previewCheckout({
        items: selectedItems.map((it) => ({
          cart_item_id: it.id,
          version: it.version,
        })),
      })
      setPreview(result)
    } catch (err) {
      if (isApiError(err)) {
        if (err.status === 409) {
          setPreviewError(
            'Điều kiện giỏ hàng hoặc flash sale đã thay đổi (giới hạn mua, suất ưu đãi hoặc phiên bản giỏ hàng). Vui lòng làm mới giỏ hàng.',
          )
        } else {
          setPreviewError(`Không thể xem trước đơn hàng (${err.code}).`)
        }
      } else {
        setPreviewError('Lỗi kết nối khi xem trước đơn hàng.')
      }
    } finally {
      setPreviewLoading(false)
    }
  }

  return (
    <div className="mx-auto max-w-5xl">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">Giỏ hàng của bạn</h1>
          <p className="mt-1 text-sm text-slate-600">
            {cart ? `${cart.items.length}/20 mặt hàng (SKU)` : '0 mặt hàng'}
          </p>
        </div>
        <Link
          to="/flash-sales"
          className="inline-flex items-center text-sm font-semibold text-amber-700 hover:text-amber-800"
        >
          ← Tiếp tục săn Flash Sale
        </Link>
      </div>

      {loading && (
        <div className="py-12 text-center text-slate-500" role="status">
          Đang tải giỏ hàng của bạn...
        </div>
      )}

      {error && (
        <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-4 text-rose-800">
          <p className="font-semibold">{error}</p>
          <button
            type="button"
            onClick={() => void loadCart()}
            className="mt-2 text-sm font-semibold underline hover:text-rose-950"
          >
            Làm mới giỏ hàng
          </button>
        </div>
      )}

      {(!cart || cart.items.length === 0) && (
        <div className="mt-8 rounded border border-dashed border-slate-300 p-12 text-center">
          <p className="text-base font-medium text-slate-700">Giỏ hàng đang trống.</p>
          <p className="mt-1 text-sm text-slate-500">Hãy thêm sản phẩm từ Flash Sale hoặc Danh mục sản phẩm!</p>
          <div className="mt-6">
            <Link
              to="/flash-sales"
              className="inline-flex rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800"
            >
              Khám phá Flash Sale
            </Link>
          </div>
        </div>
      )}

      {cart && cart.items.length > 0 && (
        <div className="mt-6 grid gap-8 lg:grid-cols-3">
          {/* Items List */}
          <div className="lg:col-span-2 space-y-4">
            <div className="flex items-center justify-between rounded bg-slate-50 px-4 py-2 border border-slate-200">
              <label className="flex items-center gap-2 text-sm font-medium text-slate-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedIds.size === cart.items.length && cart.items.length > 0}
                  onChange={toggleSelectAll}
                  className="rounded border-slate-300 text-amber-600 focus:ring-amber-500"
                />
                Chọn tất cả ({cart.items.length})
              </label>
              <span className="text-xs text-slate-500">Tối đa 20 SKU</span>
            </div>

            <div className="divide-y divide-slate-200 rounded-lg border border-slate-200 bg-white">
              {cart.items.map((item) => {
                const isSelected = selectedIds.has(item.id)
                const isBusy = updatingSkuId === item.sku_id

                return (
                  <div key={item.id} className="flex items-start gap-4 p-4">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggleSelect(item.id)}
                      className="mt-1 rounded border-slate-300 text-amber-600 focus:ring-amber-500 cursor-pointer"
                    />

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-xs text-slate-500">
                          SKU: {item.sku_id.slice(0, 8)}...
                        </span>
                        {item.price_source === 'flash_sale' ? (
                          <span className="rounded bg-rose-100 px-2 py-0.5 text-xs font-bold text-rose-800">
                            Flash Sale
                          </span>
                        ) : (
                          <span className="rounded bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700">
                            Giá thường
                          </span>
                        )}
                      </div>

                      <div className="mt-3 flex items-center gap-4">
                        <div className="flex items-center border border-slate-300 rounded">
                          <button
                            type="button"
                            disabled={isBusy || item.quantity <= 1}
                            onClick={() => void handleUpdateQuantity(item, item.quantity - 1)}
                            className="px-2.5 py-1 text-sm font-bold text-slate-600 hover:bg-slate-100 disabled:opacity-40"
                          >
                            −
                          </button>
                          <span className="px-3 py-1 text-sm font-semibold text-slate-900 min-w-[2rem] text-center">
                            {item.quantity}
                          </span>
                          <button
                            type="button"
                            disabled={isBusy}
                            onClick={() => void handleUpdateQuantity(item, item.quantity + 1)}
                            className="px-2.5 py-1 text-sm font-bold text-slate-600 hover:bg-slate-100 disabled:opacity-40"
                          >
                            +
                          </button>
                        </div>

                        <button
                          type="button"
                          disabled={isBusy}
                          onClick={() => void handleRemove(item.sku_id)}
                          className="text-xs font-semibold text-rose-600 hover:text-rose-800 disabled:opacity-40"
                        >
                          Xóa
                        </button>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Checkout Preview Sidebar */}
          <div className="lg:col-span-1">
            <div className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm sticky top-6">
              <h2 className="text-lg font-bold text-slate-900">Xem trước đơn hàng</h2>
              <p className="mt-1 text-xs text-slate-500">
                Tính toán giá ưu đãi, quota và giới hạn mua thời gian thực (chưa giữ tồn kho).
              </p>

              {previewError && (
                <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-3 text-xs text-rose-800">
                  <p className="font-semibold">{previewError}</p>
                  <button
                    type="button"
                    onClick={() => void loadCart()}
                    className="mt-1 font-semibold underline hover:text-rose-950"
                  >
                    Làm mới giỏ hàng
                  </button>
                </div>
              )}

              <div className="mt-4 border-t border-slate-200 pt-4">
                <div className="flex justify-between text-sm text-slate-600">
                  <span>Mặt hàng đã chọn</span>
                  <span className="font-semibold">{selectedIds.size}</span>
                </div>

                {preview && (
                  <div className="mt-4 space-y-3">
                    <div className="border-t border-dashed border-slate-200 pt-3">
                      <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-2">Chi tiết dòng tính giá:</h3>
                      <div className="space-y-2">
                        {preview.lines.map((line) => (
                          <div key={line.cart_item_id} className="text-xs">
                            <div className="flex justify-between font-medium text-slate-700">
                              <span>SKU {line.sku_id.slice(0, 8)}... × {line.quantity}</span>
                              <span>{formatVND(line.subtotal)}</span>
                            </div>
                            <div className="text-[10px] text-slate-400">
                              Đơn giá: {formatVND(line.unit_price)} ({line.price_source === 'flash_sale' ? 'Flash Sale' : 'Thường'})
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>

                    <div className="border-t border-slate-200 pt-3 flex justify-between text-base font-bold text-slate-900">
                      <span>Tổng cộng ({preview.currency})</span>
                      <span className="text-rose-600">{formatVND(preview.total_amount)}</span>
                    </div>

                    <p className="text-[10px] text-slate-400 italic">
                      Xem trước tại: {new Date(preview.previewed_at).toLocaleTimeString('vi-VN')}
                    </p>
                  </div>
                )}
              </div>

              <div className="mt-6">
                <button
                  type="button"
                  disabled={previewLoading || selectedIds.size === 0}
                  onClick={() => void handlePreview()}
                  className="w-full rounded bg-amber-600 px-4 py-2.5 text-sm font-bold text-white shadow-sm hover:bg-amber-500 focus-visible:outline-2 focus-visible:outline-amber-600 disabled:cursor-not-allowed disabled:bg-slate-300"
                >
                  {previewLoading ? 'Đang tính toán giá...' : 'Xem trước thanh toán'}
                </button>
                {preview && (
                  <p className="mt-2 text-center text-xs text-emerald-700 font-semibold">
                    ✓ Đã xác thực giá và điều kiện ưu đãi hiện tại.
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
