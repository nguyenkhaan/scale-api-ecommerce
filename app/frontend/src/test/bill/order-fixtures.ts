import type { Order } from '@/features/bill/bill.type'

export function makeOrder(overrides: Partial<Order> = {}): Order {
  return {
    id: 'order-1',
    order_number: 'SS-1010-0042',
    status: 'pending_payment',
    currency: 'VND',
    total_amount: 848000,
    recipient: { name: 'Nguyễn Minh Anh', phone: '0901234567', address: '12 Nguyễn Trãi, Quận 1' },
    items: [
      {
        id: 'item-1',
        sku_id: 'sku-1',
        sku_code: 'AT-DEN-M',
        product_name: 'Áo thun basic',
        sku_attributes: { Màu: 'Đen', 'Kích cỡ': 'M' },
        quantity: 2,
        unit_price: 129000,
        subtotal: 258000,
        price_source: 'flash_sale',
      },
    ],
    created_at: '2026-10-10T13:41:00Z',
    expires_at: '2026-10-10T13:46:00Z',
    payable: true,
    payment_status: 'pending',
    reconciliation_required: false,
    ...overrides,
  }
}

export type StubCall = { url: string; method: string; headers: Record<string, string>; body?: unknown }

/** Fetch stub returning queued responses in order; records every call. */
export function stubFetcher(responses: Array<{ status: number; body?: unknown }>) {
  const calls: StubCall[] = []
  const fetcher = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    calls.push({
      url: String(input),
      method: init?.method ?? 'GET',
      headers: (init?.headers ?? {}) as Record<string, string>,
      body: typeof init?.body === 'string' ? JSON.parse(init.body) : undefined,
    })
    const next = responses.shift() ?? { status: 500 }
    return new Response(next.body === undefined ? null : JSON.stringify(next.body), {
      status: next.status,
      headers: { 'Content-Type': 'application/json' },
    })
  }
  return { fetcher, calls }
}
