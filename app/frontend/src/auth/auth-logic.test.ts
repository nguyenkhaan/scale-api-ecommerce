import { describe, expect, it } from 'bun:test'
import { ApiError } from '../lib/auth-api'
import { toFormErrors, validateCredentials } from './auth-form-errors'
import { loginRedirectPath, safeNextPath } from './redirect-paths'

describe('safeNextPath', () => {
  it('accepts same-origin relative paths', () => {
    expect(safeNextPath('/account?x=1')).toBe('/account?x=1')
  })

  it('rejects absolute, protocol-relative and missing values', () => {
    expect(safeNextPath(null)).toBe('/')
    expect(safeNextPath('https://evil.test')).toBe('/')
    expect(safeNextPath('//evil.test')).toBe('/')
    expect(safeNextPath('/\\evil.test')).toBe('/')
    expect(safeNextPath('account')).toBe('/')
  })

  it('rejects control characters that the URL parser would strip', () => {
    expect(safeNextPath('/\t/evil.test')).toBe('/')
    expect(safeNextPath('/\n/evil.test')).toBe('/')
    expect(safeNextPath('/\u007f/evil.test')).toBe('/')
  })
})

describe('loginRedirectPath', () => {
  it('encodes the current path and search', () => {
    expect(loginRedirectPath('/admin', '?a=1&b=2')).toBe('/login?next=%2Fadmin%3Fa%3D1%26b%3D2')
  })
})

describe('validateCredentials', () => {
  it('requires email and password', () => {
    expect(validateCredentials('login', '', '')).toEqual({
      email: 'Vui lòng nhập email',
      password: 'Vui lòng nhập mật khẩu',
    })
  })

  it('rejects malformed and overlong emails', () => {
    expect(validateCredentials('login', 'nope', 'x').email).toBe('Email không hợp lệ')
    expect(validateCredentials('login', `${'a'.repeat(250)}@b.co`, 'x').email).toBe('Email không hợp lệ')
  })

  it('does not enforce length on login passwords', () => {
    expect(validateCredentials('login', 'a@b.co', 'x')).toEqual({})
  })

  it('counts the minimum in characters and the maximum in UTF-8 bytes, like the server', () => {
    expect(validateCredentials('register', 'a@b.co', 'short').password).toBeDefined()
    expect(validateCredentials('register', 'a@b.co', 'a'.repeat(72))).toEqual({})
    expect(validateCredentials('register', 'a@b.co', 'a'.repeat(73)).password).toBeDefined()
    // 3 characters but 9 bytes: still too short for the server
    expect(validateCredentials('register', 'a@b.co', 'ấấấ').password).toBeDefined()
    // 8 characters, 24 bytes: accepted
    expect(validateCredentials('register', 'a@b.co', 'ấ'.repeat(8))).toEqual({})
    // 25 x 3-byte chars = 75 bytes but only 25 characters
    expect(validateCredentials('register', 'a@b.co', 'ấ'.repeat(25)).password).toBeDefined()
  })
})

describe('toFormErrors', () => {
  it('maps violations to field errors', () => {
    const error = new ApiError({
      status: 422,
      code: 'VALIDATION_ERROR',
      violations: [{ field: 'email', reason: 'x' }, { field: 'password', reason: 'y' }],
    })
    const result = toFormErrors('register', error)
    expect(result.fields.email).toBeDefined()
    expect(result.fields.password).toBeDefined()
    expect(result.general).toBeUndefined()
  })

  it('maps codes to general errors', () => {
    expect(toFormErrors('login', new ApiError({ status: 401, code: 'UNAUTHENTICATED' })).general).toEqual({
      kind: 'invalid-credentials',
    })
    expect(
      toFormErrors('login', new ApiError({ status: 429, code: 'RATE_LIMITED', retryAfterSeconds: 9 })).general,
    ).toEqual({ kind: 'rate-limited', retryAfterSeconds: 9 })
    expect(toFormErrors('register', new ApiError({ status: 409, code: 'EMAIL_ALREADY_REGISTERED' })).general).toEqual({
      kind: 'email-taken',
    })
    expect(
      toFormErrors('login', new ApiError({ status: 500, code: 'INTERNAL_ERROR', requestID: 'r' })).general,
    ).toEqual({ kind: 'unexpected', requestID: 'r' })
  })

  it('treats unknown errors as unexpected', () => {
    expect(toFormErrors('login', new Error('x')).general).toEqual({ kind: 'unexpected' })
  })
})
