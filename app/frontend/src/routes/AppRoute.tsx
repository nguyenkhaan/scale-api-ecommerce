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
import { AdminShell, CustomerShell } from '@/components/layout'

export function AppRoute() {
  return (
    <Routes>
      <Route element={<CustomerShell />}>
        <Route index element={<ProductsPage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
        <Route path="flash-sales" element={<FlashSalesPage />} />
        <Route path="products" element={<ProductsPage />} />
        <Route
          path="cart"
          element={
            <RequireAuth>
              <CartPage />
            </RequireAuth>
          }
        />
        <Route
          path="account"
          element={
            <RequireAuth>
              <AccountPage />
            </RequireAuth>
          }
        />
      </Route>
      <Route path="/health" element={<HealthPage />} />
      <Route path="admin" element={<RequireRole role="admin"><AdminShell /></RequireRole>}>
        <Route index element={<AdminPage />} />
        <Route path="flash-sales" element={<AdminFlashSalePage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
