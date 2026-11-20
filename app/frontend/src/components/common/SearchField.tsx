import type { ChangeEvent, FormEvent } from 'react'
import { Button } from './Button'

export type SearchFieldProps = {
  value: string
  placeholder: string
  onChange: (value: string) => void
  onSubmit: () => void
  label?: string
}

export function SearchField({ value, placeholder, onChange, onSubmit, label = 'Tìm kiếm' }: SearchFieldProps) {
  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    onSubmit()
  }

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    onChange(event.currentTarget.value)
  }

  return (
    <form role="search" onSubmit={handleSubmit} className="flex w-full items-center gap-2">
      <label htmlFor="search-field" className="sr-only">{label}</label>
      <input
        id="search-field"
        type="search"
        value={value}
        placeholder={placeholder}
        onChange={handleChange}
        className="min-h-10 min-w-0 flex-1 rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] bg-white px-3 text-sm text-[var(--ss-color-text)] placeholder:text-[var(--ss-color-muted)]"
      />
      <Button type="submit" variant="primary">Tìm</Button>
    </form>
  )
}
