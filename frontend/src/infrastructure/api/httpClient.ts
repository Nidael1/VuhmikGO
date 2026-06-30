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

  // Rutas de auth no deben redirigir — el caller maneja el error
  const isAuthRoute = path.startsWith('/auth/')

  // Si el access token expiró en ruta protegida, intentar renovar
  if (res.status === 401 && retry && auth.refreshToken && !isAuthRoute) {
    const renewed = await tryRefresh(auth.refreshToken)
    if (renewed) {
      auth.setSession(renewed)
      return request<T>(path, options, false) // reintento sin loop
    }
    auth.clearSession()
    window.location.href = '/login'
    throw new Error('SESSION_EXPIRED')
  }

  if (!res.ok && res.status === 401 && !isAuthRoute) {
    auth.clearSession()
    window.location.href = '/login'
    throw new Error('UNAUTHORIZED')
  }

  const json = await res.json()
  if (!res.ok) {
    throw new Error(json?.error?.message || json?.error?.code || `HTTP ${res.status}`)
  }
  return json as T
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
