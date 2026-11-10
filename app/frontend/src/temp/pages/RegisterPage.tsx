import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { AuthFormView, type AuthFormViewProps } from '../auth/AuthFormView'
import { safeNextPath } from '../auth/redirect-paths'
import { useAuth } from '../auth/use-auth'
import { useAuthForm } from '../auth/use-auth-form'

type RegisterFormViewProps = Omit<AuthFormViewProps, 'mode' | 'footer' | 'loginHref'> & {
  /** Query string (with leading "?" or empty) forwarded to the sibling page to keep `next`. */
  search: string
}

export function RegisterFormView({ search, ...props }: RegisterFormViewProps) {
  return (
    <AuthFormView
      {...props}
      mode="register"
      loginHref={`/login${search}`}
      footer={
        <>
          Đã có tài khoản?{' '}
          <Link
            to={`/login${search}`}
            className="font-semibold underline underline-offset-4 outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500"
          >
            Đăng nhập
          </Link>
        </>
      }
    />
  )
}

export function RegisterPage() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const next = params.get('next')
  const form = useAuthForm({
    mode: 'register',
    submit: register,
    onSuccess: () => navigate(safeNextPath(next), { replace: true }),
  })
  const search = next === null ? '' : `?next=${encodeURIComponent(next)}`

  return <RegisterFormView {...form} search={search} />
}
