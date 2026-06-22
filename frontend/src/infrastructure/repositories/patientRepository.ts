import { http } from '@/infrastructure/api/httpClient'
import type { ApiResponse } from '@/domain/types/evidence'
import type { Patient, PatientRequest } from '@/domain/types/patient'

export const patientRepository = {
  async list(): Promise<Patient[]> {
    const res = await http.get<ApiResponse<{ items: Patient[] }>>('/patients')
    if (res.error) throw new Error(res.error.message)
    return res.data!.items
  },

  async get(id: string): Promise<Patient> {
    const res = await http.get<ApiResponse<Patient>>(`/patients/${id}`)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async create(payload: PatientRequest): Promise<Patient> {
    const res = await http.post<ApiResponse<Patient>>('/patients', payload)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async update(id: string, payload: PatientRequest): Promise<Patient> {
    const res = await http.post<ApiResponse<Patient>>(`/patients/${id}`, payload)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },
}
