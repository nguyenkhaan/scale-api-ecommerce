import { BrowserRouter, Link, Navigate, Route, Routes } from 'react-router-dom'
import { AuthNav } from './auth/AuthNav'
import { AuthProvider } from './auth/AuthProvider'
import { RequireAuth } from './auth/RequireAuth'
import { RequireRole } from './auth/RequireRole'
import { AccountPage } from './pages/AccountPage'
import { AdminFlashSalePage } from './pages/AdminFlashSalePage'
import { AdminPage } from './pages/AdminPage'
import { CartPage } from './pages/CartPage'
import { FlashSalesPage } from './pages/FlashSalesPage'
import { HealthPage } from './pages/HealthPage'
import { LoginPage } from './pages/LoginPage'
import { ProductsPage } from './pages/ProductsPage'
import { RegisterPage } from './pages/RegisterPage'

function HomePage() {
  return (
    <section aria-labelledby="home-title" className="max-w-3xl">
      <p className="text-sm font-semibold uppercase tracking-widest text-amber-700">Phase 03</p>
      <h1 id="home-title" className="mt-2 text-4xl font-bold tracking-tight text-slate-950 sm:text-5xl">
        SaleScale Flash Sale & Giỏ hàng
      </h1>
      <p className="mt-5 max-w-2xl text-lg leading-8 text-slate-600">
        Khám phá các ưu đãi Flash Sale chớp nhoáng, quản lý giỏ hàng đa SKU và xem trước tính giá thời gian thực.
      </p>
      <div className="mt-8 flex flex-wrap gap-4">
        <Link
          to="/flash-sales"
          className="inline-flex min-h-11 items-center border border-rose-600 bg-rose-600 px-4 py-2 text-sm font-semibold text-white outline-offset-4 hover:bg-rose-500 focus-visible:outline-2 focus-visible:outline-rose-500"
        >
          ⚡ Xem Flash Sale
        </Link>
        <Link
          to="/cart"
          className="inline-flex min-h-11 items-center border border-slate-900 bg-slate-900 px-4 py-2 text-sm font-semibold text-white outline-offset-4 hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-amber-500"
        >
          Giỏ hàng của bạn
        </Link>
      </div>
    </section>
  )
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/flash-sales" element={<FlashSalesPage />} />
      <Route path="/products" element={<RequireAuth><ProductsPage /></RequireAuth>} />
      <Route path="/cart" element={<RequireAuth><CartPage /></RequireAuth>} />
      <Route path="/health" element={<HealthPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/account" element={<RequireAuth><AccountPage /></RequireAuth>} />
      <Route path="/admin" element={<RequireRole role="admin"><AdminPage /></RequireRole>} />
      <Route path="/admin/flash-sales" element={<RequireRole role="admin"><AdminFlashSalePage /></RequireRole>} />
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
          <nav aria-label="Primary navigation" className="flex flex-wrap items-center justify-end gap-x-4 gap-y-1">
            <Link to="/flash-sales" className="inline-flex min-h-11 items-center text-sm font-semibold text-rose-700 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-amber-500">
              Flash Sale
            </Link>
            <Link to="/products" className="inline-flex min-h-11 items-center text-sm font-semibold text-slate-700 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-amber-500">
              Sản phẩm
            </Link>
            <Link to="/cart" className="inline-flex min-h-11 items-center text-sm font-semibold text-slate-700 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-amber-500">
              Giỏ hàng
            </Link>
            <Link to="/health" className="inline-flex min-h-11 items-center text-sm font-semibold text-slate-700 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-amber-500">
              System health
            </Link>
            <AuthNav />
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
      <AuthProvider>
        <AppLayout />
      </AuthProvider>
    </BrowserRouter>
  )
}
