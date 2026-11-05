import { useCallback, useEffect, useState } from 'react'
import { cartClient, type Cart, type CartItem } from '../lib/cart-api'
import type { LineError } from '../lib/api-error'
import { describeCartFailure, lineErrorMessage } from './cart-messages'

export type CartNotice = { kind: 'success' | 'error'; text: string }
export type LineProblem = { message: string; code?: string }

/** Cart state, line mutations and per-line problems for the cart page. */
export function useCart() {
  const [cart, setCart] = useState<Cart | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [busySku, setBusySku] = useState<string | null>(null)
  const [notice, setNotice] = useState<CartNotice | null>(null)
  const [lineProblems, setLineProblems] = useState<Record<string, LineProblem>>({})
  // Tracking deselected lines (instead of selected) auto-selects lines that appear after a reload.
  const [deselected, setDeselected] = useState<Set<string>>(new Set())

  const load = useCallback(
    () =>
      cartClient
        .getCart()
        .then((data) => {
          setCart(data)
          setLineProblems({})
        })
        .catch((err: unknown) => setLoadError(describeCartFailure(err).text))
        .finally(() => setLoading(false)),
    [],
  )

  useEffect(() => {
    void load()
  }, [load])

  const reload = useCallback(() => {
    setLoadError(null)
    return load()
  }, [load])

  const fail = useCallback(
    (err: unknown, skuId?: string) => {
      const failure = describeCartFailure(err)
      if (failure.cartChanged) {
        setNotice({ kind: 'error', text: failure.text })
        void reload()
      } else if (skuId && !failure.unauthenticated) {
        setLineProblems((prev) => ({ ...prev, [skuId]: { message: failure.text, code: failure.lineCode } }))
      } else {
        setNotice({ kind: 'error', text: failure.text })
      }
    },
    [reload],
  )

  const mutate = useCallback(
    async (skuId: string, action: () => Promise<void>) => {
      setBusySku(skuId)
      setNotice(null)
      try {
        await action()
        setLineProblems((prev) => {
          const { [skuId]: _cleared, ...rest } = prev
          return rest
        })
      } catch (err) {
        fail(err, skuId)
      } finally {
        setBusySku(null)
      }
    },
    [fail],
  )

  const changeQuantity = (item: CartItem, quantity: number) =>
    quantity < 1
      ? Promise.resolve()
      : mutate(item.sku_id, async () => {
          setCart(
            await cartClient.upsertCartItem(item.sku_id, {
              quantity,
              price_source: item.price_source,
              flash_sale_item_id: item.flash_sale_item_id,
              expected_version: item.version,
            }),
          )
        })

  const remove = (item: CartItem) =>
    mutate(item.sku_id, async () => {
      await cartClient.removeCartItem(item.sku_id)
      setCart(await cartClient.getCart())
    })

  /** Explicit, customer-initiated fallback; never done automatically. */
  const switchToRegular = (item: CartItem) =>
    mutate(item.sku_id, async () => {
      setCart(
        await cartClient.upsertCartItem(item.sku_id, {
          quantity: item.quantity,
          price_source: 'regular',
          expected_version: item.version,
        }),
      )
      setNotice({ kind: 'success', text: `Đã chuyển "${item.product_name}" sang giá thường.` })
    })

  const applyLineErrors = useCallback((errors: LineError[]) => {
    setLineProblems(
      Object.fromEntries(errors.map((e) => [e.sku_id, { message: lineErrorMessage(e), code: e.code }])),
    )
  }, [])

  const items = cart?.items ?? []
  const selectedItems = items.filter((item) => !deselected.has(item.id))

  const toggle = (itemId: string) =>
    setDeselected((prev) => {
      const next = new Set(prev)
      if (!next.delete(itemId)) next.add(itemId)
      return next
    })

  const toggleAll = () =>
    setDeselected(selectedItems.length === items.length ? new Set(items.map((item) => item.id)) : new Set())

  return {
    cart,
    items,
    loading,
    loadError,
    busySku,
    notice,
    setNotice,
    lineProblems,
    selectedItems,
    reload,
    changeQuantity,
    remove,
    switchToRegular,
    applyLineErrors,
    toggle,
    toggleAll,
  }
}

export type CartState = ReturnType<typeof useCart>
