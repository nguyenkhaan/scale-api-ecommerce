export function FieldErrors({ id, messages }: { id: string; messages: string[] }) {
  if (messages.length === 0) return null
  return (
    <p id={id} className="mt-1 text-xs font-semibold text-rose-700">
      {messages.join(' ')}
    </p>
  )
}

export const inputClass =
  'mt-1 min-h-11 w-full rounded border border-slate-300 bg-white px-3 py-2 text-sm focus:border-amber-500 focus:outline-none aria-[invalid=true]:border-rose-500'
