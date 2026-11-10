import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { useAuthForm } from '../auth-form.hook'
import { safeNextPath, useAuth } from '../auth.hook'
import { AuthFormView, type AuthFormViewProps } from '../components/AuthFormView'

type RegisterFormViewProps = Omit<AuthFormViewProps, 'mode' | 'footer' | 'loginHref'> & { search: string }

export function RegisterFormView({ search, ...props }: RegisterFormViewProps) {
  return (
    <AuthFormView
      {...props}
      mode="register"
      loginHref={`/login${search}`}
      footer={
        <>
          Đã có tài khoản?{' '}
          <Link to={`/login${search}`} className="font-semibold underline underline-offset-4 outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500">
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
