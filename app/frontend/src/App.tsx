import { BrowserRouter, Link } from 'react-router-dom'

import { AuthNav } from '@/features/auth/components/AuthNav'
import { AppRoute } from '@/routes/AppRoute'

export default function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-stone-100 text-slate-950">
        <header className="border-b border-slate-300 bg-white">
          <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-4 px-4 py-4 sm:px-6">
            <Link to="/" className="font-bold tracking-tight outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500">
              SaleScale
            </Link>
            <nav aria-label="Primary navigation" className="flex flex-wrap items-center justify-end gap-x-4 gap-y-1">
              <AuthNav />
            </nav>
          </div>
        </header>
        <main className="mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 sm:py-16">
          <AppRoute />
        </main>
      </div>
    </BrowserRouter>
  )
}
