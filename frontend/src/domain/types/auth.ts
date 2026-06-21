// Tipos de dominio para autenticación.

export interface AuthTokens {
  token: string
  tenant_id: string
  actor_id: string
}

export interface UserProfile {
  actor_id: string
  tenant_id: string
}
