import { Link } from 'react-router-dom'
import { Surface } from '@/components/common'

export function AdminPage() {
  return (
    <section aria-labelledby="admin-title" className="mx-auto w-full max-w-4xl">
      <h1 id="admin-title" className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">
        Quản trị
      </h1>
      <p className="mt-2 text-[var(--ss-color-muted)]">Khu vực dành cho quản trị viên hệ thống.</p>
      
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <Surface as="article" className="p-4">
        <Link
          to="/admin/flash-sales"
          className="block min-h-11 text-[var(--ss-color-primary-strong)] hover:underline"
        >
          <h2 className="text-lg font-semibold">Quản lý Flash Sale</h2>
          <p className="mt-2 text-xs text-[var(--ss-color-muted)]">
            Tạo chiến dịch mới, lên lịch, hủy lịch hoặc dừng chiến dịch bán hàng chớp nhoáng.
          </p>
        </Link>
        </Surface>
      </div>
    </section>
  )
}
