import { describe, expect, it } from 'bun:test'
import { errorFromResponse } from './api-error'

describe('errorFromResponse', () => {
  it('parses line_errors and drops malformed entries', async () => {
    const response = Response.json(
      {
        code: 'PURCHASE_LIMIT_EXCEEDED',
        line_errors: [
          { sku_id: 'a', requested_quantity: 5, code: 'PURCHASE_LIMIT_EXCEEDED', available_quantity: 2 },
          { sku_id: 'b', requested_quantity: 1, code: 'ITEM_UNAVAILABLE', current_unit_price: 900 },
          { code: 'no-sku' },
          'junk',
        ],
      },
      { status: 409 },
    )
    const error = await errorFromResponse(response)
    expect(error.code).toBe('PURCHASE_LIMIT_EXCEEDED')
    expect(error.lineErrors).toEqual([
      { sku_id: 'a', requested_quantity: 5, code: 'PURCHASE_LIMIT_EXCEEDED', available_quantity: 2 },
      { sku_id: 'b', requested_quantity: 1, code: 'ITEM_UNAVAILABLE', current_unit_price: 900 },
    ])
  })

  it('defaults to no line errors', async () => {
    const error = await errorFromResponse(Response.json({ code: 'VALIDATION_ERROR' }, { status: 422 }))
    expect(error.lineErrors).toEqual([])
  })
})
