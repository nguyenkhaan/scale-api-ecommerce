import { Alert, Button, FeedbackState } from '@/components/common'
import { describeSimulatorError } from '../simulator.logic'
import { SimulatorSessionCard } from '../components/SimulatorSessionCard'
import { useScheduleSimulatorEvent, useSimulatorSessions } from '../payment.hook'
import type { ScheduledPaymentEvent, SimulatorEventRequest, SimulatorSession } from '../payment.type'

export type PaymentSimulatorViewProps = {
  sessions: SimulatorSession[]
  loading: boolean
  refreshing: boolean
  error: string | null
  onRefresh: () => void
  schedule: (body: SimulatorEventRequest) => Promise<ScheduledPaymentEvent>
}

export function PaymentSimulatorPage() {
  const query = useSimulatorSessions()
  const mutation = useScheduleSimulatorEvent()
  return (
    <PaymentSimulatorView
      sessions={query.data?.items ?? []}
      loading={query.isPending}
      refreshing={query.isFetching && !query.isPending}
      error={query.error ? describeSimulatorError(query.error) : null}
      onRefresh={() => void query.refetch()}
      schedule={mutation.mutateAsync}
    />
  )
}

export function PaymentSimulatorView({ sessions, loading, refreshing, error, onRefresh, schedule }: PaymentSimulatorViewProps) {
  return (
    <div className="mx-auto max-w-4xl space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-[var(--ss-color-border)] pb-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Mô phỏng thanh toán</h1>
          <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Gửi sự kiện giả lập từ cổng thanh toán tới hệ thống để thử các kết quả.</p>
        </div>
        <Button variant="secondary" loading={refreshing} onClick={onRefresh}>Làm mới</Button>
      </div>

      <Alert tone="warning">Chỉ dùng trong môi trường thử nghiệm</Alert>

      {loading && <FeedbackState kind="loading" title="Đang tải phiên thanh toán" message="Vui lòng chờ trong giây lát." />}
      {error && <FeedbackState kind="error" title="Không thể dùng công cụ mô phỏng" message={error} action={{ label: 'Thử lại', onClick: onRefresh }} />}
      {!loading && !error && sessions.length === 0 && (
        <FeedbackState kind="empty" title="Không có phiên thanh toán nào đang chờ." message="Tạo đơn hàng và mở trang thanh toán để có phiên mới, rồi làm mới danh sách." />
      )}

      {sessions.length > 0 && (
        <ul className="space-y-4">
          {sessions.map((session) => (
            <SimulatorSessionCard key={session.id} session={session} schedule={schedule} />
          ))}
        </ul>
      )}
    </div>
  )
}
