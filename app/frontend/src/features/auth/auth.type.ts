import type { components } from '@/contract/api-contract'

type ApiSchemas = components['schemas']

export type Credentials = ApiSchemas['LoginRequest']
export type User = ApiSchemas['User']
export type Role = User['role']
export type AuthResult = Omit<ApiSchemas['AuthResult'], 'user'> & { user: User }

export type FieldViolation = {
  field: string
  reason: string
}

export type AuthApiErrorInit = {
  status: number
  code: string
  requestID?: string
  violations?: FieldViolation[]
  retryAfterSeconds?: number
}

export type AuthState =
  | { status: 'loading' }
  | { status: 'anonymous' }
  | { status: 'authenticated'; user: User }

export type AuthContextValue = {
  state: AuthState
  login: (credentials: Credentials) => Promise<User>
  register: (credentials: Credentials) => Promise<User>
  logout: () => Promise<void>
  refresh: () => Promise<void>
}
