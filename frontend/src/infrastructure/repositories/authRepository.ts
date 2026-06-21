import { http } from '@/infrastructure/api/httpClient'
import type { ApiResponse, ApiError } from '@/domain/types/evidence'
import type { AuthTokens, UserProfile } from '@/domain/types/auth'

export interface LoginPayload {
  email: string
  password: string
}

export const authRepository = {
  async login(payload: LoginPayload): Promise<AuthTokens> {
    const res = await http.post<ApiResponse<AuthTokens>>('/auth/login', payload)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async register(payload: LoginPayload): Promise<AuthTokens> {
    const res = await http.post<ApiResponse<AuthTokens>>('/auth/register', payload)
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },

  async me(): Promise<UserProfile> {
    const res = await http.get<ApiResponse<UserProfile>>('/auth/me')
    if (res.error) throw new Error(res.error.message)
    return res.data!
  },
}
