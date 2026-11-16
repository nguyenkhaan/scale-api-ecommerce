import { usePagedList } from '@/lib/use-paged-list'

import { flashSaleClient } from './flashsale.service'

const listPublicFlashSales = (cursor?: string) => flashSaleClient.listPublicFlashSales({ cursor })

export function useFlashSales() {
  return usePagedList(listPublicFlashSales, 'chương trình flash sale')
}
