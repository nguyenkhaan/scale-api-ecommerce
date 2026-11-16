import { useId, useState } from 'react'

import { activeSkus } from '@/features/catalog/catalog.hook'
import type { CatalogProduct } from '@/features/catalog/catalog.type'
import { formatVND } from '@/lib/format'

type Props = {
  product: CatalogProduct
  busy: boolean
  onAdd: (skuId: string, amount: number, productName: string) => void
}

const maxAmount = 99

export function ProductCard({ product, busy, onAdd }: Props) {
  const skus = activeSkus(product)
  const [skuId, setSkuId] = useState(skus[0]?.id ?? '')
  const [amount, setAmount] = useState(1)
  const uid = useId()
  const selected = skus.find((sku) => sku.id === skuId)
  const validAmount = Number.isInteger(amount) && amount >= 1 && amount <= maxAmount

  return (
    <li className="flex flex-col justify-between rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div>
        <h2 className="break-words text-base font-bold text-slate-900">{product.name}</h2>
        {selected ? (
          <p className="mt-1 text-xl font-bold text-slate-900">{formatVND(selected.regular_price)}</p>
        ) : (
          <p className="mt-1 text-sm text-slate-500">Hiện chưa có SKU nào đang bán.</p>
        )}
      </div>

      {selected && (
        <div className="mt-4 space-y-3">
          <div>
            <label htmlFor={`${uid}-sku`} className="block text-xs font-semibold text-slate-700">
              Phiên bản (SKU)
            </label>
            <select
              id={`${uid}-sku`}
              value={skuId}
              onChange={(event) => setSkuId(event.target.value)}
              className="mt-1 min-h-11 w-full rounded border border-slate-300 bg-white px-2 text-sm"
            >
              {skus.map((sku) => (
                <option key={sku.id} value={sku.id}>
                  {sku.code} — {formatVND(sku.regular_price)}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor={`${uid}-qty`} className="block text-xs font-semibold text-slate-700">
              Số lượng
            </label>
            <input
              id={`${uid}-qty`}
              type="number"
              min={1}
              max={maxAmount}
              value={amount}
              onChange={(event) => setAmount(Number(event.target.value))}
              aria-invalid={!validAmount}
              className="mt-1 min-h-11 w-24 rounded border border-slate-300 px-2 text-sm"
            />
          </div>
          <button
            type="button"
            disabled={busy || !validAmount}
            onClick={() => onAdd(selected.id, amount, product.name)}
            className="min-h-11 w-full rounded bg-slate-900 px-3 py-2 text-sm font-bold text-white hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:cursor-not-allowed disabled:bg-slate-300"
          >
            {busy ? 'Đang thêm...' : 'Thêm vào giỏ'}
          </button>
        </div>
      )}
    </li>
  )
}
