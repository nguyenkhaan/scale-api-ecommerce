import { describe, expect, it } from 'bun:test'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import { ProductsPage } from '@/features/catalog/pages/ProductsPage'
import { catalogQueryKeys } from '@/features/catalog/catalog.hook'
import type { CatalogProduct } from '@/features/catalog/catalog.type'

const product: CatalogProduct = {
  id: 'p1',
  name: 'Áo thun SaleScale',
  description: 'Áo cotton',
  image_url: null,
  active: true,
  skus: [
    { id: 's1', code: 'AO-S', attributes: {}, regular_price: 199000, active: true },
  ],
}

function renderPage(query: string, items: CatalogProduct[]) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  queryClient.setQueryData(catalogQueryKeys.products(20, query), {
    pages: [{ items, next_cursor: null }],
    pageParams: [undefined],
  })
  return renderToStaticMarkup(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[query ? `/products?query=${encodeURIComponent(query)}` : '/products']}>
        <ProductsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('ProductsPage', () => {
  it('renders search, product cards and add-to-cart action for anonymous catalog browsing', () => {
    const markup = renderPage('Áo', [product])

    expect(markup).toContain('role="search"')
    expect(markup).toContain('Áo thun SaleScale')
    expect(markup).toContain('Thêm vào giỏ')
    expect(markup).toContain('Xem giỏ hàng')
  })

  it('renders a recoverable empty state', () => {
    const markup = renderPage('', [])

    expect(markup).toContain('Chưa có sản phẩm nào.')
    expect(markup).toContain('role="status"')
  })
})
