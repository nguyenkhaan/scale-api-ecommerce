import { useState } from 'react'
import { cartClient, type CheckoutPreview } from '../lib/cart-api'
import { describeCartFailure } from './cart-messages'
import { detectPreviewChanges, selectionSignature, type PreviewChange } from './preview-changes'
import type { CartState } from './use-cart'

type ShownPreview = {
  data: CheckoutPreview
  signature: string
  changes: PreviewChange[]
  acknowledged: boolean
}

/**
 * Preview lifecycle. A preview is only "current" while the selected lines and their versions are unchanged,
 * so any cart edit invalidates it without extra bookkeeping.
 */
export function useCheckoutPreview(cart: CartState) {
  const [shown, setShown] = useState<ShownPreview | null>(null)
  // Last successful preview, kept after invalidation so a later preview can reveal price changes.
  const [lastData, setLastData] = useState<CheckoutPreview | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const signature = selectionSignature(cart.selectedItems)
  const current = shown && shown.signature === signature ? shown : null

  const run = async () => {
    if (cart.selectedItems.length === 0) return
    setLoading(true)
    setError(null)
    cart.applyLineErrors([])
    try {
      const data = await cartClient.previewCheckout({
        items: cart.selectedItems.map((item) => ({ cart_item_id: item.id, version: item.version })),
      })
      const changes = detectPreviewChanges(lastData, data)
      setShown({ data, signature, changes, acknowledged: changes.length === 0 })
      setLastData(data)
    } catch (err) {
      setShown(null)
      const failure = describeCartFailure(err)
      setError(failure.text)
      if (failure.lineErrors.length > 0) cart.applyLineErrors(failure.lineErrors)
      if (failure.cartChanged) void cart.reload()
    } finally {
      setLoading(false)
    }
  }

  const acknowledge = () => setShown((prev) => (prev ? { ...prev, acknowledged: true } : prev))

  return {
    preview: current?.data ?? null,
    changes: current?.changes ?? [],
    acknowledged: current?.acknowledged ?? false,
    loading,
    error,
    run,
    acknowledge,
  }
}
