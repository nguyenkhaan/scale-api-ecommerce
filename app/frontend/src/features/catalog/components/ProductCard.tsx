import { useId, useState } from 'react'

import { Button, PriceBlock, QuantityStepper, Surface } from '@/components/common'
import { Link } from 'react-router-dom'
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
    <li>
      <Surface as="article" className="flex h-full flex-col justify-between p-4">
        <div>
          <h2 className="break-words text-base font-semibold text-[var(--ss-color-text)]"><Link to={`/products/${product.id}`} className="hover:underline">{product.name}</Link></h2>
        {selected ? (
          <div className="mt-2"><PriceBlock regularPrice={selected.regular_price} effectivePrice={selected.regular_price} source="regular" compact /></div>
        ) : (
          <p className="mt-2 text-sm text-[var(--ss-color-muted)]">Hiện chưa có SKU nào đang bán.</p>
        )}
        </div>

        {selected && (
          <div className="mt-4 space-y-3">
            <div>
              <label htmlFor={`${uid}-sku`} className="block text-xs font-semibold text-[var(--ss-color-text)]">
                Phiên bản (SKU)
              </label>
              <select
                id={`${uid}-sku`}
                value={skuId}
                onChange={(event) => setSkuId(event.target.value)}
                className="mt-1 min-h-10 w-full rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] bg-white px-2 text-sm"
              >
                {skus.map((sku) => (
                  <option key={sku.id} value={sku.id}>
                    {sku.code} — {formatVND(sku.regular_price)}
                  </option>
                ))}
              </select>
            </div>
            <QuantityStepper value={amount} min={1} max={maxAmount} busy={busy} label={`Số lượng ${product.name}`} onChange={setAmount} />
            <Button loading={busy} disabled={!validAmount} onClick={() => onAdd(selected.id, amount, product.name)} className="w-full">
              Thêm vào giỏ
            </Button>
          </div>
        )}
      </Surface>
    </li>
  )
}
