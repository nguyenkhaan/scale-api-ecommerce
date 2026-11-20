import { Button } from '@/components/common'
import type { User } from '@/features/auth/auth.type'

export type AdminHeaderProps = {
  user: User
  onLogout: () => void
}

export function AdminHeader({ user, onLogout }: AdminHeaderProps) {
  return (
    <header className="flex min-h-14 items-center justify-between gap-4 border-b border-[var(--ss-color-border)] bg-[var(--ss-color-surface)] px-4 sm:px-6">
      <div>
        <p className="text-xs font-semibold uppercase tracking-wider text-[var(--ss-color-muted)]">Khu vực quản trị</p>
        <p className="text-sm font-semibold text-[var(--ss-color-text)]">SaleScale</p>
      </div>
      <div className="flex min-w-0 items-center gap-3">
        <span className="max-w-40 truncate text-sm text-[var(--ss-color-muted)] sm:max-w-64">{user.email}</span>
        <Button variant="ghost" onClick={onLogout}>Đăng xuất</Button>
      </div>
    </header>
  )
}
