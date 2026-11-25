import { Alert, Button } from '@/components/common'

import type { CheckoutOutcome } from '../checkout.type'

type Props = {
  outcome: CheckoutOutcome
  onRetry: () => void
  onCheck: () => void
  retryCountdown: number | null
}

export function CheckoutOutcomeView({ outcome, onRetry, onCheck, retryCountdown }: Props) {
  switch (outcome.kind) {
    case 'unknown':
      return (
        <Outcome tone="warning" title={outcome.requestID ? 'Hệ thống tạm thời không xử lý được' : 'Đang kiểm tra kết quả đặt đơn…'}>
          <p>{outcome.requestID ? 'Hệ thống chưa thể xác nhận kết quả đặt đơn.' : 'Kết nối bị gián đoạn nên chưa nhận được kết quả. Hệ thống đang tự kiểm tra lại, bạn không cần đặt lại đơn.'}</p>
          {outcome.requestID && <p className="mt-1 text-xs">Mã yêu cầu: {outcome.requestID}</p>}
          <div className="mt-3 flex flex-wrap gap-2">
            <Button variant="secondary" onClick={onCheck}>Kiểm tra lại</Button>
            <Button variant="ghost" onClick={onRetry}>Thử lại</Button>
          </div>
        </Outcome>
      )
    case 'processing':
      return (
        <Outcome tone="info" title="Yêu cầu trước đó đang được xử lý, vui lòng đợi.">
          <p>Hệ thống đang hoàn tất yêu cầu trước đó. Bạn không cần đặt lại đơn.</p>
          <Button variant="secondary" onClick={onCheck}>Kiểm tra lại</Button>
        </Outcome>
      )
    case 'overload':
      return (
        <Outcome tone="warning" title="Hệ thống đang quá tải">
          <p>Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên.</p>
          <Button variant="secondary" disabled={(retryCountdown ?? outcome.retryAfterSeconds) > 0} onClick={onRetry}>
            {(retryCountdown ?? outcome.retryAfterSeconds) > 0 ? `Thử lại sau ${formatCountdown(retryCountdown ?? outcome.retryAfterSeconds)}` : 'Thử lại'}
          </Button>
        </Outcome>
      )
    case 'rejected':
      return (
        <Outcome tone="danger" title="Chưa đặt được đơn hàng">
          <p>Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên.</p>
        </Outcome>
      )
    case 'price_changed':
      return (
        <Outcome tone="warning" title="Giá đã thay đổi">
          <p>Giá chính thức của một hoặc nhiều dòng đã thay đổi. Vui lòng xem lại trước khi đặt hàng.</p>
          <Button variant="secondary" onClick={onRetry}>Tôi đồng ý với giá mới</Button>
        </Outcome>
      )
    case 'cart_changed':
      return (
        <Outcome tone="warning" title="Giỏ hàng đã thay đổi">
          <p>Giỏ hàng vừa thay đổi. Vui lòng tải lại và xem trước lại.</p>
          <Button variant="secondary" onClick={onRetry}>Tải lại giỏ hàng</Button>
        </Outcome>
      )
    case 'intent_conflict':
      return (
        <Outcome tone="warning" title="Phiên đặt hàng đã thay đổi">
          <p>Vui lòng xem trước lại giỏ hàng rồi thử đặt đơn lần nữa.</p>
        </Outcome>
      )
    case 'invalid':
      return (
        <Outcome tone="danger" title="Thông tin đặt hàng chưa hợp lệ">
          <p>Vui lòng kiểm tra lại các thông tin người nhận.</p>
        </Outcome>
      )
    case 'unauthenticated':
      return (
        <Outcome tone="warning" title="Phiên đăng nhập đã hết hạn">
          <p>Vui lòng đăng nhập lại để tiếp tục đặt hàng.</p>
        </Outcome>
      )
    case 'created':
      return null
  }
}

function formatCountdown(seconds: number): string {
  const minutes = Math.floor(seconds / 60).toString().padStart(2, '0')
  const remainder = (seconds % 60).toString().padStart(2, '0')
  return `${minutes}:${remainder}`
}

function Outcome({ tone, title, children }: { tone: 'info' | 'success' | 'warning' | 'danger'; title: string; children: React.ReactNode }) {
  return (
    <Alert tone={tone}>
      <p className="font-semibold">{title}</p>
      <div className="mt-1">{children}</div>
    </Alert>
  )
}
