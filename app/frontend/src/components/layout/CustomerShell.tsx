import { Outlet } from 'react-router-dom'
import { CustomerHeader } from './CustomerHeader'

export function CustomerShell() {
  return (
    <div className="min-h-screen bg-[var(--ss-color-app-bg)] text-[var(--ss-color-text)]">
      <CustomerHeader />
      <main id="main-content" className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 sm:py-8">
        <Outlet />
      </main>
    </div>
  )
}
