import { saleStatusBadgeClass, saleStatusLabel } from '../flash-sale/flash-sale-status'
import type { FlashSaleCampaign } from '../lib/flash-sale-api'
import { formatDateTime, formatVND } from '../lib/format'
import { adminActionLabels, adminActionsFor, type AdminAction } from './admin-actions'

type Props = {
  campaign: FlashSaleCampaign
  busy: boolean
  onAction: (action: AdminAction) => void
}

const actionClass: Record<AdminAction, string> = {
  edit: 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50',
  schedule: 'bg-blue-700 text-white hover:bg-blue-600',
  cancel_schedule: 'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50',
  stop: 'bg-rose-700 text-white hover:bg-rose-600',
}

export function AdminCampaignCard({ campaign, busy, onAction }: Props) {
  const actions = adminActionsFor(campaign.status)
  return (
    <li className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm sm:p-6">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-start">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <h2 className="text-lg font-bold text-slate-900">{campaign.name}</h2>
            <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-bold ${saleStatusBadgeClass(campaign.status)}`}>
              {saleStatusLabel(campaign.status)}
            </span>
          </div>
          <p className="mt-1 text-xs text-slate-500">
            Thời gian: {formatDateTime(campaign.starts_at)} — {formatDateTime(campaign.ends_at)}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {actions.length === 0 && <span className="text-xs text-slate-500">Chỉ xem</span>}
          {actions.map((action) => (
            <button
              key={action}
              type="button"
              disabled={busy}
              onClick={() => onAction(action)}
              className={`min-h-11 rounded px-3 py-1.5 text-xs font-semibold focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-50 ${actionClass[action]}`}
            >
              {adminActionLabels[action]}
            </button>
          ))}
        </div>
      </div>

      <ul className="mt-3 divide-y divide-slate-100 border-t border-slate-100 text-xs text-slate-700">
        {campaign.items.map((item) => (
          <li key={item.id} className="py-2">
            <span className="font-semibold">{item.product_name}</span>{' '}
            <span className="font-mono text-slate-500">({item.sku_code})</span>
            <span className="block text-slate-600">
              <span className="line-through">{formatVND(item.regular_price)}</span> →{' '}
              <span className="font-semibold text-rose-700">{formatVND(item.sale_price)}</span> · suất {item.quota_total} · tối đa{' '}
              {item.purchase_limit}/khách
            </span>
          </li>
        ))}
      </ul>
    </li>
  )
}
