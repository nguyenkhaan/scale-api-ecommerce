import React, { useEffect, useState } from 'react'
import { isApiError } from '../lib/api-error'
import {
  flashSaleClient,
  type CreateFlashSaleItem,
  type FlashSaleCampaign,
} from '../lib/flash-sale-api'

function formatDateTime(isoString: string): string {
  try {
    const d = new Date(isoString)
    return d.toLocaleString('vi-VN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return isoString
  }
}

export function AdminFlashSalePage() {
  const [campaigns, setCampaigns] = useState<FlashSaleCampaign[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionSuccess, setActionSuccess] = useState<string | null>(null)
  const [busyCampaignId, setBusyCampaignId] = useState<string | null>(null)

  // Create Campaign Modal/Form state
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [formName, setFormName] = useState('')
  const [formStartsAt, setFormStartsAt] = useState('')
  const [formEndsAt, setFormEndsAt] = useState('')
  const [formItems, setFormItems] = useState<CreateFlashSaleItem[]>([
    { sku_id: '', sale_price: 10000, quota_total: 10, purchase_limit: 2 },
  ])
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)

  const loadCampaigns = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await flashSaleClient.listAdminFlashSales()
      setCampaigns(data.items)
    } catch (err) {
      if (isApiError(err)) {
        setError(`Không thể tải danh sách flash sale (${err.code}).`)
      } else {
        setError('Lỗi kết nối khi tải danh sách flash sale.')
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let active = true
    flashSaleClient
      .listAdminFlashSales()
      .then((data) => {
        if (active) setCampaigns(data.items)
      })
      .catch((err) => {
        if (!active) return
        if (isApiError(err)) {
          setError(`Không thể tải danh sách flash sale (${err.code}).`)
        } else {
          setError('Lỗi kết nối khi tải danh sách flash sale.')
        }
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [])

  const handleSchedule = async (id: string) => {
    setBusyCampaignId(id)
    setError(null)
    setActionSuccess(null)
    try {
      await flashSaleClient.scheduleFlashSale(id)
      setActionSuccess('Đã lên lịch chiến dịch thành công.')
      await loadCampaigns()
    } catch (err) {
      if (isApiError(err)) {
        if (err.status === 409) {
          setError('Không thể lên lịch: SKU đã được đăng ký trong chiến dịch khác cùng khung giờ (chống chồng SKU).')
        } else {
          setError(`Lên lịch thất bại (${err.code}).`)
        }
      } else {
        setError('Lỗi kết nối khi lên lịch.')
      }
    } finally {
      setBusyCampaignId(null)
    }
  }

  const handleCancelSchedule = async (id: string) => {
    setBusyCampaignId(id)
    setError(null)
    setActionSuccess(null)
    try {
      await flashSaleClient.cancelScheduleFlashSale(id)
      setActionSuccess('Đã hủy lịch thành công, chiến dịch quay về trạng thái bản nháp.')
      await loadCampaigns()
    } catch (err) {
      if (isApiError(err)) {
        setError(`Hủy lịch thất bại (${err.code}).`)
      } else {
        setError('Lỗi kết nối khi hủy lịch.')
      }
    } finally {
      setBusyCampaignId(null)
    }
  }

  const handleStop = async (id: string) => {
    setBusyCampaignId(id)
    setError(null)
    setActionSuccess(null)
    try {
      await flashSaleClient.stopFlashSale(id)
      setActionSuccess('Đã dừng chiến dịch thành công.')
      await loadCampaigns()
    } catch (err) {
      if (isApiError(err)) {
        setError(`Dừng chiến dịch thất bại (${err.code}).`)
      } else {
        setError('Lỗi kết nối khi dừng chiến dịch.')
      }
    } finally {
      setBusyCampaignId(null)
    }
  }

  const handleAddItemRow = () => {
    setFormItems((prev) => [
      ...prev,
      { sku_id: '', sale_price: 10000, quota_total: 10, purchase_limit: 2 },
    ])
  }

  const handleRemoveItemRow = (index: number) => {
    setFormItems((prev) => prev.filter((_, i) => i !== index))
  }

  const handleItemChange = (index: number, field: keyof CreateFlashSaleItem, value: any) => {
    setFormItems((prev) => {
      const next = [...prev]
      next[index] = { ...next[index], [field]: value }
      return next
    })
  }

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setCreateError(null)
    setCreating(true)

    try {
      const startsISO = new Date(formStartsAt).toISOString()
      const endsISO = new Date(formEndsAt).toISOString()

      await flashSaleClient.createFlashSale({
        name: formName,
        starts_at: startsISO,
        ends_at: endsISO,
        items: formItems.map((it) => ({
          sku_id: it.sku_id.trim(),
          sale_price: Number(it.sale_price),
          quota_total: Number(it.quota_total),
          purchase_limit: Number(it.purchase_limit),
        })),
      })

      setShowCreateModal(false)
      setActionSuccess('Đã tạo chiến dịch bản nháp thành công.')
      setFormName('')
      setFormStartsAt('')
      setFormEndsAt('')
      setFormItems([{ sku_id: '', sale_price: 10000, quota_total: 10, purchase_limit: 2 }])
      await loadCampaigns()
    } catch (err) {
      if (isApiError(err)) {
        if (err.violations.length > 0) {
          setCreateError(err.violations.map((v) => `${v.field}: ${v.reason}`).join(', '))
        } else {
          setCreateError(`Tạo thất bại (${err.code}).`)
        }
      } else {
        setCreateError('Lỗi kết nối khi tạo chiến dịch.')
      }
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">Quản lý Flash Sale</h1>
          <p className="mt-1 text-sm text-slate-600">
            Tạo, lên lịch và kiểm soát các chiến dịch flash sale và phân bổ quota SKU.
          </p>
        </div>
        <button
          type="button"
          onClick={() => setShowCreateModal(true)}
          className="rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-slate-800"
        >
          + Tạo chiến dịch mới
        </button>
      </div>

      {actionSuccess && (
        <div role="status" className="mt-4 rounded border border-emerald-300 bg-emerald-50 p-4 text-sm font-medium text-emerald-900">
          {actionSuccess}
        </div>
      )}

      {error && (
        <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-4 text-rose-800">
          <p className="font-semibold">{error}</p>
        </div>
      )}

      {loading && (
        <div className="py-12 text-center text-slate-500" role="status">
          Đang tải danh sách chiến dịch...
        </div>
      )}

      {!loading && campaigns.length === 0 && (
        <div className="mt-8 rounded border border-dashed border-slate-300 p-12 text-center">
          <p className="text-base font-medium text-slate-700">Chưa có chiến dịch Flash Sale nào.</p>
        </div>
      )}

      {!loading && campaigns.length > 0 && (
        <div className="mt-6 space-y-4">
          {campaigns.map((camp) => {
            const isBusy = busyCampaignId === camp.id
            return (
              <div
                key={camp.id}
                className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4"
              >
                <div>
                  <div className="flex items-center gap-3">
                    <h2 className="text-lg font-bold text-slate-900">{camp.name}</h2>
                    <span
                      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-bold ${
                        camp.status === 'active'
                          ? 'bg-emerald-100 text-emerald-800'
                          : camp.status === 'scheduled'
                            ? 'bg-blue-100 text-blue-800'
                            : camp.status === 'draft'
                              ? 'bg-slate-100 text-slate-800'
                              : camp.status === 'stopped'
                                ? 'bg-rose-100 text-rose-800'
                                : 'bg-stone-100 text-stone-600'
                      }`}
                    >
                      {camp.status.toUpperCase()}
                    </span>
                  </div>
                  <p className="mt-1 text-xs text-slate-500">
                    Thời gian: {formatDateTime(camp.starts_at)} — {formatDateTime(camp.ends_at)}
                  </p>
                  <p className="mt-1 text-xs text-slate-600">
                    Số SKU ưu đãi: {camp.items.length} mặt hàng
                  </p>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  {camp.status === 'draft' && (
                    <button
                      type="button"
                      disabled={isBusy}
                      onClick={() => void handleSchedule(camp.id)}
                      className="rounded bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-500 disabled:opacity-50"
                    >
                      Lên lịch
                    </button>
                  )}

                  {camp.status === 'scheduled' && (
                    <>
                      <button
                        type="button"
                        disabled={isBusy}
                        onClick={() => void handleCancelSchedule(camp.id)}
                        className="rounded border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 disabled:opacity-50"
                      >
                        Hủy lịch
                      </button>
                      <button
                        type="button"
                        disabled={isBusy}
                        onClick={() => void handleStop(camp.id)}
                        className="rounded bg-rose-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-rose-500 disabled:opacity-50"
                      >
                        Dừng
                      </button>
                    </>
                  )}

                  {camp.status === 'active' && (
                    <button
                      type="button"
                      disabled={isBusy}
                      onClick={() => void handleStop(camp.id)}
                      className="rounded bg-rose-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-rose-500 disabled:opacity-50"
                    >
                      Dừng khẩn cấp
                    </button>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-lg bg-white p-6 shadow-xl">
            <h2 className="text-xl font-bold text-slate-900">Tạo chiến dịch Flash Sale mới</h2>
            <p className="mt-1 text-xs text-slate-500">
              Chiến dịch mới sẽ ở trạng thái Bản nháp (draft). Bạn có thể kiểm tra trước khi Lên lịch.
            </p>

            {createError && (
              <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-3 text-xs text-rose-800">
                {createError}
              </div>
            )}

            <form onSubmit={handleCreateSubmit} className="mt-4 space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700">Tên chiến dịch</label>
                <input
                  type="text"
                  required
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder="Ví dụ: Giờ vàng 11.11"
                  className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm focus:border-amber-500 focus:outline-none"
                />
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div>
                  <label className="block text-xs font-semibold text-slate-700">Bắt đầu</label>
                  <input
                    type="datetime-local"
                    required
                    value={formStartsAt}
                    onChange={(e) => setFormStartsAt(e.target.value)}
                    className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm focus:border-amber-500 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-700">Kết thúc</label>
                  <input
                    type="datetime-local"
                    required
                    value={formEndsAt}
                    onChange={(e) => setFormEndsAt(e.target.value)}
                    className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm focus:border-amber-500 focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between">
                  <label className="block text-xs font-semibold text-slate-700">Danh sách SKU và giá ưu đãi</label>
                  <button
                    type="button"
                    onClick={handleAddItemRow}
                    className="text-xs font-semibold text-amber-700 hover:text-amber-800"
                  >
                    + Thêm dòng SKU
                  </button>
                </div>

                <div className="mt-2 space-y-3">
                  {formItems.map((item, idx) => (
                    <div key={idx} className="flex flex-wrap items-center gap-2 rounded border border-slate-200 p-2 bg-slate-50">
                      <input
                        type="text"
                        required
                        placeholder="SKU UUID"
                        value={item.sku_id}
                        onChange={(e) => handleItemChange(idx, 'sku_id', e.target.value)}
                        className="flex-1 min-w-[150px] rounded border border-slate-300 px-2 py-1 text-xs"
                      />
                      <input
                        type="number"
                        required
                        placeholder="Giá bán (₫)"
                        value={item.sale_price}
                        onChange={(e) => handleItemChange(idx, 'sale_price', Number(e.target.value))}
                        className="w-24 rounded border border-slate-300 px-2 py-1 text-xs"
                      />
                      <input
                        type="number"
                        required
                        placeholder="Quota"
                        value={item.quota_total}
                        onChange={(e) => handleItemChange(idx, 'quota_total', Number(e.target.value))}
                        className="w-20 rounded border border-slate-300 px-2 py-1 text-xs"
                      />
                      <input
                        type="number"
                        required
                        placeholder="Giới hạn mua"
                        value={item.purchase_limit}
                        onChange={(e) => handleItemChange(idx, 'purchase_limit', Number(e.target.value))}
                        className="w-20 rounded border border-slate-300 px-2 py-1 text-xs"
                      />
                      {formItems.length > 1 && (
                        <button
                          type="button"
                          onClick={() => handleRemoveItemRow(idx)}
                          className="text-xs font-bold text-rose-600 hover:text-rose-800"
                        >
                          ✕
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div className="mt-6 flex justify-end gap-3 pt-4 border-t border-slate-200">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="rounded border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50"
                >
                  Đóng
                </button>
                <button
                  type="submit"
                  disabled={creating}
                  className="rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800 disabled:opacity-50"
                >
                  {creating ? 'Đang tạo...' : 'Tạo bản nháp'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
