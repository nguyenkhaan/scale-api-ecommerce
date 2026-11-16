export const csrfCookieName = 'salescale_csrf'

export function readDocumentCookie(name: string): string | null {
  if (typeof document === 'undefined') return null

  for (const cookie of document.cookie.split(';')) {
    const separator = cookie.indexOf('=')
    if (separator <= 0 || cookie.slice(0, separator).trim() !== name) continue
    try {
      return decodeURIComponent(cookie.slice(separator + 1).trim())
    } catch {
      return null
    }
  }
  return null
}
