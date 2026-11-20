import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import { FlashSaleCampaignCard } from '@/features/flashsale/components/FlashSaleCampaignCard'
import { Countdown } from '@/features/flashsale/components/Countdown'
import { FlashSalesPageView, type FlashSalesPageViewProps } from '@/features/flashsale/pages/FlashSalesPage'
import type { FlashSaleCampaign } from '@/features/flashsale/flashsale.type'

const campaign = (status: FlashSaleCampaign['status']): FlashSaleCampaign => ({
  id: `campaign-${status}`,
  name: 'Flash Sale 11.11',
  status,
  starts_at: '2099-11-11T10:00:00Z',
  ends_at: '2099-11-11T12:00:00Z',
  items: [
    {
      id: `item-${status}`,
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

const emptySales = {
  items: [],
  loading: false,
  loadingMore: false,
  error: null,
  hasMore: false,
  reload: async () => {},
  loadMore: async () => {},
}

const baseProps: Omit<FlashSalesPageViewProps, 'sales'> = {
  cartAction: { notice: null, busyKey: null, add: async () => true },
}

function renderCard(status: FlashSaleCampaign['status']) {
  return renderToStaticMarkup(
    <MemoryRouter>
      <FlashSaleCampaignCard campaign={campaign(status)} addingSkuId={null} onAdd={() => {}} />
    </MemoryRouter>,
  )
}

describe('FlashSaleCampaignCard', () => {
  it('renders active status, price hierarchy, timer and add CTA', () => {
    const markup = renderCard('active')

    expect(markup).toContain('data-status-value="active"')
    expect(markup).toContain('129.000')
    expect(markup).toContain('199.000')
    expect(markup).toContain('Giới hạn mua:')
    expect(markup).toContain('role="timer"')
    expect(markup).toContain('aria-live="off"')
    expect(markup).toContain('Thêm vào giỏ')
  })

  it('does not expose add actions for scheduled or ended campaigns', () => {
    expect(renderCard('scheduled')).not.toContain('Thêm vào giỏ')
    expect(renderCard('ended')).not.toContain('Thêm vào giỏ')
  })
})

describe('FlashSalesPageView', () => {
  it('renders loading, empty and retryable error states with common feedback', () => {
    const loading = renderToStaticMarkup(
      <MemoryRouter><FlashSalesPageView {...baseProps} sales={{ ...emptySales, loading: true }} /></MemoryRouter>,
    )
    const empty = renderToStaticMarkup(<MemoryRouter><FlashSalesPageView {...baseProps} sales={emptySales} /></MemoryRouter>)
    const error = renderToStaticMarkup(
      <MemoryRouter><FlashSalesPageView {...baseProps} sales={{ ...emptySales, error: 'Không thể tải flash sale.' }} /></MemoryRouter>,
    )

    expect(loading).toContain('role="status"')
    expect(empty).toContain('Hiện tại không có chương trình Flash Sale nào.')
    expect(error).toContain('role="alert"')
    expect(error).toContain('Thử lại')
  })
})

describe('Countdown', () => {
  it('keeps timer updates out of the screen-reader announcement queue', () => {
    const markup = renderToStaticMarkup(
      <Countdown status="active" startsAt="2099-11-11T10:00:00Z" endsAt="2099-11-11T12:00:00Z" />,
    )

    expect(markup).toContain('role="timer"')
    expect(markup).toContain('aria-live="off"')
  })
})
