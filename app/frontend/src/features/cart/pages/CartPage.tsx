import { Link } from 'react-router-dom'

import { Alert, FeedbackState, Surface } from '@/components/common'
import { CartLineItem } from '@/features/cart/components/CartLineItem'
import { CheckoutPreviewPanel } from '@/features/cart/components/CheckoutPreviewPanel'
import { useCart, useCheckoutPreview, type CartState, type CheckoutPreviewState } from '@/features/cart/cart.hook'
import type { LineError as CartLineError } from '@/features/cart/cart.type'
import { usePlaceOrder } from '@/features/checkout/checkout.hook'
import type { PlaceOrderState } from '@/features/checkout/components/PlaceOrderSection'

const maxSkus = 20

export function CartPage() {
  const cart = useCart()
  const checkout = useCheckoutPreview(cart)
  const placeOrder = usePlaceOrder({
    preview: checkout.preview,
    acknowledged: checkout.acknowledged,
    onPriceChanged: () => void checkout.run(),
    onCartChanged: () => void cart.reload(),
    onRejected: (errors) => cart.applyLineErrors(errors.map((error) => ({ ...error, code: error.code as CartLineError['code'] }))),
  })
  return <CartPageView cart={cart} checkout={checkout} placeOrder={placeOrder} />
}

export type CartPageViewProps = {
  cart: CartState
  checkout: CheckoutPreviewState
  placeOrder?: PlaceOrderState
}

export function CartPageView({ cart, checkout, placeOrder }: CartPageViewProps) {
  const { items, selectedItems } = cart

  return (
    <div className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-[var(--ss-color-border)] pb-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Giỏ hàng của bạn</h1>
          <p className="mt-1 text-sm text-[var(--ss-color-muted)]">
            {items.length}/{maxSkus} mặt hàng (SKU)
          </p>
        </div>
        <Link to="/flash-sales" className="inline-flex min-h-11 items-center text-sm font-semibold text-[var(--ss-color-primary-strong)] hover:underline">
          Tiếp tục săn Flash Sale
        </Link>
      </div>

      <Alert tone="info" className="mt-4">
        Giỏ hàng không giữ tồn kho và không khóa giá. Giá chính thức được tính ở bước xem trước thanh toán.
      </Alert>

      {cart.loading && (
        <div className="mt-6"><FeedbackState kind="loading" title="Đang tải giỏ hàng của bạn" message="Vui lòng chờ trong giây lát." /></div>
      )}

      {cart.loadError && (
        <div className="mt-4">
          <FeedbackState kind="error" title="Không thể tải giỏ hàng" message={cart.loadError} action={{ label: 'Tải lại giỏ hàng', onClick: () => void cart.reload() }} />
        </div>
      )}

      {cart.notice && <div className="mt-4"><Alert tone={cart.notice.kind === 'error' ? 'danger' : 'success'}>{cart.notice.text}</Alert></div>}

      {!cart.loading && !cart.loadError && items.length === 0 && (
        <div className="mt-6">
          <FeedbackState kind="empty" title="Giỏ hàng đang trống." message="Hãy thêm sản phẩm từ Flash Sale hoặc danh sách sản phẩm.">
          <div className="mt-6 flex flex-wrap justify-center gap-3">
            <Link to="/flash-sales" className="inline-flex min-h-11 items-center rounded-[var(--ss-radius-control)] bg-[var(--ss-color-primary)] px-4 py-2 text-sm font-semibold text-white hover:bg-[var(--ss-color-primary-strong)]">
              Khám phá Flash Sale
            </Link>
            <Link to="/products" className="inline-flex min-h-11 items-center rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] px-4 py-2 text-sm font-semibold text-[var(--ss-color-text)] hover:bg-[var(--ss-color-app-bg)]">
              Xem sản phẩm
            </Link>
          </div>
          </FeedbackState>
        </div>
      )}

      {items.length > 0 && (
        <div className="mt-6 grid gap-8 lg:grid-cols-3">
          <div className="space-y-4 lg:col-span-2">
            <Surface as="section" className="flex items-center justify-between px-4 py-2">
              <label className="flex cursor-pointer items-center gap-2 text-sm font-medium text-[var(--ss-color-text)]">
                <input
                  type="checkbox"
                  checked={selectedItems.length === items.length}
                  onChange={cart.toggleAll}
                  className="size-4"
                />
                Chọn tất cả ({items.length})
              </label>
              <span className="text-xs text-[var(--ss-color-muted)]">Tối đa {maxSkus} SKU</span>
            </Surface>

            <ul className="space-y-3">
              {items.map((item) => (
                <CartLineItem
                  key={item.id}
                  item={item}
                  selected={selectedItems.some((selected) => selected.id === item.id)}
                  busy={cart.busySku === item.sku_id}
                  problem={cart.lineProblems[item.sku_id]}
                  onToggle={() => cart.toggle(item.id)}
                  onQuantity={(quantity) => void cart.changeQuantity(item, quantity)}
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
              placeOrder={placeOrder}
            />
          </div>
        </div>
      )}
    </div>
  )
}
