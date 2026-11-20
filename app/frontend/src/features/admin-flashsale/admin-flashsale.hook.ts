import { useState } from 'react'
import { flashSaleClient, type CreateFlashSaleRequest, type FlashSaleCampaign } from './admin-flashsale.service'
import { usePagedList } from '@/lib/use-paged-list'
import { describeAdminFailure, type AdminAction } from './admin-flashsale.logic'

const listAdmin = (cursor?: string) => flashSaleClient.listAdminFlashSales({ cursor })

export type AdminNotice = { kind: 'success' | 'error'; text: string; recovery?: 'reload' | 'login' }

const successText: Record<Exclude<AdminAction, 'edit'>, string> = {
  schedule: 'Đã lên lịch chiến dịch.',
  cancel_schedule: 'Đã hủy lịch; chiến dịch quay về bản nháp.',
  stop: 'Đã dừng chiến dịch.',
}

export function useAdminFlashSales() {
  const list = usePagedList(listAdmin, 'danh sách chiến dịch')
  const [busyId, setBusyId] = useState<string | null>(null)
  const [notice, setNotice] = useState<AdminNotice | null>(null)

  const runAction = async (campaign: FlashSaleCampaign, action: Exclude<AdminAction, 'edit'>) => {
    setBusyId(campaign.id)
    setNotice(null)
    try {
      if (action === 'schedule') await flashSaleClient.scheduleFlashSale(campaign.id)
      else if (action === 'cancel_schedule') await flashSaleClient.cancelScheduleFlashSale(campaign.id)
      else await flashSaleClient.stopFlashSale(campaign.id)
      setNotice({ kind: 'success', text: successText[action] })
      await list.reload()
    } catch (err) {
      const failure = describeAdminFailure(err)
      setNotice({ kind: 'error', text: failure.text, recovery: failure.requiresLogin ? 'login' : 'reload' })
      if (failure.stateConflict) await list.reload()
    } finally {
      setBusyId(null)
    }
  }

  /** Throws on failure so the form can show field violations; reloads on a state conflict. */
  const save = async (request: CreateFlashSaleRequest, editing?: FlashSaleCampaign) => {
    try {
      if (editing) await flashSaleClient.updateFlashSale(editing.id, request)
      else await flashSaleClient.createFlashSale(request)
    } catch (err) {
      if (describeAdminFailure(err).stateConflict) await list.reload()
      throw err
    }
    setNotice({ kind: 'success', text: editing ? 'Đã lưu thay đổi bản nháp.' : 'Đã tạo chiến dịch bản nháp.' })
    await list.reload()
  }

  return { list, busyId, notice, runAction, save }
}
