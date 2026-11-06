import { describe, expect, it } from 'bun:test'
import { ApiError } from '../lib/api-error'
import type { Cart, CartItem, UpsertCartItemRequest } from '../lib/cart-api'
import { addToCart, describeAddResult, planAddToCart } from './add-to-cart'

const existing = (over: Partial<CartItem> = {}): CartItem => ({
  id: 'i1',
  sku_id: 's1',
  quantity: 2,
  version: 5,
  price_source: 'flash_sale',
  flash_sale_item_id: 'f1',
  product_name: 'Áo',
  sku_code: 'AO-1',
  reference_unit_price: 100,
  ...over,
})

describe('planAddToCart', () => {
  it('starts at the added amount for a new line without expected_version', () => {
    const plan = planAddToCart(undefined, { skuId: 's1', amount: 1, priceSource: 'flash_sale', flashSaleItemId: 'f1' })
    expect(plan.request).toEqual({ quantity: 1, price_source: 'flash_sale', flash_sale_item_id: 'f1' })
    expect(plan.switchedFrom).toBeNull()
  })

  it('increments an existing line by the amount and guards with expected_version', () => {
    const plan = planAddToCart(existing(), { skuId: 's1', amount: 1, priceSource: 'flash_sale', flashSaleItemId: 'f1' })
    expect(plan.request).toEqual({ quantity: 3, price_source: 'flash_sale', flash_sale_item_id: 'f1', expected_version: 5 })
  })

  it('uses the regular price source without a flash_sale_item_id for catalog adds', () => {
    const plan = planAddToCart(existing({ price_source: 'regular', flash_sale_item_id: null }), { skuId: 's1', amount: 4, priceSource: 'regular' })
    expect(plan.request).toEqual({ quantity: 6, price_source: 'regular', expected_version: 5 })
    expect(plan.switchedFrom).toBeNull()
  })

  it('reports when an existing regular line is switched to flash sale', () => {
    const plan = planAddToCart(existing({ price_source: 'regular', flash_sale_item_id: null }), {
      skuId: 's1',
      amount: 1,
      priceSource: 'flash_sale',
      flashSaleItemId: 'f1',
    })
    expect(plan.switchedFrom).toBe('regular')
  })
})

describe('addToCart', () => {
  const cartWith = (...items: CartItem[]): Cart => ({ id: 'c', items })

  it('reads the cart, then PUTs the incremented absolute quantity', async () => {
    const puts: Array<[string, UpsertCartItemRequest]> = []
    const client = {
      getCart: async () => cartWith(existing()),
      upsertCartItem: async (sku: string, body: UpsertCartItemRequest) => {
        puts.push([sku, body])
        return cartWith(existing({ quantity: body.quantity }))
      },
    }
    const result = await addToCart(client, { skuId: 's1', amount: 1, priceSource: 'flash_sale', flashSaleItemId: 'f1' })
    expect(puts).toEqual([['s1', { quantity: 3, price_source: 'flash_sale', flash_sale_item_id: 'f1', expected_version: 5 }]])
    expect(result.quantity).toBe(3)
  })

  it('retries once with fresh data after CART_ITEM_CHANGED', async () => {
    let reads = 0
    const bodies: UpsertCartItemRequest[] = []
    const client = {
      getCart: async () => cartWith(existing({ version: 5 + reads++ })),
      upsertCartItem: async (_sku: string, body: UpsertCartItemRequest) => {
        bodies.push(body)
        if (bodies.length === 1) throw new ApiError({ status: 409, code: 'CART_ITEM_CHANGED' })
        return cartWith(existing())
      },
    }
    await addToCart(client, { skuId: 's1', amount: 1, priceSource: 'regular' })
    expect(bodies.map((b) => b.expected_version)).toEqual([5, 6])
  })

  it('does not retry other failures', async () => {
    const client = {
      getCart: async () => cartWith(),
      upsertCartItem: async (): Promise<Cart> => {
        throw new ApiError({ status: 409, code: 'SALE_NOT_ACTIVE' })
      },
    }
    await expect(addToCart(client, { skuId: 's1', amount: 1, priceSource: 'flash_sale', flashSaleItemId: 'f1' })).rejects.toMatchObject({
      code: 'SALE_NOT_ACTIVE',
    })
  })
})

describe('describeAddResult', () => {
  it('announces a regular to flash sale switch', () => {
    const text = describeAddResult({ cart: { id: 'c', items: [] }, quantity: 2, switchedFrom: 'regular' }, 'Áo')
    expect(text).toContain('giá thường')
    expect(text).toContain('chuyển sang giá Flash Sale')
  })
  it('has no switch note otherwise', () => {
    expect(describeAddResult({ cart: { id: 'c', items: [] }, quantity: 1, switchedFrom: null }, 'Áo')).not.toContain('chuyển')
  })
})
