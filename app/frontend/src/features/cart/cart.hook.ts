import { useCallback, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { addToCart, describeAddResult, describeCartFailure, detectPreviewChanges, lineErrorMessage, selectionSignature, type AddToCartInput, type PreviewChange } from './cart.logic'
import { getCart, previewCheckout, removeCartItem, upsertCartItem } from './cart.service'
import type { Cart, CartItem, CheckoutPreview, LineError, UpsertCartItemRequest } from './cart.type'

export const cartQueryKeys = {
  current: ['cart'] as const,
}

export function useCartQuery() {
  return useQuery({
    queryKey: cartQueryKeys.current,
    queryFn: getCart,
    retry: false,
  })
}

export function useUpsertCartItemMutation() {
  return useMutation({
    mutationKey: ['cart', 'upsert'],
    mutationFn: ({ skuId, data }: { skuId: string; data: UpsertCartItemRequest }) => upsertCartItem(skuId, data),
  })
}

export function useRemoveCartItemMutation() {
  return useMutation({
    mutationKey: ['cart', 'remove'],
    mutationFn: removeCartItem,
  })
}

export function useCheckoutPreviewMutation() {
  return useMutation({
    mutationKey: ['cart', 'preview'],
    mutationFn: previewCheckout,
  })
}

export type CartActionNoticeState = {
  kind: 'success' | 'error'
  text: string
  loginRequired?: boolean
}

export function useAddToCart() {
  const queryClient = useQueryClient()
  const [busyKey, setBusyKey] = useState<string | null>(null)
  const [notice, setNotice] = useState<CartActionNoticeState | null>(null)

  const add = useCallback(
    async (input: AddToCartInput, productName: string): Promise<boolean> => {
      setBusyKey(input.skuId)
      setNotice(null)
      try {
        const result = await addToCart({ getCart, upsertCartItem }, input)
        queryClient.setQueryData(cartQueryKeys.current, result.cart)
        setNotice({ kind: 'success', text: describeAddResult(result, productName) })
        return true
      } catch (error) {
        const failure = describeCartFailure(error)
        setNotice({
          kind: 'error',
          text: failure.cartChanged ? 'Giỏ hàng vừa thay đổi ở nơi khác. Vui lòng thử thêm lại.' : failure.text,
          loginRequired: failure.unauthenticated,
        })
        return false
      } finally {
        setBusyKey(null)
      }
    },
    [queryClient],
  )

  return { busyKey, notice, add }
}

export type CartNotice = { kind: 'success' | 'error'; text: string }
export type LineProblem = { message: string; code?: string }

export function useCart() {
  const queryClient = useQueryClient()
  const cartQuery = useCartQuery()
  const upsertMutation = useUpsertCartItemMutation()
  const removeMutation = useRemoveCartItemMutation()
  const [busySku, setBusySku] = useState<string | null>(null)
  const [notice, setNotice] = useState<CartNotice | null>(null)
  const [lineProblems, setLineProblems] = useState<Record<string, LineProblem>>({})
  const [deselected, setDeselected] = useState<Set<string>>(new Set())
  const { refetch } = cartQuery

  const reload = useCallback(async () => {
    const result = await refetch()
    if (result.data !== undefined) setLineProblems({})
  }, [refetch])

  const setCart = useCallback(
    (cart: Cart) => {
      queryClient.setQueryData(cartQueryKeys.current, cart)
    },
    [queryClient],
  )

  const fail = useCallback(
    (error: unknown, skuId?: string) => {
      const failure = describeCartFailure(error)
      if (failure.cartChanged) {
        setNotice({ kind: 'error', text: failure.text })
        void reload()
      } else if (skuId && !failure.unauthenticated) {
        setLineProblems((current) => ({ ...current, [skuId]: { message: failure.text, code: failure.lineCode } }))
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
        setLineProblems((current) => {
          const { [skuId]: cleared, ...rest } = current
          void cleared
          return rest
        })
      } catch (error) {
        fail(error, skuId)
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
            await upsertMutation.mutateAsync({
              skuId: item.sku_id,
              data: {
                quantity,
                price_source: item.price_source,
                flash_sale_item_id: item.flash_sale_item_id,
                expected_version: item.version,
              },
            }),
          )
        })

  const remove = (item: CartItem) =>
    mutate(item.sku_id, async () => {
      await removeMutation.mutateAsync(item.sku_id)
      setCart(await getCart())
    })

  const switchToRegular = (item: CartItem) =>
    mutate(item.sku_id, async () => {
      setCart(
        await upsertMutation.mutateAsync({
          skuId: item.sku_id,
          data: { quantity: item.quantity, price_source: 'regular', expected_version: item.version },
        }),
      )
      setNotice({ kind: 'success', text: `Đã chuyển "${item.product_name}" sang giá thường.` })
    })

  const applyLineErrors = useCallback((errors: LineError[]) => {
    setLineProblems(Object.fromEntries(errors.map((error) => [error.sku_id, { message: lineErrorMessage(error), code: error.code }])))
  }, [])

  const items = cartQuery.data?.items ?? []
  const selectedItems = items.filter((item) => !deselected.has(item.id))

  const toggle = (itemId: string) =>
    setDeselected((current) => {
      const next = new Set(current)
      if (!next.delete(itemId)) next.add(itemId)
      return next
    })

  const toggleAll = () =>
    setDeselected(selectedItems.length === items.length ? new Set(items.map((item) => item.id)) : new Set())

  return {
    cart: cartQuery.data ?? null,
    items,
    loading: cartQuery.isPending,
    loadError: cartQuery.isError ? describeCartFailure(cartQuery.error).text : null,
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

type ShownPreview = {
  data: CheckoutPreview
  signature: string
  changes: PreviewChange[]
  acknowledged: boolean
}

export function useCheckoutPreview(cart: CartState) {
  const previewMutation = useCheckoutPreviewMutation()
  const [shown, setShown] = useState<ShownPreview | null>(null)
  const [lastData, setLastData] = useState<CheckoutPreview | null>(null)
  const [error, setError] = useState<string | null>(null)

  const signature = selectionSignature(cart.selectedItems)
  const current = shown && shown.signature === signature ? shown : null

  const run = async () => {
    if (cart.selectedItems.length === 0) return
    setError(null)
    cart.applyLineErrors([])
    try {
      const data = await previewMutation.mutateAsync({
        items: cart.selectedItems.map((item) => ({ cart_item_id: item.id, version: item.version })),
      })
      const changes = detectPreviewChanges(lastData, data)
      setShown({ data, signature, changes, acknowledged: changes.length === 0 })
      setLastData(data)
    } catch (failure) {
      setShown(null)
      const cartFailure = describeCartFailure(failure)
      setError(cartFailure.text)
      if (cartFailure.lineErrors.length > 0) cart.applyLineErrors(cartFailure.lineErrors)
      if (cartFailure.cartChanged) void cart.reload()
    }
  }

  const acknowledge = () => setShown((current) => (current ? { ...current, acknowledged: true } : current))

  return {
    preview: current?.data ?? null,
    changes: current?.changes ?? [],
    acknowledged: current?.acknowledged ?? false,
    loading: previewMutation.isPending,
    error,
    run,
    acknowledge,
  }
}

export type CheckoutPreviewState = ReturnType<typeof useCheckoutPreview>
