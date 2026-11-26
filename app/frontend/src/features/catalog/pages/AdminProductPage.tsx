import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'

import { Button, FeedbackState } from '@/components/common'
import { catalogQueryKeys, useAdminProductQuery, useCategoriesQuery } from '@/features/catalog/catalog.hook'
import { createProduct, createSKU, isCatalogApiError, updateProduct, updateSKU } from '@/features/catalog/catalog.service'
import type { CatalogProduct, CreateSkuInput } from '@/features/catalog/catalog.type'
import { formatVND } from '@/lib/format'

type SkuDraft = { code: string; attributes: string; regularPrice: string; maxQuantity: string; initialOnHand: string }

const blankSku: SkuDraft = { code: '', attributes: '{}', regularPrice: '', maxQuantity: '1', initialOnHand: '0' }

export function AdminProductPage() {
  const { productId = '' } = useParams()
  const isNew = productId === 'new'
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const productQuery = useAdminProductQuery(isNew ? '' : productId)
  const categoriesQuery = useCategoriesQuery()
  const categories = categoriesQuery.data?.items ?? []
  const product = productQuery.data
  const [name, setName] = useState<string | null>(null)
  const [description, setDescription] = useState<string | null>(null)
  const [categoryID, setCategoryID] = useState<string | null>(null)
  const [imageURL, setImageURL] = useState<string | null>(null)
  const [active, setActive] = useState<boolean | null>(null)
  const [skuDraft, setSkuDraft] = useState<SkuDraft>(blankSku)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const formName = name ?? product?.name ?? ''
  const formDescription = description ?? product?.description ?? ''
  const formCategoryID = categoryID ?? product?.category_id ?? ''
  const formImageURL = imageURL ?? product?.image_url ?? ''
  const formActive = active ?? product?.active ?? true

  async function saveProduct(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setNotice(null)
    setSaving(true)
    try {
      if (isNew) {
        const sku = parseSkuDraft(skuDraft)
        const created = await createProduct({
          name: formName.trim(),
          description: formDescription,
          ...(formCategoryID ? { category_id: formCategoryID } : {}),
          ...(formImageURL.trim() ? { image_url: formImageURL.trim() } : {}),
          skus: [sku],
        })
        await queryClient.invalidateQueries({ queryKey: catalogQueryKeys.adminProducts(20, 'active') })
        navigate(`/admin/products/${created.id}`, { replace: true })
        return
      }
      const updated = await updateProduct(productId, {
        name: formName.trim(),
        description: formDescription,
        category_id: formCategoryID || null,
        image_url: formImageURL.trim() || null,
        active: formActive,
      })
      queryClient.setQueryData(catalogQueryKeys.adminProduct(productId), updated)
      await queryClient.invalidateQueries({ queryKey: ['catalog', 'admin-products'] })
      setNotice('Đã lưu thông tin sản phẩm.')
    } catch (cause) {
      setError(describeMutationError(cause))
    } finally {
      setSaving(false)
    }
  }

  async function addSKU(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!product) return
    setError(null)
    setSaving(true)
    try {
      await createSKU(product.id, parseSkuDraft(skuDraft))
      setSkuDraft(blankSku)
      await queryClient.invalidateQueries({ queryKey: catalogQueryKeys.adminProduct(product.id) })
      setNotice('Đã thêm SKU và ghi nhận tồn ban đầu.')
    } catch (cause) {
      setError(describeMutationError(cause))
    } finally {
      setSaving(false)
    }
  }

  async function saveSKU(skuID: string, body: { regular_price: number; active: boolean }) {
    setError(null)
    setSaving(true)
    try {
      await updateSKU(skuID, body)
      if (product) await queryClient.invalidateQueries({ queryKey: catalogQueryKeys.adminProduct(product.id) })
      setNotice('Đã cập nhật SKU.')
    } catch (cause) {
      setError(describeMutationError(cause))
    } finally {
      setSaving(false)
    }
  }

  if (!isNew && productQuery.isPending) return <FeedbackState kind="loading" title="Đang tải sản phẩm" message="Vui lòng chờ trong giây lát." />
  if (!isNew && productQuery.isError) return <FeedbackState kind={isCatalogApiError(productQuery.error) && productQuery.error.status === 404 ? 'notFound' : 'error'} title="Không thể tải sản phẩm" message="Sản phẩm không tồn tại hoặc phiên quản trị đã hết hạn." action={{ label: 'Tải lại', onClick: () => void productQuery.refetch() }} />
  if (categoriesQuery.isError) return <FeedbackState kind="error" title="Không thể tải danh mục" message="Không thể mở biểu mẫu vì thiếu danh sách danh mục." action={{ label: 'Tải lại', onClick: () => void categoriesQuery.refetch() }} />

  return (
    <section className="mx-auto max-w-4xl">
      <Link to="/admin/products" className="inline-flex min-h-11 items-center text-sm font-semibold text-[var(--ss-color-primary-strong)] hover:underline">← Danh sách sản phẩm</Link>
      <div className="mt-2 border-b border-[var(--ss-color-border)] pb-4">
        <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">{isNew ? 'Tạo sản phẩm' : 'Chỉnh sửa sản phẩm'}</h1>
        <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Giá và tồn đầu kỳ được lưu bởi API; đơn hàng cũ giữ nguyên snapshot.</p>
      </div>

      {error && <p role="alert" className="mt-4 border border-red-200 bg-red-50 p-3 text-sm text-[var(--ss-color-danger)]">{error}</p>}
      {notice && <p role="status" className="mt-4 border border-green-200 bg-green-50 p-3 text-sm">{notice}</p>}

      <form onSubmit={(event) => void saveProduct(event)} className="mt-5 space-y-4">
        <label className="block text-sm font-medium">Tên sản phẩm<input required maxLength={200} value={formName} onChange={(event) => setName(event.target.value)} className={inputClass} /></label>
        <label className="block text-sm font-medium">Mô tả<textarea maxLength={5000} rows={4} value={formDescription} onChange={(event) => setDescription(event.target.value)} className={`${inputClass} py-2`} /></label>
        <label className="block text-sm font-medium">Danh mục<select aria-label="Danh mục" value={formCategoryID} onChange={(event) => setCategoryID(event.target.value)} className={inputClass}><option value="">Không phân loại</option>{categories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label>
        <label className="block text-sm font-medium">Đường dẫn ảnh<input type="url" value={formImageURL} onChange={(event) => setImageURL(event.target.value)} className={inputClass} /></label>
        {!isNew && <label className="flex min-h-11 items-center gap-2 text-sm font-medium"><input type="checkbox" checked={formActive} onChange={(event) => setActive(event.target.checked)} /> Đang bán công khai</label>}
        {isNew && <SkuFields draft={skuDraft} onChange={setSkuDraft} />}
        <Button type="submit" loading={saving}>{isNew ? 'Tạo sản phẩm và SKU' : 'Lưu sản phẩm'}</Button>
      </form>

      {!isNew && product && (
        <div className="mt-8 border-t border-[var(--ss-color-border)] pt-5">
          <h2 className="text-lg font-semibold">SKU của sản phẩm</h2>
          <ul className="mt-3 divide-y divide-[var(--ss-color-border)]">
            {product.skus.map((sku) => <SkuEditor key={sku.id} sku={sku} saving={saving} onSave={(body) => void saveSKU(sku.id, body)} />)}
          </ul>
          <form onSubmit={(event) => void addSKU(event)} className="mt-5 space-y-4 border-t border-[var(--ss-color-border)] pt-5">
            <h3 className="font-semibold">Thêm SKU</h3>
            <SkuFields draft={skuDraft} onChange={setSkuDraft} />
            <Button type="submit" loading={saving}>Thêm SKU và tồn ban đầu</Button>
          </form>
        </div>
      )}
    </section>
  )
}

function SkuFields({ draft, onChange }: { draft: SkuDraft; onChange: (draft: SkuDraft) => void }) {
  const set = (key: keyof SkuDraft, value: string) => onChange({ ...draft, [key]: value })
  return (
    <fieldset className="grid gap-3 border border-[var(--ss-color-border)] p-4 sm:grid-cols-2">
      <legend className="px-1 text-sm font-semibold">Phiên bản đầu tiên</legend>
      <label className="block text-sm font-medium">Mã SKU<input required maxLength={100} value={draft.code} onChange={(event) => set('code', event.target.value)} className={inputClass} /></label>
      <label className="block text-sm font-medium">Thuộc tính dạng JSON<input value={draft.attributes} onChange={(event) => set('attributes', event.target.value)} className={inputClass} /><span className="mt-1 block text-xs font-normal text-[var(--ss-color-muted)]">Ví dụ: {'{"size":"M"}'}</span></label>
      <label className="block text-sm font-medium">Giá thường (VND)<input required type="number" min={1} step={1} value={draft.regularPrice} onChange={(event) => set('regularPrice', event.target.value)} className={inputClass} /></label>
      <label className="block text-sm font-medium">Giới hạn mỗi đơn<input required type="number" min={1} step={1} value={draft.maxQuantity} onChange={(event) => set('maxQuantity', event.target.value)} className={inputClass} /></label>
      <label className="block text-sm font-medium">Tồn ban đầu<input required type="number" min={0} step={1} value={draft.initialOnHand} onChange={(event) => set('initialOnHand', event.target.value)} className={inputClass} /></label>
    </fieldset>
  )
}

function SkuEditor({ sku, saving, onSave }: { sku: CatalogProduct['skus'][number]; saving: boolean; onSave: (body: { regular_price: number; active: boolean }) => void }) {
  const [price, setPrice] = useState(String(sku.regular_price))
  const [active, setActive] = useState(sku.active)
  return (
    <li className="flex flex-wrap items-end gap-3 py-4">
      <div className="mr-auto min-w-36"><p className="font-semibold">{sku.code}</p><p className="text-xs text-[var(--ss-color-muted)]">{Object.entries(sku.attributes).map(([key, value]) => `${key}: ${String(value)}`).join(' · ') || 'Không có thuộc tính'}</p><p className="mt-1 text-xs text-[var(--ss-color-muted)]">Giá hiện tại: {formatVND(sku.regular_price)}</p></div>
      <label className="text-sm">Giá mới<input aria-label={`Giá SKU ${sku.code}`} type="number" min={1} step={1} value={price} onChange={(event) => setPrice(event.target.value)} className="mt-1 min-h-10 w-36 rounded border border-[var(--ss-color-input-border)] px-2" /></label>
      <label className="flex min-h-10 items-center gap-2 text-sm"><input type="checkbox" checked={active} onChange={(event) => setActive(event.target.checked)} /> Đang bán</label>
      <Button variant="secondary" loading={saving} onClick={() => onSave({ regular_price: Number(price), active })}>Lưu SKU</Button>
    </li>
  )
}

function parseSkuDraft(draft: SkuDraft): CreateSkuInput {
  const attributes = JSON.parse(draft.attributes || '{}') as unknown
  if (attributes === null || Array.isArray(attributes) || typeof attributes !== 'object') throw new Error('Thuộc tính SKU phải là một JSON object.')
  return {
    code: draft.code.trim(),
    attributes: attributes as Record<string, unknown>,
    regular_price: Number(draft.regularPrice),
    max_quantity_per_order: Number(draft.maxQuantity),
    initial_on_hand: Number(draft.initialOnHand),
  }
}

function describeMutationError(error: unknown): string {
  if (error instanceof Error && error.message.startsWith('Thuộc tính SKU')) return error.message
  if (isCatalogApiError(error)) {
    if (error.status === 401 || error.status === 403) return 'Phiên quản trị hết hạn hoặc tài khoản không có quyền. Hãy đăng nhập lại.'
    if (error.code === 'SKU_CODE_ALREADY_EXISTS') return 'Mã SKU đã tồn tại; hãy chọn mã khác.'
    if (error.violations.length) return error.violations.map((item) => `${item.field}: ${item.reason}`).join(' · ')
    return `Thao tác chưa thành công (${error.code}). Vui lòng thử lại.`
  }
  return 'Không thể lưu dữ liệu. Kiểm tra các trường đã nhập rồi thử lại.'
}

const inputClass = 'mt-1 min-h-11 w-full rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] bg-white px-3 text-sm'
