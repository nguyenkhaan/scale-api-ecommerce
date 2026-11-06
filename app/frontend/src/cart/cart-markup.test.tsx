import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import { FlashSaleCampaignCard } from '../flash-sale/FlashSaleCampaignCard'
import type { CartItem, CheckoutPreview } from '../lib/cart-api'
import type { FlashSaleCampaign } from '../lib/flash-sale-api'
import { CartLineItem } from './CartLineItem'
import { CheckoutPreviewPanel } from './CheckoutPreviewPanel'

const item: CartItem = {
  id: 'i1',
  sku_id: 's1',
  quantity: 2,
  version: 1,
  price_source: 'flash_sale',
  flash_sale_item_id: 'f1',
  product_name: 'Áo thun',
  sku_code: 'AO-S',
  reference_unit_price: 90000,
}
const noop = () => {}
const orderButtonDisabled = /disabled=""[^>]*>Đặt hàng \(Phase 04\)/

describe('CartLineItem', () => {
  const render = (problem?: { message: string; code?: string }, source: CartItem['price_source'] = 'flash_sale') =>
    renderToStaticMarkup(
      <CartLineItem
        item={{ ...item, price_source: source }}
        selected
        busy={false}
        problem={problem}
        onToggle={noop}
        onQuantity={noop}
        onRemove={noop}
        onSwitchToRegular={noop}
      />,
    )

  it('shows product name, SKU code, badge and reference price', () => {
    const markup = render()
    expect(markup).toContain('Áo thun')
    expect(markup).toContain('SKU: AO-S')
    expect(markup).toContain('Flash Sale')
    expect(markup).toContain('90.000')
    expect(render(undefined, 'regular')).toContain('Giá thường')
  })

  it('shows the line error under the line with an explicit switch action for sale rejections', () => {
    const markup = render({ message: 'Đã hết suất giá flash sale — còn 1', code: 'SALE_QUOTA_EXCEEDED' })
    expect(markup).toContain('Đã hết suất giá flash sale — còn 1')
    expect(markup).toContain('Chuyển sang giá thường')
    expect(render({ message: 'Chỉ còn 1 sản phẩm', code: 'INSUFFICIENT_STOCK' })).not.toContain('Chuyển sang giá thường')
  })
})

describe('CheckoutPreviewPanel', () => {
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
  const base = { selectedCount: 1, items: [item], loading: false, error: null, onPreview: noop, onAcknowledge: noop }
  const change = {
    sku_id: 's1',
    oldUnitPrice: 90000,
    newUnitPrice: 100000,
    oldSource: 'flash_sale' as const,
    newSource: 'flash_sale' as const,
  }

  it('keeps the order button disabled until a preview exists', () => {
    const markup = renderToStaticMarkup(<CheckoutPreviewPanel {...base} preview={null} changes={[]} acknowledged={false} />)
    expect(markup).toMatch(orderButtonDisabled)
  })

  it('requires acknowledging changed prices before enabling the order button', () => {
    const pending = renderToStaticMarkup(
      <CheckoutPreviewPanel {...base} preview={preview} changes={[change]} acknowledged={false} />,
    )
    expect(pending).toContain('Tôi đã xem giá mới')
    expect(pending).toContain('Giá đã đổi')
    expect(pending).toMatch(orderButtonDisabled)
    const done = renderToStaticMarkup(<CheckoutPreviewPanel {...base} preview={preview} changes={[change]} acknowledged />)
    expect(done).not.toContain('Tôi đã xem giá mới')
    expect(done).not.toMatch(orderButtonDisabled)
  })

  it('shows the failure summary text', () => {
    const markup = renderToStaticMarkup(
      <CheckoutPreviewPanel
        {...base}
        preview={null}
        changes={[]}
        acknowledged={false}
        error="Chưa giữ hàng nào; giỏ hàng vẫn giữ nguyên."
      />,
    )
    expect(markup).toContain('role="alert"')
    expect(markup).toContain('Chưa giữ hàng nào')
  })
})

describe('FlashSaleCampaignCard', () => {
  const campaign = (status: FlashSaleCampaign['status']): FlashSaleCampaign => ({
    id: 'c1',
    name: 'Giờ vàng',
    status,
    starts_at: '2026-11-11T10:00:00Z',
    ends_at: '2099-11-11T12:00:00Z',
    items: [
      {
        id: 'f1',
        sku_id: 's1',
        sku_code: 'AO-S',
        product_name: 'Áo thun',
        regular_price: 150000,
        sale_price: 99000,
        quota_total: 10,
        purchase_limit: 2,
      },
    ],
  })
  const render = (status: FlashSaleCampaign['status']) =>
    renderToStaticMarkup(
      <MemoryRouter>
        <FlashSaleCampaignCard campaign={campaign(status)} addingSkuId={null} onAdd={noop} />
      </MemoryRouter>,
    )

  it('labels ended campaigns as ended and offers no add button', () => {
    const markup = render('ended')
    expect(markup).toContain('ĐÃ KẾT THÚC')
    expect(markup).not.toContain('SẮP DIỄN RA')
    expect(markup).not.toContain('Thêm vào giỏ')
  })

  it('shows product name, SKU, struck regular price, sale price, limit and the add button when active', () => {
    const markup = render('active')
    expect(markup).toContain('ĐANG DIỄN RA')
    expect(markup).toContain('Áo thun')
    expect(markup).toContain('SKU: AO-S')
    expect(markup).toContain('line-through')
    expect(markup).toContain('150.000')
    expect(markup).toContain('99.000')
    expect(markup).toContain('2 cái/khách')
    expect(markup).toContain('Thêm vào giỏ')
    expect(markup).toContain('Thời gian chính thức theo hệ thống')
  })

  it('shows upcoming campaigns without an add button', () => {
    const markup = render('scheduled')
    expect(markup).toContain('SẮP DIỄN RA')
    expect(markup).not.toContain('Thêm vào giỏ')
  })
})
