import { useCallback, useState } from 'react'
import { cartClient } from '../lib/cart-api'
import { addToCart, describeAddResult, type AddToCartInput } from './add-to-cart'
import { describeCartFailure } from './cart-messages'

export type CartActionNoticeState = {
  kind: 'success' | 'error'
  text: string
  loginRequired?: boolean
}

/** Shared by the flash sale and catalog pages. `busyKey` is the SKU id currently being added. */
export function useAddToCart() {
  const [busyKey, setBusyKey] = useState<string | null>(null)
  const [notice, setNotice] = useState<CartActionNoticeState | null>(null)

  const add = useCallback(async (input: AddToCartInput, productName: string): Promise<boolean> => {
    setBusyKey(input.skuId)
    setNotice(null)
    try {
      const result = await addToCart(cartClient, input)
      setNotice({ kind: 'success', text: describeAddResult(result, productName) })
      return true
    } catch (err) {
      const failure = describeCartFailure(err)
      setNotice({
        kind: 'error',
        text: failure.cartChanged
          ? 'Giỏ hàng vừa thay đổi ở nơi khác. Vui lòng thử thêm lại.'
          : failure.text,
        loginRequired: failure.unauthenticated,
      })
      return false
    } finally {
      setBusyKey(null)
    }
  }, [])

  return { busyKey, notice, add }
}
