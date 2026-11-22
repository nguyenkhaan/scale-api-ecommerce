import { Button, Dialog } from '@/components/common'

type Props = {
  open: boolean
  busy: boolean
  onConfirm: () => void
  onClose: () => void
}

export function CancelOrderDialog({ open, busy, onConfirm, onClose }: Props) {
  return (
    <Dialog
      open={open}
      title="Hủy đơn hàng"
      onClose={busy ? () => undefined : onClose}
      footer={
        <>
          <Button variant="secondary" disabled={busy} onClick={onClose}>Giữ đơn</Button>
          <Button variant="danger" loading={busy} onClick={onConfirm}>Xác nhận hủy đơn</Button>
        </>
      }
    >
      <p className="text-sm text-[var(--ss-color-text)]">Bạn chắc chắn muốn hủy đơn này? Hàng đang giữ sẽ được trả lại kho và bạn không thể hoàn tác.</p>
    </Dialog>
  )
}
