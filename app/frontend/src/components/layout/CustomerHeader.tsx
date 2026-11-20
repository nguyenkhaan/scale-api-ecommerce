import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { SearchField } from '@/components/common'
import { AuthNav } from '@/features/auth/components/AuthNav'

export function CustomerHeader() {
  const navigate = useNavigate()
  const location = useLocation()
  const initialQuery = new URLSearchParams(location.search).get('query') ?? ''
  const [query, setQuery] = useState(initialQuery)

  function submitSearch() {
    const value = query.trim()
    navigate(value === '' ? '/products' : `/products?query=${encodeURIComponent(value)}`)
  }

  return (
    <header className="bg-[var(--ss-color-primary)] text-white">
      <div className="mx-auto flex w-full max-w-7xl flex-wrap items-center gap-3 px-4 py-3 sm:px-6">
        <Link to="/" className="shrink-0 text-lg font-bold tracking-tight outline-offset-4 focus-visible:outline-white">
          SaleScale
        </Link>
        <div className="order-3 min-w-0 basis-full md:order-none md:flex-1">
          <SearchField value={query} placeholder="Tìm sản phẩm" onChange={setQuery} onSubmit={submitSearch} label="Tìm sản phẩm" />
        </div>
        <nav aria-label="Customer navigation" className="ml-auto flex min-h-11 items-center gap-1 sm:gap-2">
          <Link to="/flash-sales" className="inline-flex min-h-11 items-center px-2 text-sm font-semibold hover:bg-red-600/30 focus-visible:outline-white">Flash Sale</Link>
          <Link to="/cart" className="inline-flex min-h-11 items-center px-2 text-sm font-semibold hover:bg-red-600/30 focus-visible:outline-white">Giỏ hàng</Link>
          <AuthNav />
        </nav>
      </div>
    </header>
  )
}
