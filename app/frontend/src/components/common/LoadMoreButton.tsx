import { Button } from './Button'

export type LoadMoreButtonProps = {
  loading: boolean
  disabled?: boolean
  onClick: () => void
}

export function LoadMoreButton({ loading, disabled = false, onClick }: LoadMoreButtonProps) {
  return (
    <Button variant="secondary" loading={loading} disabled={disabled} onClick={onClick}>
      Tải thêm
    </Button>
  )
}
