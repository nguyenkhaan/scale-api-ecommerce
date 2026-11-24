import { orderTabs, type OrderTabKey } from '../bill.logic'

type Props = {
  active: OrderTabKey
  onChange: (tab: OrderTabKey) => void
}

export function OrderStatusTabs({ active, onChange }: Props) {
  return (
    <nav aria-label="Lọc đơn hàng theo trạng thái" className="-mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0">
      <ul className="flex min-w-max gap-2">
        {orderTabs.map((tab) => {
          const selected = tab.key === active
          return (
            <li key={tab.key}>
              <button
                type="button"
                aria-pressed={selected}
                onClick={() => onChange(tab.key)}
                className={`min-h-11 whitespace-nowrap rounded-[var(--ss-radius-control)] border px-3 text-sm font-semibold ${selected ? 'border-[var(--ss-color-primary)] bg-red-50 text-[var(--ss-color-primary-strong)]' : 'border-[var(--ss-color-input-border)] bg-white text-[var(--ss-color-muted)] hover:bg-[var(--ss-color-app-bg)]'}`}
              >
                {tab.label}
              </button>
            </li>
          )
        })}
      </ul>
    </nav>
  )
}
