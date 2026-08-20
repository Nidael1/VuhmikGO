import { http } from '@/infrastructure/api/httpClient'
import type { Appointment, AppointmentRequest } from '@/domain/types/appointment'

export const appointmentRepository = {
  async listAll(): Promise<Appointment[]> {
    const res = await http.get('/appointments') as any
    return res.data?.items ?? []
  },

  async listToday(date: string): Promise<Appointment[]> {
    const res = await http.get(`/appointments/today?date=${date}`) as any
    return res.data?.items ?? []
  },

  async listByPatient(patientId: string): Promise<Appointment[]> {
    const res = await http.get(`/patients/${patientId}/appointments`) as any
    return res.data?.items ?? []
  },

  async create(data: AppointmentRequest): Promise<{ id: string; state: string }> {
    const res = await http.post('/appointments', data) as any
    return res.data
  },

  async cancel(id: string): Promise<void> {
    await http.patch(`/appointments/${id}/state`, { state: 'cancelled' })
  },

  async noShow(id: string): Promise<void> {
    await http.patch(`/appointments/${id}/state`, { state: 'no_show' })
  },

  async complete(id: string): Promise<void> {
    await http.patch(`/appointments/${id}/state`, { state: 'completed' })
  },
}
