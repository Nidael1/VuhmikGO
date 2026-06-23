export interface AuthTokens {
  token: string
  refresh_token: string
  tenant_id: string
  actor_id: string
}

export interface UserProfile {
  actor_id: string
  tenant_id: string
}
