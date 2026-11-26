import { expect, test } from '@playwright/test'

type Product = { id: string; name: string; category_id: string | null }

test('public catalog shows active products, filters categories and hides inactive products', async ({ page, request }) => {
  const response = await request.get('/api/products?query=Seed&limit=100')
  expect(response.ok(), 'seeded catalog must be available').toBeTruthy()
  const catalog = await response.json() as { items: Product[] }
  const activeProduct = catalog.items.find((product) => product.name === 'Tai nghe Seed')
  const categoryProduct = catalog.items.find((product) => product.name === 'Áo Seed')
  expect(activeProduct, 'development fixture must include an active product').toBeDefined()
  expect(categoryProduct, 'development fixture must include an active product in another category').toBeDefined()

  await page.goto('/products')
  await page.getByRole('search').nth(1).getByRole('searchbox').fill('Seed')
  await page.getByRole('search').nth(1).getByRole('button', { name: 'Tìm' }).click()
  await expect(page.getByRole('link', { name: 'Tai nghe Seed' })).toBeVisible()
  await expect(page.getByText('Sản phẩm ẩn Seed')).toHaveCount(0)

  await page.getByLabel('Lọc theo danh mục').selectOption(categoryProduct!.category_id ?? '')
  await expect(page.getByRole('link', { name: 'Áo Seed' })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Tai nghe Seed' })).toHaveCount(0)

  await page.goto(`/products/${activeProduct!.id}`)
  await expect(page.getByRole('heading', { name: 'Tai nghe Seed' })).toBeVisible()
  await expect(page.getByRole('combobox', { name: 'SKU' })).toContainText('SEED-HEADPHONE-01')

  await page.goto('/products/13000000-0000-4000-8000-000000000004')
  await expect(page.getByText('Không tìm thấy sản phẩm', { exact: true })).toBeVisible()
})

test('customer can place an order through the browser and cancel it to release the reservation', async ({ page, request }) => {
  const email = `browser-checkout-${Date.now()}@example.test`
  const password = 'browser-test-password-42'
  const registration = await request.post('/api/auth/register', { data: { email, password } })
  expect(registration.status(), 'create an isolated browser test customer').toBe(201)

  let orderID: string | undefined
  try {
    await page.goto('/login')
    await page.locator('#email').fill(email)
    await page.locator('#password').fill(password)
    await page.getByRole('button', { name: 'Đăng nhập' }).click()
    await expect(page.getByRole('link', { name: 'Đăng nhập' })).toHaveCount(0)

    await page.goto('/products/13000000-0000-4000-8000-000000000001')
    await page.getByRole('button', { name: 'Thêm vào giỏ' }).click()
    await expect(page.getByText(/Đã thêm "Tai nghe Seed" vào giỏ hàng/)).toBeVisible()
    await page.goto('/cart')
    await expect(page.getByText('Tai nghe Seed')).toBeVisible()
    await page.getByRole('button', { name: 'Xem trước thanh toán' }).click()
    await page.getByRole('textbox').nth(0).fill('Người nhận kiểm thử')
    await page.getByRole('textbox').nth(1).fill('0901234567')
    await page.getByRole('textbox').nth(2).fill('1 Đường Kiểm Thử, Quận 1')
    await page.getByRole('button', { name: 'Đặt hàng' }).click()

    await expect(page).toHaveURL(/\/orders\/[0-9a-f-]+$/)
    orderID = page.url().split('/').pop()
    await expect(page.getByRole('heading', { name: /Đơn / })).toBeVisible()
    await expect(page.getByText('Hạn thanh toán:')).toBeVisible()

    await page.getByRole('button', { name: 'Hủy đơn' }).click()
    await page.getByRole('button', { name: 'Xác nhận hủy đơn' }).click()
    await expect(page.getByText('Đã hủy')).toBeVisible()
    orderID = undefined
  } finally {
    if (orderID) {
      await page.evaluate(async (id) => {
        const csrf = document.cookie.split(';').map((part) => part.trim()).find((part) => part.startsWith('salescale_csrf='))?.split('=').slice(1).join('=') ?? ''
        await fetch(`/api/orders/${id}/cancel`, { method: 'POST', credentials: 'same-origin', headers: { 'X-CSRF-Token': decodeURIComponent(csrf) } })
      }, orderID)
    }
  }
})
