export interface CadastroPayload {
  nome: string
  email: string
  senha: string
}

export interface UsuarioDTO {
  id: number
  nome: string
  email: string
  idioma: string
  created_at: string
}

export type LoginPayload = Pick<CadastroPayload, 'email' | 'senha'>
export interface SessaoDTO {
  access_token: string
  refresh_token: string
  usuario: UsuarioDTO
}

export class AuthApiError extends Error {
  readonly codigo: string
  readonly status?: number

  constructor(codigo: string, message: string, status?: number) {
    super(message)
    this.name = 'AuthApiError'
    this.codigo = codigo
    this.status = status
  }
}

const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')

async function request<T>(path: string, options: RequestInit): Promise<T> {
  const response = await fetch(`${baseUrl}/api/v1${path}`, options)
  if (!response.ok) {
    let codigo = 'fallback'
    let mensagem = ''
    try {
      const data = await response.json()
      if (typeof data?.error?.codigo === 'string') {
        codigo = data.error.codigo
        mensagem = typeof data.error.mensagem === 'string' ? data.error.mensagem : ''
      }
    } catch {
      codigo = 'fallback'
    }
    throw new AuthApiError(codigo, mensagem, response.status)
  }
  if (response.status === 204) return undefined as T
  const json = await response.json()
  return json.data
}

function post<T>(path: string, payload: unknown): Promise<T> {
  return request(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify(payload),
  })
}

export const authService = {
  cadastrar: (payload: CadastroPayload) => post<UsuarioDTO>('/auth/register', payload),
  login: (payload: LoginPayload) => post<SessaoDTO>('/auth/login', payload),
  refresh: (refreshToken: string) =>
    post<Pick<SessaoDTO, 'access_token'>>('/auth/refresh', { refresh_token: refreshToken }),
  logout: (accessToken: string, refreshToken: string) =>
    authService.authenticatedRequest<void>('/auth/logout', accessToken, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    }),
  authenticatedRequest<T>(path: string, accessToken: string, options: RequestInit = {}) {
    const headers = new Headers(options.headers)
    headers.set('Authorization', `Bearer ${accessToken}`)
    headers.set('Accept', 'application/json')
    return request<T>(path, { ...options, headers })
  },
}
