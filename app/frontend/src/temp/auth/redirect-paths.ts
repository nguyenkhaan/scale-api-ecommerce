/**
 * Returns `next` only when it is a same-origin relative path; otherwise `/`. Control characters are rejected
 * because the URL parser strips them, which would turn `/\t/evil.test` into the protocol-relative `//evil.test`.
 */
export function safeNextPath(next: string | null): string {
  if (
    next === null ||
    !next.startsWith('/') ||
    next.startsWith('//') ||
    next.includes('\\') ||
    hasControlCharacter(next)
  ) {
    return '/'
  }
  return next
}

export function loginRedirectPath(pathname: string, search: string): string {
  return `/login?next=${encodeURIComponent(`${pathname}${search}`)}`
}

function hasControlCharacter(value: string): boolean {
  for (let index = 0; index < value.length; index++) {
    const code = value.charCodeAt(index)
    if (code < 0x20 || code === 0x7f) return true
  }
  return false
}
