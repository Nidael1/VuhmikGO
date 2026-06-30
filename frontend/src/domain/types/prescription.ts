export interface Prescription {
  id: string
  tenant_id: string
  patient_id: string
  medicamento_generico: string
  dosis: string
  diagnostico?: string
  indicaciones?: string
  seguimiento?: string
  state: string
  created_at: string
  issued_at?: string
  consultation_id?: string
}

export interface PrescriptionRequest {
  medicamento_generico: string
  dosis: string
  diagnostico?: string
  indicaciones?: string
  seguimiento?: string
  consultation_id?: string
}
