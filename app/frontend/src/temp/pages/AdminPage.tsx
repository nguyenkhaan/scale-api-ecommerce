import { Link } from 'react-router-dom'

export function AdminPage() {
  return (
    <section aria-labelledby="admin-title" className="mx-auto w-full max-w-3xl">
      <h1 id="admin-title" className="text-3xl font-bold tracking-tight text-slate-950">
        Quản trị
      </h1>
      <p className="mt-3 text-slate-600">Khu vực dành cho quản trị viên hệ thống.</p>
      
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <Link
          to="/admin/flash-sales"
          className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm hover:border-amber-500 hover:shadow-md transition"
        >
          <h2 className="text-lg font-bold text-slate-900">⚡ Quản lý Flash Sale</h2>
          <p className="mt-2 text-xs text-slate-600">
            Tạo chiến dịch mới, lên lịch, hủy lịch hoặc dừng chiến dịch bán hàng chớp nhoáng.
          </p>
        </Link>
      </div>
    </section>
  )
}
