import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { HealthStatusView } from './HealthPage'

describe('HealthStatusView', () => {
  it('announces the loading state', () => {
    const markup = renderToStaticMarkup(
      <HealthStatusView state={{ type: 'loading' }} onRetry={() => undefined} />,
    )

    expect(markup).toContain('aria-busy="true"')
    expect(markup).toContain('Checking backend readiness')
  })

  it('shows the successful connection state', () => {
    const markup = renderToStaticMarkup(
      <HealthStatusView
        state={{ type: 'success', requestID: 'request-123' }}
        onRetry={() => undefined}
      />,
    )

    expect(markup).toContain('Backend connected')
    expect(markup).toContain('request-123')
    expect(markup).toContain('role="status"')
  })

  it('renders a keyboard-accessible retry action for network errors', () => {
    const markup = renderToStaticMarkup(
      <HealthStatusView state={{ type: 'error' }} onRetry={() => undefined} />,
    )

    expect(markup).toContain('role="alert"')
    expect(markup).toContain('<button type="button"')
    expect(markup).toContain('Try again')
  })
})
