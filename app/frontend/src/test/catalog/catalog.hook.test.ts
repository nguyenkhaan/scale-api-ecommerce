import { describe, expect, it } from 'bun:test'

import { buildSkuOptions, catalogQueryKeys, describeProductsLoadError, productListParams, mergeCampaignSkuOptions } from '@/features/catalog/catalog.hook'
import type { CatalogProduct } from '@/features/catalog/catalog.type'

const product = (over: Partial<CatalogProduct>): CatalogProduct => ({
  id: 'p',
  name: 'Áo thun',
  active: true,
  image_url: null,
  skus: [],
  ...over,
})

const sku = (id: string, code: string, active: boolean, price = 100000) => ({
  id,
  code,
  attributes: {},
  regular_price: price,
  active,
})

describe('buildSkuOptions', () => {
  it('keeps only active SKUs of active products, labelled "name — code — price"', () => {
    const options = buildSkuOptions([
      product({ skus: [sku('s1', 'AO-S', true), sku('s2', 'AO-M', false)] }),
      product({ id: 'p2', name: 'Quần', active: false, skus: [sku('s3', 'Q-1', true)] }),
    ])

    expect(options).toHaveLength(1)
    expect(options[0].skuId).toBe('s1')
    expect(options[0].label).toStartWith('Áo thun — AO-S — ')
    expect(options[0].label).toContain('100.000')
  })
})

describe('catalog query state', () => {
  it('separates empty and non-empty search terms in the query key', () => {
    expect(catalogQueryKeys.products(20, '')).not.toEqual(catalogQueryKeys.products(20, 'áo'))
  })

  it('passes the search term and resets the cursor for a new query', () => {
    expect(productListParams('áo', undefined)).toEqual({ limit: 20, query: 'áo' })
    expect(productListParams('  ', 'next-cursor')).toEqual({ limit: 20, cursor: 'next-cursor' })
  })

  it('does not tell public catalog users to log in for a generic 401', () => {
    expect(describeProductsLoadError({ status: 401 })).toBe('Không thể tải danh sách sản phẩm. Vui lòng thử lại.')
  })
})

describe('mergeCampaignSkuOptions', () => {
  it('adds campaign SKUs missing from active options so editing keeps them selectable', () => {
    const merged = mergeCampaignSkuOptions(
      [{ skuId: 's1', regularPrice: 1, label: 'x' }],
      [
        { sku_id: 's1', sku_code: 'A', product_name: 'P', regular_price: 1 },
        { sku_id: 's9', sku_code: 'Z', product_name: 'Old', regular_price: 5000 },
      ],
    )

    expect(merged.map((option) => option.skuId)).toEqual(['s1', 's9'])
    expect(merged[1].label).toStartWith('Old — Z — ')
  })
})
