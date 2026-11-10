import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { useAuthForm } from '../auth-form.hook'
import { safeNextPath, useAuth } from '../auth.hook'
import { AuthFormView, type AuthFormViewProps } from '../components/AuthFormView'

type LoginFormViewProps = Omit<AuthFormViewProps, 'mode' | 'footer' | 'loginHref'> & { search: string }

export function LoginFormView({ search, ...props }: LoginFormViewProps) {
  return (
    <AuthFormView
      {...props}
      mode="login"
      loginHref={`/login${search}`}
      footer={
        <>
          Chưa có tài khoản?{' '}
          <Link to={`/register${search}`} className="font-semibold underline underline-offset-4 outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500">
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
