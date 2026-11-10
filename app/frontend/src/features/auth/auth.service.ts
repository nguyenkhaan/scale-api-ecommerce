import axios, { type AxiosResponse } from 'axios'

import { privateClient, publicClient } from '@/api/client'

import type { AuthApiErrorInit, AuthResult, Credentials, FieldViolation, User } from './auth.type'

const csrfCookieName = 'salescale_csrf'
const csrfHeaderName = 'X-CSRF-Token'

type ProblemResponse = {
  code?: unknown
  request_id?: unknown
  violations?: unknown
}

export class AuthApiError extends Error {
  readonly status: number
  readonly code: string
  readonly requestID?: string
  readonly violations: FieldViolation[]
  readonly retryAfterSeconds?: number

  constructor(init: AuthApiErrorInit) {
    super(`api error ${init.code} (${init.status})`)
    this.name = 'AuthApiError'
    this.status = init.status
    this.code = init.code
    this.requestID = init.requestID
    this.violations = init.violations ?? []
    this.retryAfterSeconds = init.retryAfterSeconds
  }
}

export function isAuthApiError(error: unknown): error is AuthApiError {
  return error instanceof AuthApiError
}

export async function register(credentials: Credentials): Promise<AuthResult> {
  return request(publicClient.post<AuthResult>('/auth/register', credentials))
}

export async function login(credentials: Credentials): Promise<AuthResult> {
  return request(publicClient.post<AuthResult>('/auth/login', credentials))
}

export async function getCurrentUser(): Promise<User | null> {
  try {
    return await request(privateClient.get<User>('/auth/me'))
  } catch (error) {
    if (isAuthApiError(error) && error.status === 401) return null
    throw error
  }
}

export async function logout(): Promise<void> {
  await request(
    privateClient.post('/auth/logout', undefined, {
      headers: { [csrfHeaderName]: readDocumentCookie(csrfCookieName) ?? '' },
    }),
  )
}

function request<T>(operation: Promise<AxiosResponse<T>>): Promise<T> {
  return operation.then(
    (response) => response.data,
    (error: unknown) => Promise.reject(toAuthApiError(error)),
  )
}

function toAuthApiError(error: unknown): AuthApiError {
  if (isAuthApiError(error)) return error
  if (!axios.isAxiosError<ProblemResponse>(error) || error.response === undefined) {
    return new AuthApiError({ status: 0, code: 'NETWORK_ERROR' })
  }

  const { data, headers, status } = error.response
  return new AuthApiError({
    status,
    code: typeof data?.code === 'string' ? data.code : 'UNKNOWN_ERROR',
    requestID: typeof data?.request_id === 'string' ? data.request_id : headerValue(headers['x-request-id']),
    violations: parseViolations(data?.violations),
    retryAfterSeconds: parseRetryAfter(headers['retry-after']),
  })
}

function readDocumentCookie(name: string): string | null {
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

function parseViolations(value: unknown): FieldViolation[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((violation): FieldViolation[] => {
    if (!isRecord(violation) || typeof violation.field !== 'string' || typeof violation.reason !== 'string') return []
    return [{ field: violation.field, reason: violation.reason }]
  })
}

function parseRetryAfter(value: unknown): number | undefined {
  const raw = headerValue(value)
  return raw !== undefined && /^\d+$/.test(raw) ? Number(raw) : undefined
}

function headerValue(value: unknown): string | undefined {
  if (typeof value === 'string') return value
  if (Array.isArray(value) && typeof value[0] === 'string') return value[0]
  return undefined
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
