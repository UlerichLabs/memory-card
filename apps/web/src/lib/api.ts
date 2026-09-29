export class ApiError extends Error {
  readonly codigo: string
  readonly status?: number
  constructor(codigo: string, message: string, status?: number) {
    super(message); this.name = 'ApiError'; this.codigo = codigo; this.status = status
  }
}
const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')
export const REFRESH_TOKEN_KEY = 'refresh_token'
let currentAccessToken: string | null = null, refreshPromise: Promise<string> | null = null
let onAuthFailureCallback: (() => void) | null = null, onTokenRefreshedCallback: ((t: string) => void) | null = null
export const getAccessToken = () => currentAccessToken
export const setAccessToken = (t: string | null) => { currentAccessToken = t }
export const getStoredRefreshToken = () => { try { return localStorage.getItem(REFRESH_TOKEN_KEY) } catch { return null } }
export const setStoredRefreshToken = (t: string | null) => {
  try { if (t) localStorage.setItem(REFRESH_TOKEN_KEY, t); else localStorage.removeItem(REFRESH_TOKEN_KEY) } catch {}
}
export const configureApi = (h: { onAuthFailure?: () => void; onTokenRefreshed?: (t: string) => void }) => {
  if (h.onAuthFailure !== undefined) onAuthFailureCallback = h.onAuthFailure
  if (h.onTokenRefreshed !== undefined) onTokenRefreshedCallback = h.onTokenRefreshed
}
export const resetApiState = () => {
  currentAccessToken = null; refreshPromise = null; onAuthFailureCallback = null
  onTokenRefreshedCallback = null; setStoredRefreshToken(null)
}
async function executarRefresh(): Promise<string> {
  const token = getStoredRefreshToken()
  if (!token) throw new ApiError('auth.session.expired', '', 401)
  const res = await globalThis.fetch(`${baseUrl}/api/v1/auth/refresh`, {
    method: 'POST', headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify({ refresh_token: token }),
  })
  if (!res.ok) throw new ApiError('auth.session.expired', '', res.status)
  const json = await res.json(), novo: string = json.data.access_token
  setAccessToken(novo); onTokenRefreshedCallback?.(novo)
  return novo
}
export function renovarSessao(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = executarRefresh()
      .catch((err) => { setAccessToken(null); setStoredRefreshToken(null); onAuthFailureCallback?.(); throw err })
      .finally(() => { refreshPromise = null })
  }
  return refreshPromise
}
export async function apiRequestRaw<T>(
  path: string, options: RequestInit = {}, tokenOverride?: string, isRetry = false,
): Promise<T> {
  const headers = new Headers(options.headers)
  if (!headers.has('Accept')) headers.set('Accept', 'application/json')
  const token = tokenOverride ?? currentAccessToken
  if (token && !headers.has('Authorization')) headers.set('Authorization', `Bearer ${token}`)
  if (options.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const res = await globalThis.fetch(path.startsWith('http') ? path : `${baseUrl}/api/v1${path}`, { ...options, headers })
  if (res.status === 401 && !isRetry && !path.includes('/auth/refresh') && !path.includes('/auth/login')) {
    try {
      const novo = await renovarSessao()
      const retryHeaders = new Headers(options.headers); retryHeaders.set('Authorization', `Bearer ${novo}`)
      return await apiRequestRaw<T>(path, { ...options, headers: retryHeaders }, novo, true)
    } catch {}
  }
  if (!res.ok) {
    let codigo = 'fallback', mensagem = ''
    try {
      const data = await res.json()
      if (typeof data?.error?.codigo === 'string') {
        codigo = data.error.codigo; mensagem = typeof data.error.mensagem === 'string' ? data.error.mensagem : ''
      }
    } catch {}
    throw new ApiError(codigo, mensagem, res.status)
  }
  return res.status === 204 ? (undefined as T) : res.json()
}
export async function apiRequest<T>(path: string, options: RequestInit = {}, tokenOverride?: string): Promise<T> {
  const json = await apiRequestRaw<{ data?: T } | T>(path, options, tokenOverride)
  return json && typeof json === 'object' && 'data' in json ? (json as { data: T }).data : (json as T)
}
let interceptorInstalled = false
export function setupFetchInterceptor() {
  if (interceptorInstalled) return
  interceptorInstalled = true; const original = globalThis.fetch
  globalThis.fetch = async (input: RequestInfo | URL, init: RequestInit = {}) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
    if (!url.includes('/api/v1') || url.includes('/auth/refresh') || url.includes('/auth/login') || (init as { _internal?: boolean })._internal) return original(input, init)
    const headers = new Headers(init.headers || (typeof input === 'object' && 'headers' in input ? input.headers : undefined))
    if (!headers.has('Accept')) headers.set('Accept', 'application/json')
    if (currentAccessToken && !headers.has('Authorization')) headers.set('Authorization', `Bearer ${currentAccessToken}`)
    if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
    const res = await original(input, { ...init, headers, _internal: true } as RequestInit)
    if (res.status === 401 && !(init as { _retry?: boolean })._retry) {
      try {
        const novo = await renovarSessao()
        const retryHeaders = new Headers(init.headers); retryHeaders.set('Authorization', `Bearer ${novo}`)
        return await original(input, { ...init, headers: retryHeaders, _internal: true, _retry: true } as RequestInit)
      } catch { return res }
    }
    return res
  }
}
