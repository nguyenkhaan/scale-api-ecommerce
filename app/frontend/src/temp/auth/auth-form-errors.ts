import { ApiError } from '../lib/auth-api'

export type AuthMode = 'login' | 'register'

export type FieldErrors = { email?: string; password?: string }

export type GeneralError =
  | { kind: 'invalid-credentials' }
  | { kind: 'rate-limited'; retryAfterSeconds?: number }
  | { kind: 'email-taken' }
  | { kind: 'unexpected'; requestID?: string }

export type FormErrors = {
  fields: FieldErrors
  general?: GeneralError
}

const maxEmailLength = 254
// Mirrors the server: at least 8 characters, at most 72 UTF-8 bytes (the bcrypt limit).
const minPasswordCharacters = 8
const maxPasswordBytes = 72

const messages = {
  emailRequired: 'Vui lòng nhập email',
  emailInvalid: 'Email không hợp lệ',
  passwordRequired: 'Vui lòng nhập mật khẩu',
  passwordLength: 'Mật khẩu cần ít nhất 8 ký tự và không quá 72 byte',
}

/** Client-side check for UX only; the server stays authoritative. */
export function validateCredentials(mode: AuthMode, email: string, password: string): FieldErrors {
  const errors: FieldErrors = {}
  const trimmed = email.trim()
  if (trimmed === '') errors.email = messages.emailRequired
  else if (trimmed.length > maxEmailLength || !/^[^\s@]+@[^\s@]+$/.test(trimmed)) {
    errors.email = messages.emailInvalid
  }

  if (password === '') {
    errors.password = messages.passwordRequired
  } else if (mode === 'register') {
    const characters = [...password].length
    const bytes = new TextEncoder().encode(password).length
    if (characters < minPasswordCharacters || bytes > maxPasswordBytes) errors.password = messages.passwordLength
  }
  return errors
}

export function hasFieldErrors(errors: FieldErrors): boolean {
  return errors.email !== undefined || errors.password !== undefined
}

/** Maps a failed request to renderable errors by stable `code` and `violations`. */
export function toFormErrors(mode: AuthMode, error: unknown): FormErrors {
  if (!(error instanceof ApiError)) return { fields: {}, general: { kind: 'unexpected' } }

  switch (error.code) {
    case 'VALIDATION_ERROR': {
      const fields: FieldErrors = {}
      for (const violation of error.violations) {
        if (violation.field === 'email') fields.email = messages.emailInvalid
        if (violation.field === 'password') {
          fields.password = mode === 'register' ? messages.passwordLength : messages.passwordRequired
        }
      }
      if (hasFieldErrors(fields)) return { fields }
      return { fields: {}, general: { kind: 'unexpected', requestID: error.requestID } }
    }
    case 'UNAUTHENTICATED':
      return { fields: {}, general: { kind: 'invalid-credentials' } }
    case 'RATE_LIMITED':
      return { fields: {}, general: { kind: 'rate-limited', retryAfterSeconds: error.retryAfterSeconds } }
    case 'EMAIL_ALREADY_REGISTERED':
      return { fields: {}, general: { kind: 'email-taken' } }
    default:
      return { fields: {}, general: { kind: 'unexpected', requestID: error.requestID } }
  }
}
