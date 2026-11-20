import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import { AdminPage } from '@/features/admin-flashsale/pages/AdminPage'
import { AdminFlashSalePageView, type AdminFlashSalePageViewProps } from '@/features/admin-flashsale/pages/AdminFlashSalePage'
import { AdminCampaignCard } from '@/features/admin-flashsale/components/AdminCampaignCard'
import { CampaignFormModal } from '@/features/admin-flashsale/components/CampaignFormModal'
import { FieldErrors } from '@/features/admin-flashsale/components/FieldErrors'
import type { FlashSaleCampaign } from '@/features/flashsale/flashsale.type'

const campaign = (status: FlashSaleCampaign['status']): FlashSaleCampaign => ({
  id: `c-${status}`,
  name: 'Flash Sale 11.11',
  status,
  starts_at: '2026-11-11T10:00:00Z',
  ends_at: '2026-11-11T12:00:00Z',
  items: [
    {
      id: `i-${status}`,
      sku_id: 's1',
      sku_code: 'AO-S',
      product_name: 'Áo thun',
      regular_price: 199000,
      sale_price: 129000,
      quota_total: 10,
      purchase_limit: 2,
    },
  ],
})

const noop = () => {}

function renderCard(status: FlashSaleCampaign['status']) {
  return renderToStaticMarkup(
    <AdminCampaignCard campaign={campaign(status)} busy={false} onAction={noop} />,
  )
}

describe('AdminCampaignCard', () => {
  it('exposes only valid actions for each campaign state', () => {
    const draft = renderCard('draft')
    const scheduled = renderCard('scheduled')
    const active = renderCard('active')
    const ended = renderCard('ended')

    expect(draft).toContain('Chỉnh sửa')
    expect(draft).toContain('Lên lịch')
    expect(draft).not.toContain('Dừng chiến dịch')
    expect(scheduled).toContain('Hủy lịch')
    expect(scheduled).not.toContain('Chỉnh sửa')
    expect(active).toContain('Dừng chiến dịch')
    expect(active).not.toContain('Lên lịch')
    expect(ended).toContain('Chỉ xem')
  })

  it('uses semantic admin status and price markup without emoji icons', () => {
    const markup = renderCard('active')
    expect(markup).toContain('data-status-domain="flashSale"')
    expect(markup).toContain('<del')
    expect(markup).not.toContain('⚡')
  })
})

describe('CampaignFormModal', () => {
  it('uses dialog semantics, shared fields, and a disabled-safe submit state', () => {
    const markup = renderToStaticMarkup(
      <CampaignFormModal
        skuOptions={[{ skuId: 's1', regularPrice: 199000, label: 'Áo thun — AO-S — 199.000 ₫' }]}
        skuLoadError={null}
        onSubmit={async () => {}}
        onClose={noop}
      />,
    )

    expect(markup).toContain('role="dialog"')
    expect(markup).toContain('aria-modal="true"')
    expect(markup).toContain('aria-label="Đóng"')
    expect(markup).toContain('Tên chiến dịch')
    expect(markup).toContain('Tạo bản nháp')
    expect(markup).not.toContain('⚡')
  })

  it('renders field errors as an alert region', () => {
    const markup = renderToStaticMarkup(<FieldErrors id="name-errors" messages={['Tên chiến dịch là bắt buộc.']} />)
    expect(markup).toContain('role="alert"')
    expect(markup).toContain('Tên chiến dịch là bắt buộc.')
  })
})

describe('AdminFlashSalePageView', () => {
  const baseProps: Omit<AdminFlashSalePageViewProps, 'list'> = {
    busyId: null,
    notice: null,
    onCreate: noop,
    onAction: noop,
    onLoadMore: noop,
  }

  it('renders empty and retryable error states without dead navigation links', () => {
    const empty = renderToStaticMarkup(
      <MemoryRouter><AdminFlashSalePageView {...baseProps} list={{ items: [], loading: false, loadingMore: false, error: null, hasMore: false, reload: async () => {}, loadMore: async () => {} }} /></MemoryRouter>,
    )
    const error = renderToStaticMarkup(
      <MemoryRouter><AdminFlashSalePageView {...baseProps} list={{ items: [], loading: false, loadingMore: false, error: 'Lỗi API', hasMore: false, reload: async () => {}, loadMore: async () => {} }} /></MemoryRouter>,
    )

    expect(empty).toContain('Chưa có chiến dịch Flash Sale nào.')
    expect(empty).not.toContain('Tồn kho')
    expect(empty).not.toContain('Sản phẩm')
    expect(error).toContain('role="alert"')
    expect(error).toContain('Thử lại')
  })

  it('keeps the admin landing page free of emoji and unimplemented navigation', () => {
    const markup = renderToStaticMarkup(<MemoryRouter><AdminPage /></MemoryRouter>)
    expect(markup).not.toContain('⚡')
    expect(markup).not.toContain('Tồn kho')
    expect(markup).not.toContain('Quản lý sản phẩm')
  })

  it('renders a recovery action for failed admin mutations', () => {
    const markup = renderToStaticMarkup(
      <MemoryRouter>
        <AdminFlashSalePageView
          {...baseProps}
          notice={{ kind: 'error', text: 'Phiên đăng nhập đã hết hạn.', recovery: 'login' }}
          onNoticeAction={noop}
          list={{ items: [], loading: false, loadingMore: false, error: null, hasMore: false, reload: async () => {}, loadMore: async () => {} }}
        />
      </MemoryRouter>,
    )

    expect(markup).toContain('role="alert"')
    expect(markup).toContain('Đăng nhập lại')
  })
})
