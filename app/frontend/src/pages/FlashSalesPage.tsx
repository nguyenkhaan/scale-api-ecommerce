import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { isApiError } from '../lib/api-error'
import { cartClient } from '../lib/cart-api'
import { flashSaleClient, type FlashSaleCampaign } from '../lib/flash-sale-api'

function formatVND(amount: number): string {
  return new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount)
}

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

export function FlashSalesPage() {
  const [campaigns, setCampaigns] = useState<FlashSaleCampaign[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionMessage, setActionMessage] = useState<string | null>(null)
  const [addingSkuId, setAddingSkuId] = useState<string | null>(null)

  const loadSales = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await flashSaleClient.listPublicFlashSales()
      setCampaigns(data.items)
    } catch (err) {
      if (isApiError(err)) {
        setError(`Không thể tải chương trình flash sale (${err.code}).`)
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
      .listPublicFlashSales()
      .then((data) => {
        if (active) setCampaigns(data.items)
      })
      .catch((err) => {
        if (!active) return
        if (isApiError(err)) {
          setError(`Không thể tải chương trình flash sale (${err.code}).`)
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

  const handleAddToCart = async (skuId: string, flashSaleItemId: string) => {
    setAddingSkuId(skuId)
    setActionMessage(null)
    try {
      await cartClient.upsertCartItem(skuId, {
        quantity: 1,
        price_source: 'flash_sale',
        flash_sale_item_id: flashSaleItemId,
      })
      setActionMessage(`Đã thêm sản phẩm vào giỏ hàng!`)
    } catch (err) {
      if (isApiError(err)) {
        if (err.status === 401) {
          setActionMessage('Vui lòng đăng nhập để thêm vào giỏ hàng.')
        } else {
          setActionMessage(`Thêm vào giỏ thất bại: ${err.code}`)
        }
      } else {
        setActionMessage('Lỗi kết nối khi thêm vào giỏ hàng.')
      }
    } finally {
      setAddingSkuId(null)
    }
  }

  return (
    <div className="mx-auto max-w-5xl">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">Flash Sale</h1>
          <p className="mt-1 text-sm text-slate-600">
            Săn deal chớp nhoáng với giá ưu đãi đặc biệt trong thời gian giới hạn.
          </p>
        </div>
        <Link
          to="/cart"
          className="inline-flex items-center gap-2 rounded bg-amber-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-amber-500 focus-visible:outline-2 focus-visible:outline-amber-600"
        >
          Xem giỏ hàng
        </Link>
      </div>

      {actionMessage && (
        <div
          role="status"
          className="mt-4 rounded border border-amber-300 bg-amber-50 p-4 text-sm font-medium text-amber-900"
        >
          {actionMessage}
        </div>
      )}

      {loading && (
        <div className="py-12 text-center text-slate-500" role="status">
          Đang tải danh sách sự kiện flash sale...
        </div>
      )}

      {error && (
        <div role="alert" className="mt-6 rounded border border-rose-200 bg-rose-50 p-4 text-rose-800">
          <p className="font-semibold">{error}</p>
          <button
            type="button"
            onClick={() => void loadSales()}
            className="mt-2 text-sm font-semibold underline hover:text-rose-950"
          >
            Thử lại
          </button>
        </div>
      )}

      {!loading && !error && campaigns.length === 0 && (
        <div className="mt-8 rounded border border-dashed border-slate-300 p-12 text-center">
          <p className="text-base font-medium text-slate-700">Hiện tại không có chương trình Flash Sale nào.</p>
          <p className="mt-1 text-sm text-slate-500">Vui lòng quay lại sau!</p>
        </div>
      )}

      <div className="mt-6 space-y-8">
        {campaigns.map((campaign) => {
          const isActive = campaign.status === 'active'
          const statusBadge = isActive ? (
            <span className="inline-flex items-center rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800 animate-pulse">
              ● ĐANG DIỄN RA
            </span>
          ) : (
            <span className="inline-flex items-center rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-800">
              SẮP DIỄN RA
            </span>
          )

          return (
            <section
              key={campaign.id}
              className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm"
              aria-labelledby={`campaign-title-${campaign.id}`}
            >
              <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-100 bg-slate-50 px-6 py-4">
                <div>
                  <div className="flex items-center gap-3">
                    <h2 id={`campaign-title-${campaign.id}`} className="text-xl font-bold text-slate-900">
                      {campaign.name}
                    </h2>
                    {statusBadge}
                  </div>
                  <p className="mt-1 text-xs text-slate-500">
                    Thời gian: {formatDateTime(campaign.starts_at)} — {formatDateTime(campaign.ends_at)}
                  </p>
                </div>
              </div>

              <div className="p-6">
                <h3 className="text-sm font-semibold uppercase tracking-wider text-slate-500">
                  Sản phẩm ưu đãi ({campaign.items.length})
                </h3>
                <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                  {campaign.items.map((item) => (
                    <div
                      key={item.id}
                      className="flex flex-col justify-between rounded border border-slate-200 p-4 transition-shadow hover:shadow-md"
                    >
                      <div>
                        <div className="text-xs font-mono text-slate-400">SKU: {item.sku_id.slice(0, 8)}...</div>
                        <div className="mt-2 text-2xl font-bold text-rose-600">
                          {formatVND(item.sale_price)}
                        </div>
                        <div className="mt-1 text-xs text-slate-600">
                          Giới hạn mua: <span className="font-semibold">{item.purchase_limit} cái/khách</span>
                        </div>
                      </div>

                      <div className="mt-4 pt-3 border-t border-slate-100">
                        <button
                          type="button"
                          disabled={!isActive || addingSkuId === item.sku_id}
                          onClick={() => void handleAddToCart(item.sku_id, item.id)}
                          className="w-full rounded bg-rose-600 px-3 py-2 text-xs font-bold text-white transition hover:bg-rose-500 disabled:cursor-not-allowed disabled:bg-slate-300"
                        >
                          {!isActive
                            ? 'Chưa mở bán'
                            : addingSkuId === item.sku_id
                              ? 'Đang thêm...'
                              : 'Thêm vào giỏ'}
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </section>
          )
        })}
      </div>
    </div>
  )
}
