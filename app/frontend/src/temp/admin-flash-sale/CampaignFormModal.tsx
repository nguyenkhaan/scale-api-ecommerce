import { useState, type FormEvent } from 'react'
import { mergeCampaignSkuOptions, type SkuOption } from '../catalog/product-skus'
import { isApiError, type FieldViolation } from '../lib/api-error'
import type { CreateFlashSaleRequest, FlashSaleCampaign } from '../lib/flash-sale-api'
import { fieldMessages, unclaimedMessages } from '../lib/violation-messages'
import { describeAdminFailure } from './admin-actions'
import {
  buildRequest,
  emptyForm,
  emptyItem,
  formFromCampaign,
  isClaimedField,
  type CampaignForm,
  type FormItem,
} from './campaign-form'
import { CampaignItemRow } from './CampaignItemRow'
import { FieldErrors, inputClass } from './FieldErrors'

type Props = {
  /** Present when editing a draft; absent when creating. */
  campaign?: FlashSaleCampaign
  skuOptions: SkuOption[]
  skuLoadError: string | null
  onSubmit: (request: CreateFlashSaleRequest) => Promise<void>
  onClose: () => void
}

export function CampaignFormModal({ campaign, skuOptions, skuLoadError, onSubmit, onClose }: Props) {
  const [form, setForm] = useState<CampaignForm>(() => (campaign ? formFromCampaign(campaign) : emptyForm()))
  const [violations, setViolations] = useState<FieldViolation[]>([])
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const options = mergeCampaignSkuOptions(skuOptions, campaign?.items ?? [])
  const regularPriceOf = (skuId: string) => options.find((o) => o.skuId === skuId)?.regularPrice
  const patch = (changes: Partial<CampaignForm>) => setForm((prev) => ({ ...prev, ...changes }))
  const patchItem = (index: number, field: keyof FormItem, value: string) =>
    setForm((prev) => ({
      ...prev,
      items: prev.items.map((item, i) => (i === index ? { ...item, [field]: value } : item)),
    }))

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setFormError(null)
    const built = buildRequest(form, regularPriceOf)
    if (!built.ok) {
      setViolations(built.violations)
      return
    }
    setViolations([])
    setSubmitting(true)
    try {
      await onSubmit(built.request)
    } catch (err) {
      if (isApiError(err) && err.code === 'VALIDATION_ERROR' && err.violations.length > 0) {
        setViolations(err.violations)
      } else {
        setFormError(describeAdminFailure(err).text)
      }
    } finally {
      setSubmitting(false)
    }
  }

  const summary = unclaimedMessages(violations, (field) => isClaimedField(field, form.items.length))
  const textField = (field: 'name' | 'starts_at' | 'ends_at') => fieldMessages(violations, field)
  const title = campaign ? 'Chỉnh sửa chiến dịch (bản nháp)' : 'Tạo chiến dịch Flash Sale mới'

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="campaign-form-title"
        onKeyDown={(e) => e.key === 'Escape' && onClose()}
        className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-lg bg-white p-4 shadow-xl sm:p-6"
      >
        <h2 id="campaign-form-title" className="text-xl font-bold text-slate-900">{title}</h2>
        <p className="mt-1 text-xs text-slate-500">
          Chiến dịch sẽ ở trạng thái bản nháp (draft). Bạn có thể kiểm tra trước khi lên lịch.
        </p>

        {(formError || summary.length > 0 || skuLoadError) && (
          <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-3 text-xs text-rose-800">
            {[formError, skuLoadError, ...summary].filter(Boolean).map((text) => (
              <p key={text}>{text}</p>
            ))}
          </div>
        )}

        <form onSubmit={(e) => void handleSubmit(e)} noValidate className="mt-4 space-y-4">
          <div>
            <label htmlFor="campaign-name" className="block text-xs font-semibold text-slate-700">Tên chiến dịch</label>
            <input
              id="campaign-name"
              type="text"
              autoFocus
              value={form.name}
              onChange={(e) => patch({ name: e.target.value })}
              placeholder="Ví dụ: Giờ vàng 11.11"
              aria-invalid={textField('name').length > 0}
              aria-describedby="campaign-name-error"
              className={inputClass}
            />
            <FieldErrors id="campaign-name-error" messages={textField('name')} />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {(
              [
                ['starts_at', 'campaign-starts', 'Bắt đầu', form.startsAt, 'startsAt'],
                ['ends_at', 'campaign-ends', 'Kết thúc', form.endsAt, 'endsAt'],
              ] as const
            ).map(([field, id, label, value, key]) => (
              <div key={id}>
                <label htmlFor={id} className="block text-xs font-semibold text-slate-700">{label}</label>
                <input
                  id={id}
                  type="datetime-local"
                  value={value}
                  onChange={(e) => patch({ [key]: e.target.value })}
                  aria-invalid={textField(field).length > 0}
                  aria-describedby={`${id}-error`}
                  className={inputClass}
                />
                <FieldErrors id={`${id}-error`} messages={textField(field)} />
              </div>
            ))}
          </div>

          <div>
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-semibold text-slate-700">Danh sách SKU và giá ưu đãi</h3>
              <button
                type="button"
                onClick={() => patch({ items: [...form.items, emptyItem()] })}
                className="min-h-9 text-xs font-semibold text-amber-700 hover:text-amber-800 focus-visible:outline-2 focus-visible:outline-amber-500"
              >
                + Thêm dòng SKU
              </button>
            </div>
            <FieldErrors id="campaign-items-error" messages={fieldMessages(violations, 'items')} />
            <div className="mt-2 space-y-3">
              {form.items.map((item, index) => (
                <CampaignItemRow
                  key={index}
                  index={index}
                  item={item}
                  options={options}
                  violations={violations}
                  canRemove={form.items.length > 1}
                  onChange={(field, value) => patchItem(index, field, value)}
                  onRemove={() => patch({ items: form.items.filter((_, i) => i !== index) })}
                />
              ))}
            </div>
          </div>

          <div className="mt-6 flex justify-end gap-3 border-t border-slate-200 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="min-h-11 rounded border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-amber-500"
            >
              Đóng
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="min-h-11 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-50"
            >
              {submitting ? 'Đang lưu...' : campaign ? 'Lưu thay đổi' : 'Tạo bản nháp'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
