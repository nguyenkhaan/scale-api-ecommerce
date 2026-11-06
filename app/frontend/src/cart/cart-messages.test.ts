import { describe, expect, it } from 'bun:test'
import { ApiError } from '../lib/api-error'
import { canSwitchToRegular, describeCartFailure, lineErrorMessage } from './cart-messages'

describe('lineErrorMessage', () => {
  it('maps each code to Vietnamese copy with the remaining quantity', () => {
    expect(lineErrorMessage({ code: 'PURCHASE_LIMIT_EXCEEDED', available_quantity: 2 })).toBe('Vượt giới hạn mua — còn được mua 2')
    expect(lineErrorMessage({ code: 'SALE_QUOTA_EXCEEDED', available_quantity: 0 })).toBe('Đã hết suất giá flash sale — còn 0')
    expect(lineErrorMessage({ code: 'SALE_NOT_ACTIVE' })).toBe('Chương trình chưa mở hoặc đã kết thúc')
    expect(lineErrorMessage({ code: 'ITEM_UNAVAILABLE' })).toBe('Sản phẩm đã ngừng bán')
    expect(lineErrorMessage({ code: 'INSUFFICIENT_STOCK', available_quantity: 4 })).toBe('Chỉ còn 4 sản phẩm')
  })

  it('never leaks an unknown raw code', () => {
    expect(lineErrorMessage({ code: 'SOMETHING_NEW' })).not.toContain('SOMETHING_NEW')
    expect(lineErrorMessage({ code: 'INSUFFICIENT_STOCK' })).toBe('Không đủ tồn kho')
  })
})

describe('canSwitchToRegular', () => {
  it('offers the explicit fallback only for flash lines with sale-related rejections', () => {
    expect(canSwitchToRegular('flash_sale', 'SALE_NOT_ACTIVE')).toBe(true)
    expect(canSwitchToRegular('flash_sale', 'SALE_QUOTA_EXCEEDED')).toBe(true)
    expect(canSwitchToRegular('flash_sale', 'PURCHASE_LIMIT_EXCEEDED')).toBe(true)
    expect(canSwitchToRegular('flash_sale', 'INSUFFICIENT_STOCK')).toBe(false)
    expect(canSwitchToRegular('regular', 'SALE_NOT_ACTIVE')).toBe(false)
  })
})

describe('describeCartFailure', () => {
  it('flags CART_ITEM_CHANGED so the caller reloads', () => {
    const f = describeCartFailure(new ApiError({ status: 409, code: 'CART_ITEM_CHANGED' }))
    expect(f.cartChanged).toBe(true)
    expect(f.text).not.toContain('CART_ITEM_CHANGED')
  })

  it('keeps line errors and the no-hold summary for a rejected preview', () => {
    const lineErrors = [{ sku_id: 's', requested_quantity: 2, code: 'INSUFFICIENT_STOCK', available_quantity: 1 }]
    const f = describeCartFailure(new ApiError({ status: 409, code: 'INSUFFICIENT_STOCK', lineErrors }))
    expect(f.lineErrors).toEqual(lineErrors)
    expect(f.text).toContain('Chưa giữ hàng nào; giỏ hàng vẫn giữ nguyên')
  })

  it('maps 422 violations to friendly text instead of VALIDATION_ERROR', () => {
    const cartLimit = describeCartFailure(
      new ApiError({ status: 422, code: 'VALIDATION_ERROR', violations: [{ field: 'cart', reason: 'cannot exceed 20 distinct items' }] }),
    )
    expect(cartLimit.text).toBe('Giỏ hàng tối đa 20 SKU khác nhau.')
    const quantity = describeCartFailure(
      new ApiError({ status: 422, code: 'VALIDATION_ERROR', violations: [{ field: 'quantity', reason: 'exceeds maximum allowed quantity' }] }),
    )
    expect(quantity.text).toContain('Số lượng vượt mức tối đa')
  })

  it('treats 401 as unauthenticated and network errors as connectivity problems', () => {
    expect(describeCartFailure(new ApiError({ status: 401, code: 'UNAUTHENTICATED' })).unauthenticated).toBe(true)
    expect(describeCartFailure(new ApiError({ status: 0, code: 'NETWORK_ERROR' })).text).toContain('kết nối')
  })

  it('exposes a single-line rejection code for the line being edited', () => {
    const f = describeCartFailure(new ApiError({ status: 409, code: 'SALE_NOT_ACTIVE' }))
    expect(f.lineCode).toBe('SALE_NOT_ACTIVE')
    expect(f.text).toBe('Chương trình chưa mở hoặc đã kết thúc')
  })
})
