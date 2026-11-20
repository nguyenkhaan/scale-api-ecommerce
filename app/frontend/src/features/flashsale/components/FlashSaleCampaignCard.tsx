import type { FlashSaleCampaign } from '../flashsale.service'
import { Button, PriceBlock, StatusBadge, Surface } from '@/components/common'
import { formatDateTime } from '@/lib/format'
import { Countdown } from './Countdown'

type Props = {
  campaign: FlashSaleCampaign
  addingSkuId: string | null
  onAdd: (skuId: string, flashSaleItemId: string, productName: string) => void
}

export function FlashSaleCampaignCard({ campaign, addingSkuId, onAdd }: Props) {
  const isActive = campaign.status === 'active'
  return (
    <Surface
      as="section"
      className="overflow-hidden"
      aria-labelledby={`campaign-title-${campaign.id}`}
    >
      <div className="border-b border-[var(--ss-color-border)] bg-[var(--ss-color-app-bg)] px-4 py-4 sm:px-6">
        <div className="flex flex-wrap items-center gap-3">
          <h2 id={`campaign-title-${campaign.id}`} className="text-lg font-semibold text-[var(--ss-color-text)]">
            {campaign.name}
          </h2>
          <StatusBadge domain="flashSale" value={campaign.status} audience="customer" />
        </div>
        <p className="mt-1 text-xs text-[var(--ss-color-muted)]">
          Thời gian: {formatDateTime(campaign.starts_at)} — {formatDateTime(campaign.ends_at)}
        </p>
        <Countdown status={campaign.status} startsAt={campaign.starts_at} endsAt={campaign.ends_at} />
      </div>

      <div className="p-4 sm:p-6">
        <h3 className="text-sm font-semibold text-[var(--ss-color-muted)]">
          Sản phẩm ưu đãi ({campaign.items.length})
        </h3>
        <ul className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {campaign.items.map((item) => {
            const adding = addingSkuId === item.sku_id
            return (
              <li key={item.id}>
                <Surface as="article" className="flex h-full flex-col justify-between p-4">
                <div>
                  <p className="break-words font-semibold text-[var(--ss-color-text)]">{item.product_name}</p>
                  <p className="font-mono text-xs text-[var(--ss-color-muted)]">SKU: {item.sku_code}</p>
                  <div className="mt-3"><PriceBlock regularPrice={item.regular_price} effectivePrice={item.sale_price} source="flash_sale" endsAt={campaign.ends_at} compact /></div>
                  <p className="mt-2 text-xs text-[var(--ss-color-muted)]">
                    Giới hạn mua: <span className="font-semibold">{item.purchase_limit} cái/khách</span>
                  </p>
                </div>
                {isActive && (
                  <div className="mt-4 border-t border-[var(--ss-color-border)] pt-3">
                    <Button loading={adding} onClick={() => onAdd(item.sku_id, item.id, item.product_name)} className="w-full">
                      Thêm vào giỏ
                    </Button>
                  </div>
                )}
                </Surface>
              </li>
            )
          })}
        </ul>
      </div>
    </Surface>
  )
}
