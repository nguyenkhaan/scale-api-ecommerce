import { Link } from 'react-router-dom'
import type { CartActionNoticeState } from './use-add-to-cart'

type Props = {
  notice: CartActionNoticeState | null
  /** Where to return after logging in, e.g. `/products`. */
  loginNext: string
}

export function CartActionNotice({ notice, loginNext }: Props) {
  if (!notice) return null
  const isError = notice.kind === 'error'
  return (
    <div
      role={isError ? 'alert' : 'status'}
      className={`mt-4 rounded border p-4 text-sm font-medium ${
        isError ? 'border-rose-200 bg-rose-50 text-rose-800' : 'border-amber-300 bg-amber-50 text-amber-900'
      }`}
    >
      <p>{notice.text}</p>
      {notice.loginRequired && (
        <Link
          to={`/login?next=${encodeURIComponent(loginNext)}`}
          className="mt-1 inline-block font-semibold underline"
        >
          Đăng nhập
        </Link>
      )}
      {notice.kind === 'success' && (
        <Link to="/cart" className="mt-1 inline-block font-semibold underline">
          Xem giỏ hàng
        </Link>
      )}
    </div>
  )
}
