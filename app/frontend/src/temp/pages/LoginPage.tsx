import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { AuthFormView, type AuthFormViewProps } from '../auth/AuthFormView'
import { safeNextPath } from '../auth/redirect-paths'
import { useAuth } from '../auth/use-auth'
import { useAuthForm } from '../auth/use-auth-form'

type LoginFormViewProps = Omit<AuthFormViewProps, 'mode' | 'footer' | 'loginHref'> & {
  /** Query string (with leading "?" or empty) forwarded to the sibling page to keep `next`. */
  search: string
}

export function LoginFormView({ search, ...props }: LoginFormViewProps) {
  return (
    <AuthFormView
      {...props}
      mode="login"
      loginHref={`/login${search}`}
      footer={
        <>
          Chưa có tài khoản?{' '}
          <Link
            to={`/register${search}`}
            className="font-semibold underline underline-offset-4 outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500"
          >
            Đăng ký
          </Link>
        </>
      }
    />
  )
}

export function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const next = params.get('next')
  const form = useAuthForm({
    mode: 'login',
    submit: login,
    onSuccess: () => navigate(safeNextPath(next), { replace: true }),
  })
  const search = next === null ? '' : `?next=${encodeURIComponent(next)}`

  return <LoginFormView {...form} search={search} />
}
