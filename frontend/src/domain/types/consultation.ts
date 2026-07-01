export interface Consultation {
  id: string
  tenant_id: string
  patient_id: string
  ta?: string
  fc?: string
  fr?: string
  temp?: string
  peso?: string
  talla?: string
  sao2?: string
  state: string
  created_at: string
  issued_at?: string
  tiene_receta?: boolean
}

export interface ConsultationRequest {
  ta?: string
  fc?: string
  fr?: string
  temp?: string
  peso?: string
  talla?: string
  sao2?: string
  tiene_receta?: boolean
}
