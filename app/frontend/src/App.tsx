import { BrowserRouter, Link, Navigate, Route, Routes } from 'react-router-dom'
import { HealthPage } from './pages/HealthPage'

function HomePage() {
  return (
    <section aria-labelledby="home-title" className="max-w-3xl">
      <p className="text-sm font-semibold uppercase tracking-widest text-amber-700">Phase 00</p>
      <h1 id="home-title" className="mt-2 text-4xl font-bold tracking-tight text-slate-950 sm:text-5xl">
        SaleScale foundation
      </h1>
      <p className="mt-5 max-w-2xl text-lg leading-8 text-slate-600">
        A Go and React foundation for testing correct multi-SKU checkout under concurrent flash-sale traffic.
      </p>
      <Link
        to="/health"
        className="mt-8 inline-flex min-h-11 items-center border border-slate-900 bg-slate-900 px-4 py-2 text-sm font-semibold text-white outline-offset-4 hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-amber-500"
      >
        Check backend connection
      </Link>
    </section>
  )
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/health" element={<HealthPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

function AppLayout() {
  return (
    <div className="min-h-screen bg-stone-100 text-slate-950">
      <header className="border-b border-slate-300 bg-white">
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-4 px-4 py-4 sm:px-6">
          <Link to="/" className="font-bold tracking-tight outline-offset-4 focus-visible:outline-2 focus-visible:outline-amber-500">
            SaleScale
          </Link>
          <nav aria-label="Primary navigation">
            <Link to="/health" className="text-sm font-semibold text-slate-700 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-amber-500">
              System health
            </Link>
          </nav>
        </div>
      </header>
      <main className="mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 sm:py-16">
        <AppRoutes />
      </main>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AppLayout />
    </BrowserRouter>
  )
}
