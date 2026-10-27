import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import type { AuthState } from './auth-context'
import { roleLabels } from './role-labels'
import { useAuth } from './use-auth'

const focusRing = 'outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500'
const linkClass = `inline-flex min-h-11 items-center text-sm font-semibold text-slate-700 underline-offset-4 hover:underline ${focusRing}`

type AuthNavViewProps = {
  state: AuthState
  onLogout: () => void
  logoutFailed?: boolean
}

export function AuthNavView({ state, onLogout, logoutFailed = false }: AuthNavViewProps) {
  if (state.status === 'loading') return null

  if (state.status === 'anonymous') {
    return (
      <>
        <Link to="/login" className={linkClass}>
          Đăng nhập
        </Link>
        <Link to="/register" className={linkClass}>
          Đăng ký
        </Link>
      </>
    )
  }

  return (
    <>
      <Link to="/account" className={linkClass}>
        <span className="max-w-40 truncate sm:max-w-64">{state.user.email}</span>
        <span className="ml-2 border border-slate-300 px-1.5 py-0.5 text-xs font-medium text-slate-600">
          {roleLabels[state.user.role]}
        </span>
      </Link>
      <button
        type="button"
        onClick={onLogout}
        className={`min-h-11 border border-slate-900 px-3 text-sm font-semibold text-slate-900 hover:bg-stone-100 ${focusRing}`}
      >
        Đăng xuất
      </button>
      {logoutFailed && (
        <p role="alert" className="w-full text-sm font-semibold text-rose-700 sm:w-auto">
          Đăng xuất chưa thành công. Vui lòng thử lại.
        </p>
      )}
    </>
  )
}

export function AuthNav() {
  const { state, logout } = useAuth()
  const navigate = useNavigate()
  const [logoutFailed, setLogoutFailed] = useState(false)
  // Leave the page first so a protected route does not turn an explicit logout into /login?next=<that page>.
  const handleLogout = async () => {
    setLogoutFailed(false)
    navigate('/', { replace: true })
    try {
      await logout()
    } catch {
      setLogoutFailed(true)
    }
  }
  return <AuthNavView state={state} onLogout={() => void handleLogout()} logoutFailed={logoutFailed} />
}
