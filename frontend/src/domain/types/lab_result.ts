export interface LabResult {
  id: string
  tenant_id: string
  patient_id: string
  estudio: string
  fecha_estudio: string
  resultado?: string
  laboratorio?: string
  unidades?: string
  valor_referencia?: string
  notas?: string
  state: string
}
