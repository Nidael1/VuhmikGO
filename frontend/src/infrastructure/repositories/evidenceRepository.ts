import { http } from '@/infrastructure/api/httpClient'
import type { ApiResponse } from '@/domain/types/evidence'
import type { Evidence } from '@/domain/types/evidence'

export interface DraftPayload {
  subject_ref: string
  content: string
}

export interface VoidPayload {
  reason_code: string
}

export interface ReplacePayload {
  reason_code: string
  replacement_id: string
}

export const evidenceRepository = {
  async list(): Promise<Evidence[]> {
    const res = await http.get<ApiResponse<{ items: Evidence[] }>>('/evidence')
    if (res.error) throw new Error(res.error.message)
    return res.data!.items
  },

  async get(id: string): Promise<Evidence> {
    const res = await http.get<ApiResponse<Evidence>>(`/evidence/${id}`)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async draft(payload: DraftPayload): Promise<Evidence> {
    const res = await http.post<ApiResponse<Evidence>>('/evidence/draft', payload)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async emit(id: string): Promise<Evidence> {
    const res = await http.post<ApiResponse<Evidence>>(`/evidence/${id}/emit`)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async void(id: string, payload: VoidPayload): Promise<Evidence> {
    const res = await http.post<ApiResponse<Evidence>>(`/evidence/${id}/void`, payload)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async replace(id: string, payload: ReplacePayload): Promise<{ voided: Evidence; replacement: Evidence }> {
    const res = await http.post<ApiResponse<{ voided: Evidence; replacement: Evidence }>>(
      `/evidence/${id}/replace`,
      payload,
    )
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async export(id: string): Promise<Blob> {
    const { useAuthStore } = await import('@/app/stores/auth')
    const auth = useAuthStore()
    const res = await fetch(`/api/v1/evidence/${id}/export`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${auth.token}` },
    })
    if (!res.ok) throw new Error('export fallido')
    return res.blob()
  },

  async exportWithFormat(id: string, accept: string): Promise<Blob> {
    const { useAuthStore } = await import('@/app/stores/auth')
    const auth = useAuthStore()
    const res = await fetch(`/api/v1/evidence/${id}/export`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${auth.token}`,
        Accept: accept,
      },
    })
    if (!res.ok) throw new Error('export fallido')
    return res.blob()
  },
}
