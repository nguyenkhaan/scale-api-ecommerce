import { Button, PriceBlock, StatusBadge, Surface } from '@/components/common'
import type { FlashSaleCampaign } from '@/features/flashsale/flashsale.type'
import { formatDateTime } from '@/lib/format'
import { adminActionLabels, adminActionsFor, type AdminAction } from '../admin-flashsale.logic'

type Props = {
  campaign: FlashSaleCampaign
  busy: boolean
  onAction: (action: AdminAction) => void
}

const actionVariant: Record<AdminAction, 'primary' | 'secondary' | 'danger' | 'ghost'> = {
  edit: 'secondary',
  schedule: 'primary',
  cancel_schedule: 'secondary',
  stop: 'danger',
}

export function AdminCampaignCard({ campaign, busy, onAction }: Props) {
  const actions = adminActionsFor(campaign.status)
  return (
    <li>
      <Surface as="article" className="p-4 sm:p-5">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-start">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <h2 className="text-lg font-semibold text-[var(--ss-color-text)]">{campaign.name}</h2>
            <StatusBadge domain="flashSale" value={campaign.status} audience="admin" />
          </div>
          <p className="mt-1 text-xs text-[var(--ss-color-muted)]">
            Thời gian: {formatDateTime(campaign.starts_at)} — {formatDateTime(campaign.ends_at)}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {actions.length === 0 && <span className="text-xs text-[var(--ss-color-muted)]">Chỉ xem</span>}
          {actions.map((action) => (
            <Button
              key={action}
              variant={actionVariant[action]}
              disabled={busy}
              onClick={() => onAction(action)}
            >
              {adminActionLabels[action]}
            </Button>
          ))}
        </div>
      </div>

      <ul className="mt-3 divide-y divide-[var(--ss-color-border)] border-t border-[var(--ss-color-border)] text-xs text-[var(--ss-color-text)]">
        {campaign.items.map((item) => (
          <li key={item.id} className="py-2">
            <span className="font-semibold">{item.product_name}</span>{' '}
            <span className="font-mono text-[var(--ss-color-muted)]">({item.sku_code})</span>
            <span className="mt-1 block">
              <PriceBlock regularPrice={item.regular_price} effectivePrice={item.sale_price} source="flash_sale" compact />
              <span className="mt-1 block text-[var(--ss-color-muted)]">Suất {item.quota_total} · tối đa {item.purchase_limit}/khách</span>
            </span>
          </li>
        ))}
      </ul>
      </Surface>
    </li>
  )
}
