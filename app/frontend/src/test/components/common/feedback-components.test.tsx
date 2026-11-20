import { describe, expect, it, mock } from 'bun:test'
import type { ReactElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { Alert, Button, Dialog, FeedbackState, IconButton } from '@/components/common'

describe('shared feedback components', () => {
  it('renders a primary button with disabled and loading semantics', () => {
    const disabledMarkup = renderToStaticMarkup(
      <Button variant="primary" disabled>
        Lưu
      </Button>,
    )
    const loadingMarkup = renderToStaticMarkup(<Button loading>Đăng nhập</Button>)

    expect(disabledMarkup).toContain('data-variant="primary"')
    expect(disabledMarkup).toContain('disabled=""')
    expect(loadingMarkup).toContain('aria-busy="true"')
    expect(loadingMarkup).toContain('disabled=""')
    expect(loadingMarkup).toContain('Đang xử lý')
  })

  it('exposes the supplied accessible name from an icon button', () => {
    const markup = renderToStaticMarkup(
      <IconButton label="Mở giỏ hàng">
        <span aria-hidden="true">cart</span>
      </IconButton>,
    )

    expect(markup).toContain('aria-label="Mở giỏ hàng"')
    expect(markup).toContain('min-h-11')
  })

  it('uses an alert role for danger messages', () => {
    const markup = renderToStaticMarkup(<Alert tone="danger">Không thể lưu thay đổi.</Alert>)

    expect(markup).toContain('role="alert"')
    expect(markup).toContain('Không thể lưu thay đổi.')
  })

  it('uses status for loading and alert for errors', () => {
    const loadingMarkup = renderToStaticMarkup(
      <FeedbackState kind="loading" title="Đang tải" message="Vui lòng chờ." />,
    )
    const errorMarkup = renderToStaticMarkup(
      <FeedbackState kind="error" title="Tải thất bại" message="Thử lại sau." />,
    )

    expect(loadingMarkup).toContain('role="status"')
    expect(errorMarkup).toContain('role="alert"')
  })

  it('renders an accessible dialog and closes when Escape is pressed', () => {
    const onClose = mock(() => {})
    const markup = renderToStaticMarkup(
      <Dialog open title="Xác nhận" onClose={onClose}>
        Nội dung
      </Dialog>,
    )

    expect(markup).toContain('role="dialog"')
    expect(markup).toContain('aria-modal="true"')
    expect(markup).toContain('aria-label="Đóng"')

    const element = Dialog({ open: true, title: 'Xác nhận', onClose, children: 'Nội dung' }) as ReactElement<{
      children: ReactElement<{
        onKeyDown: (event: { key: string; preventDefault: () => void }) => void
      }>
    }>
    const panel = element.props.children
    panel.props.onKeyDown({ key: 'Escape', preventDefault: () => {} })

    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
