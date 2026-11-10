import { useId, useState } from 'react'
import type { CatalogProduct } from '../lib/catalog-api'
import { formatVND } from '../lib/format'
import { activeSkus } from './product-skus'

type Props = {
  product: CatalogProduct
  busy: boolean
  onAdd: (skuId: string, amount: number, productName: string) => void
}

const MAX_AMOUNT = 99

export function ProductCard({ product, busy, onAdd }: Props) {
  const skus = activeSkus(product)
  const [skuId, setSkuId] = useState(skus[0]?.id ?? '')
  const [amount, setAmount] = useState(1)
  const uid = useId()
  const selected = skus.find((s) => s.id === skuId)
  const validAmount = Number.isInteger(amount) && amount >= 1 && amount <= MAX_AMOUNT

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
              onChange={(e) => setSkuId(e.target.value)}
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
              max={MAX_AMOUNT}
              value={amount}
              onChange={(e) => setAmount(Number(e.target.value))}
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
