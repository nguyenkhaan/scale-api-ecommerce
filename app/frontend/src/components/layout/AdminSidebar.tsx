import { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'

const links = [
  { to: '/admin', label: 'Tổng quan', end: true },
  { to: '/admin/flash-sales', label: 'Quản lý Flash Sale', end: false },
]

function linkClass(active: boolean) {
  return `flex min-h-11 items-center border-l-2 px-4 text-sm font-semibold ${active ? 'border-[var(--ss-color-primary)] bg-red-50 text-[var(--ss-color-primary-strong)]' : 'border-transparent text-[var(--ss-color-muted)] hover:bg-[var(--ss-color-app-bg)] hover:text-[var(--ss-color-text)]'}`
}

export function AdminSidebar() {
  const location = useLocation()
  const [open, setOpen] = useState(false)

  return (
    <aside className="w-full shrink-0 border-b border-[var(--ss-color-border)] bg-[var(--ss-color-surface)] md:min-h-screen md:w-60 md:border-b-0 md:border-r">
      <div className="flex min-h-14 items-center justify-between px-4 md:px-5">
        <NavLink to="/admin" className="font-bold tracking-tight text-[var(--ss-color-text)]">SaleScale Admin</NavLink>
        <button
          type="button"
          aria-expanded={open}
          aria-controls="admin-navigation"
          onClick={() => setOpen((value) => !value)}
          className="min-h-11 min-w-11 border border-[var(--ss-color-input-border)] px-2 text-sm font-semibold md:hidden"
        >
          Menu
        </button>
      </div>
      <nav id="admin-navigation" aria-label="Admin navigation" className={`${open ? 'block' : 'hidden md:block'} border-t border-[var(--ss-color-border)] py-2 md:border-t-0`}>
        {links.map((link) => {
          const active = link.end ? location.pathname === link.to : location.pathname.startsWith(link.to)
          return (
            <NavLink key={link.to} to={link.to} end={link.end} className={() => linkClass(active)}>
              {link.label}
            </NavLink>
          )
        })}
      </nav>
    </aside>
  )
}
