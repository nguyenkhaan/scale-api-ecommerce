import { describe, expect, it, mock } from 'bun:test'
import type { ReactElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import {
  FormField,
  PriceBlock,
  QuantityStepper,
  SearchField,
  StatusBadge,
} from '@/components/common'

describe('shared domain components', () => {
  it('uses audience-specific flash-sale labels with a visible status icon', () => {
    const customerMarkup = renderToStaticMarkup(
      <StatusBadge domain="flashSale" value="scheduled" audience="customer" />,
    )
    const adminMarkup = renderToStaticMarkup(<StatusBadge domain="flashSale" value="scheduled" audience="admin" />)

    expect(customerMarkup).toContain('Sắp diễn ra')
    expect(adminMarkup).toContain('Đã lên lịch')
    expect(customerMarkup).toContain('data-status-icon=')
    expect(customerMarkup).toContain('aria-hidden="true"')
  })

  it('renders the sale price, old price, source and end time distinctly', () => {
    const markup = renderToStaticMarkup(
      <PriceBlock
        regularPrice={199000}
        effectivePrice={129000}
        source="flash_sale"
        endsAt="2026-11-11T14:00:00.000Z"
      />,
    )

    expect(markup).toContain('129.000')
    expect(markup).toContain('199.000')
    expect(markup).toContain('Flash sale')
    expect(markup).toContain('Kết thúc')
  })

  it('disables quantity actions at bounds and while busy', () => {
    const atMinimum = renderToStaticMarkup(
      <QuantityStepper value={1} min={1} max={5} busy={false} label="Số lượng Áo" onChange={() => {}} />,
    )
    const atMaximum = renderToStaticMarkup(
      <QuantityStepper value={5} min={1} max={5} busy={false} label="Số lượng Áo" onChange={() => {}} />,
    )
    const busy = renderToStaticMarkup(
      <QuantityStepper value={2} min={1} max={5} busy label="Số lượng Áo" onChange={() => {}} />,
    )

    expect(atMinimum).toContain('role="group"')
    expect(atMinimum).toContain('aria-label="Giảm Số lượng Áo" disabled=""')
    expect(atMaximum).toContain('aria-label="Tăng Số lượng Áo" disabled=""')
    expect(busy.match(/disabled=""/g)?.length).toBe(2)
  })

  it('connects form errors to the input through aria attributes', () => {
    const markup = renderToStaticMarkup(
      <FormField id="email" label="Email" errorMessages={['Email chưa đúng định dạng']} helper="Dùng email đăng nhập.">
        {(field) => <input id="email" type="email" {...field} />}
      </FormField>,
    )

    expect(markup).toContain('for="email"')
    expect(markup).toContain('aria-invalid="true"')
    expect(markup).toContain('aria-describedby="email-description email-errors"')
    expect(markup).toContain('Email chưa đúng định dạng')
  })

  it('submits search through a native form and passes the current value', () => {
    const onSubmit = mock(() => {})
    const markup = renderToStaticMarkup(
      <SearchField value="áo" placeholder="Tìm sản phẩm" onChange={() => {}} onSubmit={onSubmit} />,
    )
    const form = SearchField({ value: 'áo', placeholder: 'Tìm sản phẩm', onChange: () => {}, onSubmit }) as ReactElement<{
      onSubmit: (event: { preventDefault: () => void }) => void
    }>

    expect(markup).toContain('role="search"')
    expect(markup).toContain('value="áo"')
    form.props.onSubmit({ preventDefault: () => {} })
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })
})
