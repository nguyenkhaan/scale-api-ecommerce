import type { FormEvent, ReactNode } from 'react'
import type { AuthMode, FormErrors } from './auth-form-errors'
import { GeneralErrorAlert } from './GeneralErrorAlert'

export type AuthFormViewProps = {
  mode: AuthMode
  email: string
  password: string
  showPassword: boolean
  submitting: boolean
  errors: FormErrors
  /** Link to the login page (kept with the `next` param), used by the "email already registered" message. */
  loginHref: string
  footer: ReactNode
  onEmailChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onToggleShowPassword: () => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

const focusRing = 'outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500'
const inputClass = `block min-h-11 w-full border bg-white px-3 py-2 text-base text-slate-950 ${focusRing}`

const copy = {
  login: { title: 'Đăng nhập', submit: 'Đăng nhập', submitting: 'Đang đăng nhập...', autocomplete: 'current-password' },
  register: { title: 'Đăng ký', submit: 'Đăng ký', submitting: 'Đang đăng ký...', autocomplete: 'new-password' },
} as const

export function AuthFormView(props: AuthFormViewProps) {
  const { mode, errors, submitting, showPassword } = props
  const text = copy[mode]
  const emailError = errors.fields.email
  const passwordError = errors.fields.password
  const hint = mode === 'register' ? 'passwordHint' : undefined
  const passwordDescribedBy = [hint, passwordError ? 'passwordError' : undefined].filter(Boolean).join(' ') || undefined

  return (
    <section aria-labelledby="auth-title" className="mx-auto w-full max-w-md">
      <h1 id="auth-title" className="text-3xl font-bold tracking-tight text-slate-950">
        {text.title}
      </h1>
      <form noValidate onSubmit={props.onSubmit} className="mt-6 border border-slate-300 bg-white p-5 sm:p-6">
        {errors.general !== undefined && <GeneralErrorAlert error={errors.general} loginHref={props.loginHref} />}

        <label htmlFor="email" className="block text-sm font-semibold text-slate-800">
          Email
        </label>
        <input
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          value={props.email}
          onChange={(event) => props.onEmailChange(event.target.value)}
          aria-invalid={emailError !== undefined}
          aria-describedby={emailError ? 'emailError' : undefined}
          className={`${inputClass} mt-1 ${emailError ? 'border-rose-600' : 'border-slate-400'}`}
        />
        {emailError && (
          <p id="emailError" className="mt-1 text-sm text-rose-700">
            {emailError}
          </p>
        )}

        <label htmlFor="password" className="mt-5 block text-sm font-semibold text-slate-800">
          Mật khẩu
        </label>
        <div className="mt-1 flex">
          <input
            id="password"
            name="password"
            type={showPassword ? 'text' : 'password'}
            autoComplete={text.autocomplete}
            value={props.password}
            onChange={(event) => props.onPasswordChange(event.target.value)}
            aria-invalid={passwordError !== undefined}
            aria-describedby={passwordDescribedBy}
            className={`${inputClass} min-w-0 flex-1 ${passwordError ? 'border-rose-600' : 'border-slate-400'}`}
          />
          <button
            type="button"
            onClick={props.onToggleShowPassword}
            aria-pressed={showPassword}
            className={`min-h-11 border border-l-0 border-slate-400 bg-white px-3 text-sm font-semibold text-slate-700 hover:bg-stone-100 ${focusRing}`}
          >
            {showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
          </button>
        </div>
        {mode === 'register' && (
          <p id={hint} className="mt-1 text-sm text-slate-600">
            Ít nhất 8 ký tự, tối đa 72 byte (chữ có dấu chiếm 2–3 byte).
          </p>
        )}
        {passwordError && (
          <p id="passwordError" className="mt-1 text-sm text-rose-700">
            {passwordError}
          </p>
        )}

        <button
          type="submit"
          disabled={submitting}
          aria-busy={submitting}
          className={`mt-6 min-h-11 w-full border border-slate-900 bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60 ${focusRing}`}
        >
          {submitting ? text.submitting : text.submit}
        </button>
      </form>
      <p className="mt-5 text-sm text-slate-700">{props.footer}</p>
    </section>
  )
}
