import type { components } from '@/contract/api-contract'

type ApiSchemas = components['schemas']

export type PriceSource = ApiSchemas['CartItem']['price_source']
export type CartItem = ApiSchemas['CartItem']
export type Cart = ApiSchemas['Cart']
export type UpsertCartItemRequest = ApiSchemas['UpsertCartItemRequest']
export type SelectedCartItem = ApiSchemas['SelectedCartItem']
export type CheckoutSelection = ApiSchemas['CheckoutSelection']
export type CheckoutLine = ApiSchemas['CheckoutLine']
export type CheckoutPreview = ApiSchemas['CheckoutPreview']
export type FieldViolation = ApiSchemas['FieldViolation']
export type LineError = ApiSchemas['LineProblem']

export type CartApiErrorInit = {
  status: number
  code: string
  requestID?: string
  violations?: FieldViolation[]
  lineErrors?: LineError[]
  retryAfterSeconds?: number
}
