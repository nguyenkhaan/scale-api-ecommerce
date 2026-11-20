import { useState } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { AdminHeader } from './AdminHeader'
import { AdminSidebar } from './AdminSidebar'
import { useAuth } from '@/features/auth/auth.hook'

export function AdminShell() {
  const { state, logout } = useAuth()
  const navigate = useNavigate()
  const [logoutFailed, setLogoutFailed] = useState(false)

  if (state.status !== 'authenticated') return <Outlet />

  async function handleLogout() {
    setLogoutFailed(false)
    try {
      await logout()
      navigate('/', { replace: true })
    } catch {
      setLogoutFailed(true)
    }
  }

  return (
    <div className="flex min-h-screen flex-col bg-[var(--ss-color-app-bg)] text-[var(--ss-color-text)] md:flex-row">
      <AdminSidebar />
      <div className="min-w-0 flex-1">
        <AdminHeader user={state.user} onLogout={() => void handleLogout()} />
        {logoutFailed && <p role="alert" className="border-b border-red-200 bg-red-50 px-4 py-2 text-sm text-[var(--ss-color-danger)] sm:px-6">Đăng xuất chưa thành công. Vui lòng thử lại.</p>}
        <main id="admin-main-content" className="mx-auto w-full max-w-7xl p-4 sm:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
