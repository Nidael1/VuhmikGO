import { http } from '@/infrastructure/api/httpClient'
import type { Prescription, PrescriptionRequest } from '@/domain/types/prescription'

export const prescriptionRepository = {
  async listByPatient(patientId: string): Promise<Prescription[]> {
    const res = await http.get(`/patients/${patientId}/prescriptions`) as any
    return res.data?.items ?? []
  },

  async listAll(): Promise<Prescription[]> {
    const res = await http.get('/prescriptions') as any
    return res.data?.items ?? []
  },

  async create(patientId: string, data: PrescriptionRequest): Promise<{ id: string; state: string }> {
    const res = await http.post(`/patients/${patientId}/prescriptions`, data) as any
    return res.data
  },

  async emit(prescriptionId: string): Promise<{ id: string; state: string; issued_at: string }> {
    const res = await http.post(`/prescriptions/${prescriptionId}/emit`, {}) as any
    return res.data
  },

  async void(prescriptionId: string): Promise<void> {
    await http.post(`/prescriptions/${prescriptionId}/void`, {})
  },
}
