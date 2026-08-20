export interface Appointment {
  id: string
  patient_id: string
  patient_nombre?: string
  scheduled_at: string
  duration_min: number
  reason: string
  state: 'scheduled' | 'completed' | 'cancelled' | 'no_show'
  consultation_id?: string
  notes?: string
  created_at: string
}

export interface AppointmentRequest {
  patient_id: string
  scheduled_at: string
  duration_min?: number
  reason?: string
  notes?: string
}
