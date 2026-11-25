import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'

import { CheckoutPreviewPanel } from '@/features/cart/components/CheckoutPreviewPanel'
import { CheckoutOutcomeView } from '@/features/checkout/components/CheckoutOutcomeView'
import { PlaceOrderSection, type PlaceOrderState } from '@/features/checkout/components/PlaceOrderSection'
import type { CheckoutPreview } from '@/features/cart/cart.type'
import type { CheckoutOutcome, RecipientInput } from '@/features/checkout/checkout.type'

const preview: CheckoutPreview = {
  currency: 'VND',
  total_amount: 200000,
  previewed_at: '2026-11-05T00:00:00Z',
  lines: [
    {
      cart_item_id: 'i1',
      cart_item_version: 1,
      sku_id: 's1',
      quantity: 2,
      unit_price: 100000,
      subtotal: 200000,
      price_source: 'flash_sale',
      flash_sale_item_id: 'f1',
    },
  ],
}

const recipient: RecipientInput = { name: 'Nguyễn Minh Anh', phone: '0901234567', address: '12 Nguyễn Trãi' }

function state(overrides: Partial<PlaceOrderState> = {}): PlaceOrderState {
  return {
    recipient,
    setRecipient: () => {},
    errors: {},
    submitting: false,
    outcome: null,
    retryCountdown: null,
    submit: async () => {},
    retry: async () => {},
    checkResult: async () => {},
    ...overrides,
  }
}

function panel(orderState: PlaceOrderState | null = null, acknowledged = true) {
  return renderToStaticMarkup(
    <MemoryRouter>
      <CheckoutPreviewPanel
        selectedCount={1}
        items={[]}
        preview={acknowledged ? preview : null}
        changes={[]}
        acknowledged={acknowledged}
        loading={false}
        error={null}
        onPreview={() => {}}
        onAcknowledge={() => {}}
        placeOrder={orderState ?? undefined}
      />
    </MemoryRouter>,
  )
}

describe('checkout form and outcomes', () => {
  it('does not render recipient form or place-order action before an acknowledged preview', () => {
    const markup = panel(null, false)

    expect(markup).not.toContain('Họ tên')
    expect(markup).not.toContain('Đặt hàng')
  })

  it('renders recipient form, place-order action and five-minute hold after acknowledgement', () => {
    const markup = panel(state())

    expect(markup).toContain('Họ tên')
    expect(markup).toContain('Số điện thoại')
    expect(markup).toContain('Địa chỉ')
    expect(markup).toContain('Đặt hàng')
    expect(markup).toContain('Giữ hàng sau khi đặt')
    expect(markup).toContain('5 phút')
  })

  it('shows honest submitting state', () => {
    const markup = panel(state({ submitting: true }))

    expect(markup).toContain('Đang đặt hàng…')
    expect(markup).toContain('Đang gửi yêu cầu. Vui lòng không đóng trang.')
    expect(markup).toMatch(/disabled=""[^>]*>Đang đặt hàng…/)
  })

  it('connects recipient errors to fields accessibly', () => {
    const markup = renderToStaticMarkup(
      <PlaceOrderSection
        state={state({ errors: { name: 'Vui lòng nhập họ tên.', phone: 'Số điện thoại cần 8–20 ký tự.' } })}
      />,
    )

    expect(markup).toContain('Vui lòng nhập họ tên.')
    expect(markup).toContain('Số điện thoại cần 8–20 ký tự.')
    expect((markup.match(/aria-invalid="true"/g) ?? []).length).toBe(2)
  })

  it('never renders success for an unknown result', () => {
    const outcome: CheckoutOutcome = { kind: 'unknown' }
    const markup = renderToStaticMarkup(<CheckoutOutcomeView outcome={outcome} onRetry={() => {}} onCheck={() => {}} retryCountdown={null} />)

    expect(markup).toContain('Đang kiểm tra kết quả đặt đơn…')
    expect(markup).toContain('Kết nối bị gián đoạn nên chưa nhận được kết quả. Hệ thống đang tự kiểm tra lại, bạn không cần đặt lại đơn.')
    expect(markup).toContain('Kiểm tra lại')
    expect(markup).toContain('Thử lại')
    expect(markup).not.toContain('Đặt đơn thành công')
  })

  it('shows processing and request-id-specific unknown copy', () => {
    expect(renderToStaticMarkup(<CheckoutOutcomeView outcome={{ kind: 'processing', retryAfterSeconds: 1 }} onRetry={() => {}} onCheck={() => {}} retryCountdown={null} />)).toContain(
      'Yêu cầu trước đó đang được xử lý, vui lòng đợi.',
    )
    const markup = renderToStaticMarkup(<CheckoutOutcomeView outcome={{ kind: 'unknown', requestID: 'req_1' }} onRetry={() => {}} onCheck={() => {}} retryCountdown={null} />)
    expect(markup).toContain('Hệ thống tạm thời không xử lý được')
    expect(markup).toContain('Mã yêu cầu: req_1')
  })

  it('shows overload countdown and rejected/price-changed actions', () => {
    const overload = renderToStaticMarkup(
      <CheckoutOutcomeView outcome={{ kind: 'overload', retryAfterSeconds: 15 }} onRetry={() => {}} onCheck={() => {}} retryCountdown={15} />,
    )
    expect(overload).toContain('Hệ thống đang quá tải')
    expect(overload).toContain('Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên.')
    expect(overload).toMatch(/disabled=""[^>]*>Thử lại sau 00:15/)

    const rejected = renderToStaticMarkup(<CheckoutOutcomeView outcome={{ kind: 'rejected', code: 'INSUFFICIENT_STOCK', lineErrors: [] }} onRetry={() => {}} onCheck={() => {}} retryCountdown={null} />)
    expect(rejected).toContain('Chưa đặt được đơn hàng')
    expect(rejected).toContain('Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên.')

    const changed = renderToStaticMarkup(<CheckoutOutcomeView outcome={{ kind: 'price_changed', lineErrors: [] }} onRetry={() => {}} onCheck={() => {}} retryCountdown={null} />)
    expect(changed).toContain('Giá đã thay đổi')
    expect(changed).toContain('Tôi đồng ý với giá mới')
  })
})
