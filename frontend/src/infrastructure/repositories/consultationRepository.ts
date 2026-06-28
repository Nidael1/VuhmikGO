import { http } from '@/infrastructure/api/httpClient'
import type { Consultation, ConsultationRequest } from '@/domain/types/consultation'

export const consultationRepository = {
  async listByPatient(patientId: string): Promise<Consultation[]> {
    const res = await http.get(`/patients/${patientId}/consultations`) as any
    return res.data?.items ?? []
  },

  async listAll(): Promise<Consultation[]> {
    const res = await http.get('/consultations') as any
    return res.data?.items ?? []
  },

  async create(patientId: string, data: ConsultationRequest): Promise<{ id: string; state: string }> {
    const res = await http.post(`/patients/${patientId}/consultations`, data) as any
    return res.data
  },

  async get(id: string): Promise<Consultation> {
    const res = await http.get(`/consultations/${id}`) as any
    return res.data
  },
}
