export interface Immunization {
  id: string
  tenant_id: string
  patient_id: string
  vacuna: string
  fecha_aplicacion: string
  lote?: string
  dosis?: string
  via?: string
  aplicada_por?: string
  notas?: string
  state: string
}
