import { http } from '@/infrastructure/api/httpClient'
import type { Immunization } from '@/domain/types/immunization'

export const immunizationRepository = {
  async list(patientId: string): Promise<Immunization[]> {
    const res = await http.get(`/patients/${patientId}/immunizations`) as any
    return res.data?.items ?? []
  },
  async create(patientId: string, data: {
    vacuna: string
    fecha_aplicacion: string
    lote?: string
    dosis?: string
    via?: string
    aplicada_por?: string
    notas?: string
  }): Promise<Immunization> {
    const res = await http.post(`/patients/${patientId}/immunizations`, data) as any
    return res.data
  },
  async void(immunizationId: string): Promise<void> {
    await http.post(`/immunizations/${immunizationId}/void`, {})
  }
}
