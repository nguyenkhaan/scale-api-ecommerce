import { Link } from 'react-router-dom'
import { CartLineItem } from '../cart/CartLineItem'
import { CheckoutPreviewPanel } from '../cart/CheckoutPreviewPanel'
import { useCart } from '../cart/use-cart'
import { useCheckoutPreview } from '../cart/use-checkout-preview'

const MAX_SKUS = 20

export function CartPage() {
  const cart = useCart()
  const checkout = useCheckoutPreview(cart)
  const { items, selectedItems } = cart

  return (
    <div className="mx-auto max-w-5xl">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">Giỏ hàng của bạn</h1>
          <p className="mt-1 text-sm text-slate-600">{items.length}/{MAX_SKUS} mặt hàng (SKU)</p>
        </div>
        <Link to="/flash-sales" className="inline-flex min-h-11 items-center text-sm font-semibold text-amber-700 hover:text-amber-800">
          ← Tiếp tục săn Flash Sale
        </Link>
      </div>

      <p className="mt-4 rounded border border-slate-200 bg-slate-50 p-3 text-sm text-slate-700">
        Giỏ hàng không giữ tồn kho và không khóa giá. Giá chính thức được tính ở bước xem trước thanh toán.
      </p>

      {cart.loading && (
        <div className="py-12 text-center text-slate-500" role="status">
          Đang tải giỏ hàng của bạn...
        </div>
      )}

      {cart.loadError && (
        <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-4 text-rose-800">
          <p className="font-semibold">{cart.loadError}</p>
          <button type="button" onClick={() => void cart.reload()} className="mt-2 min-h-9 text-sm font-semibold underline hover:text-rose-950">
            Tải lại giỏ hàng
          </button>
        </div>
      )}

      {cart.notice && (
        <div
          role={cart.notice.kind === 'error' ? 'alert' : 'status'}
          className={`mt-4 rounded border p-3 text-sm font-medium ${
            cart.notice.kind === 'error'
              ? 'border-rose-200 bg-rose-50 text-rose-800'
              : 'border-emerald-300 bg-emerald-50 text-emerald-900'
          }`}
        >
          {cart.notice.text}
        </div>
      )}

      {!cart.loading && !cart.loadError && items.length === 0 && (
        <div className="mt-8 rounded border border-dashed border-slate-300 p-12 text-center">
          <p className="text-base font-medium text-slate-700">Giỏ hàng đang trống.</p>
          <p className="mt-1 text-sm text-slate-500">Hãy thêm sản phẩm từ Flash Sale hoặc danh sách sản phẩm.</p>
          <div className="mt-6 flex flex-wrap justify-center gap-3">
            <Link to="/flash-sales" className="inline-flex min-h-11 items-center rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800">
              Khám phá Flash Sale
            </Link>
            <Link to="/products" className="inline-flex min-h-11 items-center rounded border border-slate-900 px-4 py-2 text-sm font-semibold text-slate-900 hover:bg-stone-100">
              Xem sản phẩm
            </Link>
          </div>
        </div>
      )}

      {items.length > 0 && (
        <div className="mt-6 grid gap-8 lg:grid-cols-3">
          <div className="space-y-4 lg:col-span-2">
            <div className="flex items-center justify-between rounded border border-slate-200 bg-slate-50 px-4 py-2">
              <label className="flex cursor-pointer items-center gap-2 text-sm font-medium text-slate-700">
                <input
                  type="checkbox"
                  checked={selectedItems.length === items.length}
                  onChange={cart.toggleAll}
                  className="size-4 rounded border-slate-300 text-amber-600 focus:ring-amber-500"
                />
                Chọn tất cả ({items.length})
              </label>
              <span className="text-xs text-slate-500">Tối đa {MAX_SKUS} SKU</span>
            </div>

            <ul className="divide-y divide-slate-200 rounded-lg border border-slate-200 bg-white">
              {items.map((item) => (
                <CartLineItem
                  key={item.id}
                  item={item}
                  selected={selectedItems.some((s) => s.id === item.id)}
                  busy={cart.busySku === item.sku_id}
                  problem={cart.lineProblems[item.sku_id]}
                  onToggle={() => cart.toggle(item.id)}
                  onQuantity={(q) => void cart.changeQuantity(item, q)}
                  onRemove={() => void cart.remove(item)}
                  onSwitchToRegular={() => void cart.switchToRegular(item)}
                />
              ))}
            </ul>
          </div>

          <div className="lg:col-span-1">
            <CheckoutPreviewPanel
              selectedCount={selectedItems.length}
              items={items}
              preview={checkout.preview}
              changes={checkout.changes}
              acknowledged={checkout.acknowledged}
              loading={checkout.loading}
              error={checkout.error}
              onPreview={() => void checkout.run()}
              onAcknowledge={checkout.acknowledge}
            />
          </div>
        </div>
      )}
    </div>
  )
}
