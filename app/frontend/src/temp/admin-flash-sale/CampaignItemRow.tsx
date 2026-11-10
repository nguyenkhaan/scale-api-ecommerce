import type { SkuOption } from '../catalog/product-skus'
import type { FieldViolation } from '../lib/api-error'
import { formatVND } from '../lib/format'
import { fieldMessages } from '../lib/violation-messages'
import type { FormItem } from './campaign-form'
import { FieldErrors, inputClass } from './FieldErrors'

type Props = {
  index: number
  item: FormItem
  options: SkuOption[]
  violations: FieldViolation[]
  canRemove: boolean
  onChange: (field: keyof FormItem, value: string) => void
  onRemove: () => void
}

const numericFields: Array<{ key: Exclude<keyof FormItem, 'sku_id'>; label: string }> = [
  { key: 'sale_price', label: 'Giá flash sale (₫)' },
  { key: 'quota_total', label: 'Tổng số suất' },
  { key: 'purchase_limit', label: 'Giới hạn mua / khách' },
]

export function CampaignItemRow({ index, item, options, violations, canRemove, onChange, onRemove }: Props) {
  const idFor = (name: string) => `item-${index}-${name}`
  const errorsFor = (name: string) => fieldMessages(violations, `items[${index}].${name}`)
  const selected = options.find((o) => o.skuId === item.sku_id)

  return (
    <fieldset className="rounded border border-slate-200 bg-slate-50 p-3">
      <legend className="px-1 text-xs font-bold text-slate-700">Dòng {index + 1}</legend>
      <div>
        <label htmlFor={idFor('sku_id')} className="block text-xs font-semibold text-slate-700">
          Sản phẩm — mã SKU — giá thường
        </label>
        <select
          id={idFor('sku_id')}
          value={item.sku_id}
          onChange={(e) => onChange('sku_id', e.target.value)}
          aria-invalid={errorsFor('sku_id').length > 0}
          aria-describedby={idFor('sku_id-error')}
          className={inputClass}
        >
          <option value="">— Chọn SKU —</option>
          {options.map((o) => (
            <option key={o.skuId} value={o.skuId}>
              {o.label}
            </option>
          ))}
        </select>
        <FieldErrors id={idFor('sku_id-error')} messages={errorsFor('sku_id')} />
      </div>

      <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-3">
        {numericFields.map(({ key, label }) => (
          <div key={key}>
            <label htmlFor={idFor(key)} className="block text-xs font-semibold text-slate-700">
              {label}
            </label>
            <input
              id={idFor(key)}
              type="number"
              min={1}
              step={1}
              inputMode="numeric"
              value={item[key]}
              onChange={(e) => onChange(key, e.target.value)}
              aria-invalid={errorsFor(key).length > 0}
              aria-describedby={idFor(`${key}-error`)}
              className={inputClass}
            />
            <FieldErrors id={idFor(`${key}-error`)} messages={errorsFor(key)} />
          </div>
        ))}
      </div>
      {selected && (
        <p className="mt-1 text-[11px] text-slate-500">
          Giá flash sale phải thấp hơn giá thường ({formatVND(selected.regularPrice)}).
        </p>
      )}

      {canRemove && (
        <button
          type="button"
          onClick={onRemove}
          className="mt-2 min-h-9 text-xs font-bold text-rose-700 hover:text-rose-900 focus-visible:outline-2 focus-visible:outline-amber-500"
        >
          Xóa dòng {index + 1}
        </button>
      )}
    </fieldset>
  )
}
