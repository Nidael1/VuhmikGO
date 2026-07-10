import { http } from '@/infrastructure/api/httpClient'
import type { LabResult } from '@/domain/types/lab_result'

export const labResultRepository = {
  async list(patientId: string): Promise<LabResult[]> {
    const res = await http.get(`/patients/${patientId}/lab-results`) as any
    return res.data?.items ?? []
  },
  async create(patientId: string, data: {
    estudio: string
    fecha_estudio: string
    resultado?: string
    laboratorio?: string
    unidades?: string
    valor_referencia?: string
    notas?: string
  }): Promise<LabResult> {
    const res = await http.post(`/patients/${patientId}/lab-results`, data) as any
    return res.data
  },
  async void(labResultId: string): Promise<void> {
    await http.post(`/lab-results/${labResultId}/void`, {})
  }
}
