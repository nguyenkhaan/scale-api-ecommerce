import type { components } from '@/contract/api-contract'
import type { FieldViolation, LineError } from '@/lib/api-error'

type ApiSchemas = components['schemas']

export type PlaceOrderRequest = ApiSchemas['PlaceOrderRequest']
export type IdempotencyResult = ApiSchemas['IdempotencyResult']
export type Recipient = ApiSchemas['Recipient']
export type Order = ApiSchemas['Order']

export type RecipientInput = {
  name: string
  phone: string
  address: string
}

export type RecipientErrors = Partial<Record<keyof RecipientInput, string>>

export type CheckoutIntent = {
  key: string
  fingerprint: string
}

export type CheckoutOutcome =
  | { kind: 'created'; order: Order }
  | { kind: 'unknown'; requestID?: string }
  | { kind: 'processing'; retryAfterSeconds: number }
  | { kind: 'overload'; retryAfterSeconds: number }
  | { kind: 'rejected'; code: string; lineErrors: LineError[] }
  | { kind: 'price_changed'; lineErrors: LineError[] }
  | { kind: 'cart_changed' }
  | { kind: 'intent_conflict' }
  | { kind: 'invalid'; violations: FieldViolation[] }
  | { kind: 'unauthenticated' }
