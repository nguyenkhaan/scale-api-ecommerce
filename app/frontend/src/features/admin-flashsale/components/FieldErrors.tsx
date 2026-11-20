export function FieldErrors({ id, messages }: { id: string; messages: string[] }) {
  if (messages.length === 0) return null
  return (
    <p id={id} role="alert" className="mt-1 text-xs font-semibold text-[var(--ss-color-danger)]">
      {messages.join(' ')}
    </p>
  )
}

export const inputClass =
  'mt-1 min-h-10 w-full rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] bg-white px-3 py-2 text-sm text-[var(--ss-color-text)] focus:border-[var(--ss-color-primary)] focus:outline-none aria-[invalid=true]:border-[var(--ss-color-danger)]'
