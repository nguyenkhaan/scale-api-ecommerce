import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { roleLabels, useAuth } from '../auth.hook'
import type { AuthState } from '../auth.type'

const focusRing = 'outline-offset-4 focus-visible:outline-2 focus-visible:outline-white'
const linkClass = `inline-flex min-h-11 items-center text-sm font-semibold text-white underline-offset-4 hover:underline ${focusRing}`

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
        <Link to="/login" className={linkClass}>Đăng nhập</Link>
        <Link to="/register" className={linkClass}>Đăng ký</Link>
      </>
    )
  }

  return (
    <>
      <Link to="/account" className={linkClass}>
        <span className="max-w-40 truncate sm:max-w-64">{state.user.email}</span>
        <span className="ml-2 border border-white/60 px-1.5 py-0.5 text-xs font-medium text-white">
          {roleLabels[state.user.role]}
        </span>
      </Link>
      <button type="button" onClick={onLogout} className={`min-h-11 border border-white px-3 text-sm font-semibold text-white hover:bg-red-600/30 ${focusRing}`}>
        Đăng xuất
      </button>
      {logoutFailed && <p role="alert" className="w-full text-sm font-semibold text-white sm:w-auto">Đăng xuất chưa thành công. Vui lòng thử lại.</p>}
    </>
  )
}

export function AuthNav() {
  const { state, logout } = useAuth()
  const navigate = useNavigate()
  const [logoutFailed, setLogoutFailed] = useState(false)

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
