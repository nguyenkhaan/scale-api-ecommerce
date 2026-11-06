import type { FieldViolation } from '../lib/api-error'
import type { CreateFlashSaleRequest, FlashSaleCampaign } from '../lib/flash-sale-api'

export type FormItem = { sku_id: string; sale_price: string; quota_total: string; purchase_limit: string }

export type CampaignForm = {
  name: string
  startsAt: string
  endsAt: string
  items: FormItem[]
}

export const emptyItem = (): FormItem => ({ sku_id: '', sale_price: '', quota_total: '10', purchase_limit: '2' })

export const emptyForm = (): CampaignForm => ({ name: '', startsAt: '', endsAt: '', items: [emptyItem()] })

const pad = (n: number) => String(n).padStart(2, '0')

/** ISO instant -> `YYYY-MM-DDTHH:mm` in the browser's zone, as `datetime-local` expects. */
export function toLocalInput(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formFromCampaign(campaign: FlashSaleCampaign): CampaignForm {
  return {
    name: campaign.name,
    startsAt: toLocalInput(campaign.starts_at),
    endsAt: toLocalInput(campaign.ends_at),
    items: campaign.items.map((item) => ({
      sku_id: item.sku_id,
      sale_price: String(item.sale_price),
      quota_total: String(item.quota_total),
      purchase_limit: String(item.purchase_limit),
    })),
  }
}

const positiveInt = (value: string) => /^\d+$/.test(value.trim()) && Number(value) > 0

export type BuildResult =
  | { ok: true; request: CreateFlashSaleRequest }
  | { ok: false; violations: FieldViolation[] }

/**
 * Client-side pre-checks with the same field names the API uses, so inline errors look identical.
 * The server stays authoritative: these only save a round trip for obvious mistakes.
 */
export function buildRequest(form: CampaignForm, regularPriceOf: (skuId: string) => number | undefined): BuildResult {
  const violations: FieldViolation[] = []
  const add = (field: string, reason: string) => violations.push({ field, reason })

  if (form.name.trim() === '') add('name', 'is required')
  const starts = new Date(form.startsAt)
  const ends = new Date(form.endsAt)
  if (form.startsAt === '' || Number.isNaN(starts.getTime())) add('starts_at', 'is required')
  if (form.endsAt === '' || Number.isNaN(ends.getTime())) add('ends_at', 'is required')
  if (violations.every((v) => v.field !== 'starts_at' && v.field !== 'ends_at') && ends <= starts) {
    add('ends_at', 'must be strictly after starts_at')
  }
  if (form.items.length === 0) add('items', 'must have at least one item')

  form.items.forEach((item, i) => {
    const at = (name: string) => `items[${i}].${name}`
    if (item.sku_id === '') add(at('sku_id'), 'is required')
    if (!positiveInt(item.sale_price)) add(at('sale_price'), 'must be greater than 0')
    else {
      const regular = regularPriceOf(item.sku_id)
      if (regular !== undefined && Number(item.sale_price) >= regular) {
        add(at('sale_price'), 'must be less than regular price')
      }
    }
    if (!positiveInt(item.quota_total)) add(at('quota_total'), 'must be greater than 0')
    if (!positiveInt(item.purchase_limit)) add(at('purchase_limit'), 'must be greater than 0')
  })
  if (new Set(form.items.map((i) => i.sku_id).filter(Boolean)).size < form.items.filter((i) => i.sku_id).length) {
    add('items', 'contains duplicate sku_id')
  }

  if (violations.length > 0) return { ok: false, violations }
  return {
    ok: true,
    request: {
      name: form.name.trim(),
      starts_at: starts.toISOString(),
      ends_at: ends.toISOString(),
      items: form.items.map((item) => ({
        sku_id: item.sku_id,
        sale_price: Number(item.sale_price),
        quota_total: Number(item.quota_total),
        purchase_limit: Number(item.purchase_limit),
      })),
    },
  }
}

const itemFields = ['sku_id', 'sale_price', 'quota_total', 'purchase_limit']

/** True when a rendered control owns this violation field (so it is shown inline, not in the summary). */
export function isClaimedField(field: string, itemCount: number): boolean {
  if (['name', 'starts_at', 'ends_at', 'items'].includes(field)) return true
  const m = /^items\[(\d+)\]\.(.+)$/.exec(field)
  return m !== null && Number(m[1]) < itemCount && itemFields.includes(m[2])
}
