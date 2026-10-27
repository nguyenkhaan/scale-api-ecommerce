import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import type { FormErrors } from '../auth/auth-form-errors'
import { LoginFormView } from './LoginPage'
import { RegisterFormView } from './RegisterPage'

const noop = () => undefined

function base(errors: FormErrors = { fields: {} }, overrides: { submitting?: boolean; showPassword?: boolean } = {}) {
  return {
    email: '',
    password: '',
    showPassword: overrides.showPassword ?? false,
    submitting: overrides.submitting ?? false,
    errors,
    search: '',
    onEmailChange: noop,
    onPasswordChange: noop,
    onToggleShowPassword: noop,
    onSubmit: noop,
  }
}

const login = (props: ReturnType<typeof base>) =>
  renderToStaticMarkup(
    <MemoryRouter>
      <LoginFormView {...props} />
    </MemoryRouter>,
  )
const register = (props: ReturnType<typeof base>) =>
  renderToStaticMarkup(
    <MemoryRouter>
      <RegisterFormView {...props} />
    </MemoryRouter>,
  )

describe('LoginFormView', () => {
  it('renders labeled inputs, toggle button and link to register', () => {
    const markup = login(base())
    expect(markup).toContain('for="email"')
    expect(markup).toContain('for="password"')
    expect(markup).toContain('type="password"')
    expect(markup).toContain('<button type="button"')
    expect(markup).toContain('Hiện mật khẩu')
    expect(markup).toContain('href="/register"')
    expect(markup).not.toContain('role="alert"')
  })

  it('reveals the password and keeps the toggle pressed', () => {
    const markup = login(base({ fields: {} }, { showPassword: true }))
    expect(markup).toContain('type="text"')
    expect(markup).toContain('aria-pressed="true"')
    expect(markup).toContain('Ẩn mật khẩu')
  })

  it('links field errors with aria-invalid and aria-describedby', () => {
    const markup = login(base({ fields: { email: 'Email không hợp lệ', password: 'Vui lòng nhập mật khẩu' } }))
    expect(markup).toContain('aria-invalid="true"')
    expect(markup).toContain('aria-describedby="emailError"')
    expect(markup).toContain('id="emailError"')
    expect(markup).toContain('aria-describedby="passwordError"')
    expect(markup).toContain('Vui lòng nhập mật khẩu')
  })

  it('shows a generic credentials error in an alert', () => {
    const markup = login(base({ fields: {}, general: { kind: 'invalid-credentials' } }))
    expect(markup).toContain('role="alert"')
    expect(markup).toContain('Email hoặc mật khẩu không đúng')
  })

  it('shows the wait time for RATE_LIMITED', () => {
    const markup = login(base({ fields: {}, general: { kind: 'rate-limited', retryAfterSeconds: 42 } }))
    expect(markup).toContain('Vui lòng thử lại sau 42 giây.')
  })

  it('shows request id on unexpected errors', () => {
    const markup = login(base({ fields: {}, general: { kind: 'unexpected', requestID: 'req-77' } }))
    expect(markup).toContain('req-77')
    expect(markup).toContain('role="alert"')
  })

  it('disables the submit button while submitting', () => {
    const markup = login(base({ fields: {} }, { submitting: true }))
    expect(markup).toContain('disabled=""')
    expect(markup).toContain('Đang đăng nhập...')
  })

  it('forwards the next param to the register link', () => {
    const markup = login({ ...base(), search: '?next=%2Faccount' })
    expect(markup).toContain('href="/register?next=%2Faccount"')
  })
})

describe('RegisterFormView', () => {
  it('describes the password rule and links to login', () => {
    const markup = register(base())
    expect(markup).toContain('autoComplete="new-password"')
    expect(markup).toContain('id="passwordHint"')
    expect(markup).toContain('aria-describedby="passwordHint"')
    expect(markup).toContain('href="/login"')
  })

  it('shows the already registered message with a login link', () => {
    const markup = register(base({ fields: {}, general: { kind: 'email-taken' } }))
    expect(markup).toContain('role="alert"')
    expect(markup).toContain('Email này đã được đăng ký')
    expect(markup).toContain('href="/login"')
  })
})
