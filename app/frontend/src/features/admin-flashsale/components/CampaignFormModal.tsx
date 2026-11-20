import { useState, type FormEvent } from 'react'
import { Alert, Button, Dialog, FormField, StatusBadge } from '@/components/common'
import { mergeCampaignSkuOptions, type SkuOption } from '@/features/catalog/catalog.hook'
import { isApiError, type FieldViolation } from '@/lib/api-error'
import type { CreateFlashSaleRequest, FlashSaleCampaign } from '@/features/flashsale/flashsale.type'
import { fieldMessages, unclaimedMessages } from '@/lib/violation-messages'
import { describeAdminFailure } from '../admin-flashsale.logic'
import {
  buildRequest,
  emptyForm,
  emptyItem,
  formFromCampaign,
  isClaimedField,
  type CampaignForm,
  type FormItem,
} from '../campaign-form'
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

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
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
    <Dialog open title={title} onClose={onClose}>
      <p className="flex flex-wrap items-center gap-2 text-xs text-[var(--ss-color-muted)]">
        Chiến dịch sẽ ở trạng thái bản nháp.
        <StatusBadge domain="flashSale" value="draft" audience="admin" />
      </p>

        {(formError || summary.length > 0 || skuLoadError) && (
          <Alert tone="danger" className="mt-4 text-xs">
            {[formError, skuLoadError, ...summary].filter(Boolean).map((text) => (
              <p key={text}>{text}</p>
            ))}
          </Alert>
        )}

        <form onSubmit={(e) => void handleSubmit(e)} noValidate className="mt-4 space-y-4">
          <FormField id="campaign-name" label="Tên chiến dịch" errorMessages={textField('name')} helper="Tên hiển thị cho quản trị viên và khách hàng.">
            {(field) => (
              <input
                id="campaign-name"
                type="text"
                value={form.name}
                onChange={(e) => patch({ name: e.target.value })}
                placeholder="Ví dụ: Giờ vàng 11.11"
                className={inputClass}
                {...field}
              />
            )}
          </FormField>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {(
              [
                ['starts_at', 'campaign-starts', 'Bắt đầu', form.startsAt, 'startsAt'],
                ['ends_at', 'campaign-ends', 'Kết thúc', form.endsAt, 'endsAt'],
              ] as const
            ).map(([field, id, label, value, key]) => (
              <FormField key={id} id={id} label={label} errorMessages={textField(field)}>
                {(control) => (
                  <input
                    id={id}
                    type="datetime-local"
                    value={value}
                    onChange={(e) => patch({ [key]: e.target.value })}
                    className={inputClass}
                    {...control}
                  />
                )}
              </FormField>
            ))}
          </div>

          <div>
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-semibold text-[var(--ss-color-text)]">Danh sách SKU và giá ưu đãi</h3>
              <Button type="button" variant="ghost" onClick={() => patch({ items: [...form.items, emptyItem()] })} className="px-0 text-xs underline">
                + Thêm dòng SKU
              </Button>
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

          <div className="mt-6 flex justify-end gap-3 border-t border-[var(--ss-color-border)] pt-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Đóng
            </Button>
            <Button type="submit" disabled={submitting} loading={submitting}>
              {campaign ? 'Lưu thay đổi' : 'Tạo bản nháp'}
            </Button>
          </div>
        </form>
    </Dialog>
  )
}
