export interface Diagnosis {
  id: string
  tenant_id: string
  patient_id: string
  descripcion: string
  codigo_cie10?: string
  tipo?: string
  estado_problema?: string
  fecha_inicio?: string
  notas?: string
  state: string
}
