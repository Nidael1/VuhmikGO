// Tipos de dominio para pacientes (CRM).
// NOM-004-SSA3-2012, numeral 5.9.

export interface Patient {
  id: string
  tenant_id: string
  nombre: string
  fecha_nacimiento: string
  sexo: 'M' | 'F' | 'I'
  num_expediente: string
  created_at: string
  updated_at: string
}

export interface PatientRequest {
  nombre: string
  fecha_nacimiento: string
  sexo: string
}
