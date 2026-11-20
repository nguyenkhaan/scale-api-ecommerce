import type { ReactNode } from 'react'

export type FormControlProps = {
  'aria-invalid': boolean
  'aria-describedby'?: string
}

export type FormFieldProps = {
  id: string
  label: string
  errorMessages: string[]
  helper?: string
  children: (controlProps: FormControlProps) => ReactNode
}

export function FormField({ id, label, errorMessages, helper, children }: FormFieldProps) {
  const hasErrors = errorMessages.length > 0
  const descriptionIDs = [helper ? `${id}-description` : '', hasErrors ? `${id}-errors` : ''].filter(Boolean).join(' ')
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-[var(--ss-color-text)]">
        {label}
      </label>
      {children({ 'aria-invalid': hasErrors, 'aria-describedby': descriptionIDs || undefined })}
      {helper && (
        <span id={`${id}-description`} className="text-xs text-[var(--ss-color-muted)]">
          {helper}
        </span>
      )}
      {hasErrors && (
        <ul id={`${id}-errors`} role="alert" className="flex list-none flex-col gap-1 p-0 text-xs text-[var(--ss-color-danger)]">
          {errorMessages.map((message) => <li key={message}>{message}</li>)}
        </ul>
      )}
    </div>
  )
}
