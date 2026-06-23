// Cliente HTTP base para la API /api/v1.
// Maneja renovacion automatica de access token via refresh token.
// Access token: 15 minutos. Refresh token: 7 dias.

const BASE_URL = '/api/v1'

async function request<T>(
  path: string,
  options: RequestInit = {},
  retry = true,
): Promise<T> {
  const { useAuthStore } = await import('@/app/stores/auth')
  const auth = useAuthStore()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }

  if (auth.token) {
    headers['Authorization'] = `Bearer ${auth.token}`
  }

  const res = await fetch(`${BASE_URL}${path}`, { ...options, headers })

  // Si el access token expiró, intentar renovar con refresh token
  if (res.status === 401 && retry && auth.refreshToken) {
    const renewed = await tryRefresh(auth.refreshToken)
    if (renewed) {
      auth.setSession(renewed)
      return request<T>(path, options, false) // reintento sin loop
    }
    auth.clearSession()
    window.location.href = '/login'
    throw new Error('SESSION_EXPIRED')
  }

  if (!res.ok && res.status === 401) {
    auth.clearSession()
    window.location.href = '/login'
    throw new Error('UNAUTHORIZED')
  }

  return res.json() as Promise<T>
}

async function tryRefresh(refreshToken: string) {
  try {
    const res = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    if (!res.ok) return null
    const data = await res.json()
    return data.data ?? null
  } catch {
    return null
  }
}

export const http = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),
}
