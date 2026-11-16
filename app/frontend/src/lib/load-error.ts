import { isApiError } from './api-error'

/** Vietnamese copy for a failed read; never exposes raw error codes. */
export function describeLoadError(err: unknown, what: string): string {
  if (isApiError(err)) {
    if (err.code === 'NETWORK_ERROR') return `Lỗi kết nối khi tải ${what}.`
    if (err.status === 401) return `Vui lòng đăng nhập để xem ${what}.`
    if (err.status === 403) return `Bạn không có quyền xem ${what}.`
    if (err.status === 429) return `Bạn thao tác quá nhanh khi tải ${what}. Vui lòng thử lại sau.`
  }
  return `Không thể tải ${what}. Vui lòng thử lại.`
}
