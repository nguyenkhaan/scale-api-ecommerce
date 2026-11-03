import { errorFromResponse, invalidResponseError, networkError } from './api-error'
import { csrfCookieName, readDocumentCookie } from './auth-api'

export type FlashSaleItem = {
  id: string
  sku_id: string
  sale_price: number
  quota_total: number
  purchase_limit: number
}

export type FlashSaleCampaign = {
  id: string
  name: string
  status: 'draft' | 'scheduled' | 'active' | 'stopped' | 'ended'
  starts_at: string
  ends_at: string
  items: FlashSaleItem[]
}

export type FlashSalePage = {
  items: FlashSaleCampaign[]
  next_cursor: string | null
}

export type CreateFlashSaleItem = {
  sku_id: string
  sale_price: number
  quota_total: number
  purchase_limit: number
}

export type CreateFlashSaleRequest = {
  name: string
  starts_at: string
  ends_at: string
  items: CreateFlashSaleItem[]
}

export type UpdateFlashSaleRequest = {
  name?: string
  starts_at?: string
  ends_at?: string
  items?: CreateFlashSaleItem[]
}

export type ListParams = {
  status?: string
  cursor?: string
  limit?: number
}

type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

export type FlashSaleClientOptions = {
  baseURL: string
  fetcher?: Fetcher
  readCookie?: (name: string) => string | null
}

export function createFlashSaleClient(options: FlashSaleClientOptions) {
  const baseURL = options.baseURL.replace(/\/$/, '')
  const fetcher: Fetcher = options.fetcher ?? ((input, init) => globalThis.fetch(input, init))
  const readCookie = options.readCookie ?? readDocumentCookie

  async function send(path: string, init: { method: string; body?: unknown; csrf?: string }) {
    const headers: Record<string, string> = { Accept: 'application/json' }
    if (init.body !== undefined) headers['Content-Type'] = 'application/json'
    if (init.csrf !== undefined) headers['X-CSRF-Token'] = init.csrf
    try {
      return await fetcher(`${baseURL}${path}`, {
        method: init.method,
        credentials: 'include',
        headers,
        body: init.body === undefined ? undefined : JSON.stringify(init.body),
      })
    } catch {
      throw networkError()
    }
  }

  async function readJSON<T>(response: Response): Promise<T> {
    try {
      return (await response.json()) as T
    } catch {
      throw invalidResponseError(response)
    }
  }

  function getCsrf() {
    return readCookie(csrfCookieName) ?? ''
  }

  function buildQuery(params?: ListParams): string {
    if (!params) return ''
    const q = new URLSearchParams()
    if (params.status) q.set('status', params.status)
    if (params.cursor) q.set('cursor', params.cursor)
    if (params.limit) q.set('limit', String(params.limit))
    const str = q.toString()
    return str ? `?${str}` : ''
  }

  return {
    async listPublicFlashSales(params?: ListParams): Promise<FlashSalePage> {
      const res = await send(`/api/v1/flash-sales${buildQuery(params)}`, { method: 'GET' })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSalePage>(res)
    },

    async getPublicFlashSale(id: string): Promise<FlashSaleCampaign> {
      const res = await send(`/api/v1/flash-sales/${id}`, { method: 'GET' })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSaleCampaign>(res)
    },

    async listAdminFlashSales(params?: ListParams): Promise<FlashSalePage> {
      const res = await send(`/api/v1/admin/flash-sales${buildQuery(params)}`, { method: 'GET' })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSalePage>(res)
    },

    async getAdminFlashSale(id: string): Promise<FlashSaleCampaign> {
      const res = await send(`/api/v1/admin/flash-sales/${id}`, { method: 'GET' })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSaleCampaign>(res)
    },

    async createFlashSale(data: CreateFlashSaleRequest): Promise<FlashSaleCampaign> {
      const res = await send('/api/v1/admin/flash-sales', {
        method: 'POST',
        body: data,
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSaleCampaign>(res)
    },

    async updateFlashSale(id: string, data: UpdateFlashSaleRequest): Promise<FlashSaleCampaign> {
      const res = await send(`/api/v1/admin/flash-sales/${id}`, {
        method: 'PATCH',
        body: data,
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSaleCampaign>(res)
    },

    async scheduleFlashSale(id: string): Promise<FlashSaleCampaign> {
      const res = await send(`/api/v1/admin/flash-sales/${id}/schedule`, {
        method: 'POST',
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSaleCampaign>(res)
    },

    async cancelScheduleFlashSale(id: string): Promise<FlashSaleCampaign> {
      const res = await send(`/api/v1/admin/flash-sales/${id}/cancel-schedule`, {
        method: 'POST',
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSaleCampaign>(res)
    },

    async stopFlashSale(id: string): Promise<FlashSaleCampaign> {
      const res = await send(`/api/v1/admin/flash-sales/${id}/stop`, {
        method: 'POST',
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<FlashSaleCampaign>(res)
    },
  }
}

export const flashSaleClient = createFlashSaleClient({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
})
