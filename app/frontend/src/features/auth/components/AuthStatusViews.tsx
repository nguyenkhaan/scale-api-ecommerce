export function AuthLoadingView() {
  return (
    <div aria-busy="true" role="status" className="border-l-4 border-slate-400 bg-white p-5">
      <p className="font-semibold text-slate-950">Đang kiểm tra phiên đăng nhập</p>
      <p className="mt-1 text-sm text-slate-600">Vui lòng chờ trong giây lát.</p>
    </div>
  )
}

export function ForbiddenView() {
  return (
    <section aria-labelledby="forbidden-title" className="border-l-4 border-rose-600 bg-white p-5">
      <h1 id="forbidden-title" className="text-xl font-bold text-rose-800">
        403
      </h1>
      <p className="mt-1 text-slate-700">Bạn không có quyền truy cập trang này</p>
    </section>
  )
}
