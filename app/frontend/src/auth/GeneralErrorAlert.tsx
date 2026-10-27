import { Link } from 'react-router-dom'
import type { GeneralError } from './auth-form-errors'

type Props = {
  error: GeneralError
  loginHref: string
}

export function GeneralErrorAlert({ error, loginHref }: Props) {
  return (
    <div role="alert" className="mb-5 border-l-4 border-rose-600 bg-rose-50 p-4 text-sm text-rose-900">
      <GeneralErrorBody error={error} loginHref={loginHref} />
    </div>
  )
}

function GeneralErrorBody({ error, loginHref }: Props) {
  switch (error.kind) {
    case 'invalid-credentials':
      return <p>Email hoặc mật khẩu không đúng</p>
    case 'rate-limited':
      return (
        <p>
          {error.retryAfterSeconds === undefined
            ? 'Bạn đã thử quá nhiều lần. Vui lòng thử lại sau ít phút.'
            : `Bạn đã thử quá nhiều lần. Vui lòng thử lại sau ${error.retryAfterSeconds} giây.`}
        </p>
      )
    case 'email-taken':
      return (
        <p>
          Email này đã được đăng ký.{' '}
          <Link to={loginHref} className="font-semibold underline underline-offset-4 outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500">
            Đăng nhập
          </Link>
        </p>
      )
    case 'unexpected':
      return (
        <>
          <p>Đã có lỗi xảy ra. Vui lòng thử lại.</p>
          {error.requestID !== undefined && (
            <p className="mt-2 break-all font-mono text-xs text-rose-800">
              Mã yêu cầu: <span className="select-all">{error.requestID}</span>
            </p>
          )}
        </>
      )
  }
}
