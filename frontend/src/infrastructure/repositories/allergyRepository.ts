import { http } from '@/infrastructure/api/httpClient'
import type { Allergy } from '@/domain/types/allergy'

export const allergyRepository = {
  async list(patientId: string): Promise<Allergy[]> {
    const res = await http.get(`/api/v1/patients/${patientId}/allergies`) as any
    return res.data?.items ?? []
  },

  async create(patientId: string, data: {
    agente: string
    tipo_reaccion: string
    criticidad?: string
    certeza?: string
    notas?: string
  }): Promise<Allergy> {
    const res = await http.post(`/api/v1/patients/${patientId}/allergies`, data) as any
    return res.data
  },

  async void(allergyId: string): Promise<void> {
    await http.post(`/api/v1/allergies/${allergyId}/void`, {})
  }
}
