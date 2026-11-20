export function normalizeApiBaseUrl(raw: string | undefined): string {
  const normalized = (raw ?? '').trim().replace(/\/+$/, '')
  if (normalized === '') return '/api'
  return normalized.endsWith('/api') ? normalized : `${normalized}/api`
}

export const API_BASE_URL = normalizeApiBaseUrl(import.meta.env.VITE_API_BASE_URL)
