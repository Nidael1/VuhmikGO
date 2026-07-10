import { http } from '@/infrastructure/api/httpClient'
import type { Diagnosis } from '@/domain/types/diagnosis'

export const diagnosisRepository = {
  async list(patientId: string): Promise<Diagnosis[]> {
    const res = await http.get(`/patients/${patientId}/diagnoses`) as any
    return res.data?.items ?? []
  },
  async create(patientId: string, data: {
    descripcion: string
    codigo_cie10?: string
    tipo?: string
    estado_problema?: string
    fecha_inicio?: string
    notas?: string
  }): Promise<Diagnosis> {
    const res = await http.post(`/patients/${patientId}/diagnoses`, data) as any
    return res.data
  },
  async void(diagnosisId: string): Promise<void> {
    await http.post(`/diagnoses/${diagnosisId}/void`, {})
  }
}
