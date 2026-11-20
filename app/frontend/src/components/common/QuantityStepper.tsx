export type QuantityStepperProps = {
  value: number
  min: number
  max: number
  busy: boolean
  label: string
  onChange: (value: number) => void
}

export function QuantityStepper({ value, min, max, busy, label, onChange }: QuantityStepperProps) {
  const decrementDisabled = busy || value <= min
  const incrementDisabled = busy || value >= max
  return (
    <div role="group" aria-label={label} className="inline-flex items-center rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)]">
      <button
        type="button"
        aria-label={`Giảm ${label}`}
        disabled={decrementDisabled}
        onClick={() => onChange(value - 1)}
        className="min-h-11 min-w-11 px-2 text-base font-semibold text-[var(--ss-color-text)] hover:bg-[var(--ss-color-app-bg)] disabled:cursor-not-allowed disabled:opacity-40"
      >
        −
      </button>
      <output aria-live="polite" className="min-w-10 px-2 text-center text-sm font-semibold text-[var(--ss-color-text)]">
        {value}
      </output>
      <button
        type="button"
        aria-label={`Tăng ${label}`}
        disabled={incrementDisabled}
        onClick={() => onChange(value + 1)}
        className="min-h-11 min-w-11 px-2 text-base font-semibold text-[var(--ss-color-text)] hover:bg-[var(--ss-color-app-bg)] disabled:cursor-not-allowed disabled:opacity-40"
      >
        +
      </button>
    </div>
  )
}
