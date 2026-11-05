import { useEffect, useState } from 'react'
import { buildSkuOptions, type SkuOption } from '../catalog/product-skus'
import { catalogClient, type CatalogProduct } from '../lib/catalog-api'
import { describeLoadError } from '../lib/load-error'

const PAGE_SIZE = 100
const MAX_PAGES = 10

/** Loads the admin catalog (following cursors up to a cap) and builds the SKU picker options. */
export function useAdminSkuOptions() {
  const [options, setOptions] = useState<SkuOption[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      const products: CatalogProduct[] = []
      let cursor: string | undefined
      try {
        for (let page = 0; page < MAX_PAGES; page++) {
          const result = await catalogClient.listAdminProducts({ limit: PAGE_SIZE, cursor })
          products.push(...result.items)
          if (!result.next_cursor) break
          cursor = result.next_cursor
        }
        if (!cancelled) setOptions(buildSkuOptions(products))
      } catch (err) {
        if (!cancelled) setError(describeLoadError(err, 'danh sách SKU'))
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  return { options, error }
}
