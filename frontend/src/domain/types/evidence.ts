// Tipos de dominio para evidencia clínica.
// Espeja el contrato de la API /api/v1 — no contiene lógica de negocio.

export type EvidenceState = 'draft' | 'issued' | 'locked' | 'voided'

export interface Evidence {
  id: string
  tenant_id: string
  state: EvidenceState
  created_at: string
  issued_at: string | null
  voided_at: string | null
  replaced_by_id: string | null
}

export interface ApiResponse<T> {
  data: T | null
  error: ApiError | null
}

export interface ApiError {
  code: string
  message: string
}
