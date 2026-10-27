import { useState, type FormEvent } from 'react'
import type { Credentials, User } from '../lib/auth-api'
import { hasFieldErrors, toFormErrors, validateCredentials, type AuthMode, type FormErrors } from './auth-form-errors'

type Options = {
  mode: AuthMode
  submit: (credentials: Credentials) => Promise<User>
  onSuccess: () => void
}

export function useAuthForm({ mode, submit, onSuccess }: Options) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [errors, setErrors] = useState<FormErrors>({ fields: {} })

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (submitting) return
    const trimmedEmail = email.trim()
    const fields = validateCredentials(mode, trimmedEmail, password)
    if (hasFieldErrors(fields)) {
      setErrors({ fields })
      return
    }
    setErrors({ fields: {} })
    setSubmitting(true)
    try {
      await submit({ email: trimmedEmail, password })
      onSuccess()
    } catch (error) {
      setErrors(toFormErrors(mode, error))
      setSubmitting(false)
    }
  }

  return {
    mode,
    email,
    password,
    showPassword,
    submitting,
    errors,
    onEmailChange: setEmail,
    onPasswordChange: setPassword,
    onToggleShowPassword: () => setShowPassword((current) => !current),
    onSubmit,
  }
}
