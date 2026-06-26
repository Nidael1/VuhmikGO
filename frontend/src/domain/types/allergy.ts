export interface Allergy {
  id: string
  tenant_id: string
  patient_id: string
  agente: string
  tipo_reaccion: string
  criticidad?: string
  certeza?: string
  fecha_inicio?: string
  notas?: string
  state: string
}
