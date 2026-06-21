// Cliente HTTP base para la API /api/v1.
// Inyecta el JWT token desde el store de auth en cada request.
// No contiene lógica de negocio.

const BASE_URL = '/api/v1'

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  // Importación dinámica para evitar dependencia circular con el store
  const { useAuthStore } = await import('@/app/stores/auth')
  const auth = useAuthStore()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }

  if (auth.token) {
    headers['Authorization'] = `Bearer ${auth.token}`
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
  })

  if (!res.ok && res.status === 401) {
    auth.clearSession()
    window.location.href = '/login'
    throw new Error('UNAUTHORIZED')
  }

  return res.json() as Promise<T>
}

export const http = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),
}
