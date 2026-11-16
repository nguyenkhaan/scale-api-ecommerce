import { Navigate, Route, Routes } from 'react-router-dom'

import { RequireAuth } from '@/features/auth/components/RequireAuth'
import { AccountPage } from '@/features/auth/pages/AccountPage'
import { LoginPage } from '@/features/auth/pages/LoginPage'
import { RegisterPage } from '@/features/auth/pages/RegisterPage'
import { ProductsPage } from '@/features/catalog/pages/ProductsPage'
import { CartPage } from '@/features/cart/pages/CartPage'
import { RequireRole } from '@/features/auth/components/RequireRole'
import { AdminFlashSalePage } from '@/features/admin-flashsale/pages/AdminFlashSalePage'
import { AdminPage } from '@/features/admin-flashsale/pages/AdminPage'
import { FlashSalesPage } from '@/features/flashsale/pages/FlashSalesPage'
import { HealthPage } from '@/features/health/pages/HealthPage'

export function AppRoute() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequireAuth>
            <Navigate to="/account" replace />
          </RequireAuth>
        }
      />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/flash-sales" element={<FlashSalesPage />} />
      <Route path="/health" element={<HealthPage />} />
      <Route
        path="/products"
        element={
          <RequireAuth>
            <ProductsPage />
          </RequireAuth>
        }
      />
      <Route
        path="/cart"
        element={
          <RequireAuth>
            <CartPage />
          </RequireAuth>
        }
      />
      <Route
        path="/account"
        element={
          <RequireAuth>
            <AccountPage />
          </RequireAuth>
        }
      />
      <Route path="/admin" element={<RequireRole role="admin"><AdminPage /></RequireRole>} />
      <Route path="/admin/flash-sales" element={<RequireRole role="admin"><AdminFlashSalePage /></RequireRole>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
