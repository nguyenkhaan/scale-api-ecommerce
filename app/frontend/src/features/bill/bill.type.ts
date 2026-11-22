export type OrderStatus = 'pending_payment' | 'confirmed' | 'payment_failed' | 'cancelled' | 'expired'

export type OrderPaymentStatus = 'pending' | 'unknown' | 'succeeded' | 'failed'

export type OrderRecipient = {
  name: string
  phone: string
  address: string
}

export type OrderItem = {
  id: string
  sku_id: string
  sku_code: string
  product_name: string
  sku_attributes?: Record<string, unknown>
  quantity: number
  unit_price: number
  subtotal: number
  price_source: 'regular' | 'flash_sale'
}

/** Order snapshot. `payable` and `status` are decided by the server; the client only displays them. */
export type Order = {
  id: string
  order_number: string
  status: OrderStatus
  currency: 'VND'
  total_amount: number
  recipient: OrderRecipient
  items: OrderItem[]
  created_at: string
  expires_at: string
  payable: boolean
  payment_status: OrderPaymentStatus | null
  reconciliation_required: boolean
}

export type OrderPage = {
  items: Order[]
  next_cursor: string | null
}

export type OrderListParams = {
  cursor?: string
  limit?: number
  status?: OrderStatus
}
