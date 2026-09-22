export type Dificuldade = 'C' | 'B' | 'A' | 'AA' | 'AAA'

export interface JogoZeradoDTO {
  id: number
  usuario_id: number
  igdb_id?: number | null
  nome: string
  console: string
  genero?: string
  tipo?: string
  iniciado_em?: string | null
  finalizado_em: string
  tempo_jogado: number
  nota: number
  dificuldade: Dificuldade
  condicao_zeramento?: string
  destaque: boolean
  igdb_capa_url?: string
  igdb_descricao?: string
  created_at?: string
  updated_at?: string
}

export interface IGDBJogoSugestao {
  id: number
  name: string
  cover?: { id?: number; url?: string }
  first_release_date?: number
  summary?: string
  platforms?: Array<{ id: number; name: string }>
  genres?: Array<{ id: number; name: string }>
}

export interface SalvarJogoPayload {
  igdb_id?: number | null
  nome: string
  console: string
  genero?: string
  tipo?: string
  iniciado_em?: string | null
  finalizado_em: string
  tempo_jogado_horas?: number; tempo_jogado_minutos?: number; tempo_jogado_segundos?: number; tempo_jogado?: number
  nota: number
  dificuldade: Dificuldade
  condicao_zeramento?: string
  destaque: boolean
  igdb_capa_url?: string
  igdb_descricao?: string
}

export class JogosApiError extends Error {
  readonly codigo: string
  readonly status?: number

  constructor(codigo: string, message: string, status?: number) {
    super(message)
    this.name = 'JogosApiError'
    this.codigo = codigo
    this.status = status
  }
}

const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')

async function request<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`)
  if (options.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')

  const response = await fetch(`${baseUrl}/api/v1${path}`, { ...options, headers })
  if (!response.ok) {
    let codigo = 'fallback'
    let mensagem = ''
    try {
      const data = await response.json()
      if (typeof data?.error?.codigo === 'string') {
        codigo = data.error.codigo
        mensagem = typeof data.error.mensagem === 'string' ? data.error.mensagem : ''
      }
    } catch {}
    throw new JogosApiError(codigo, mensagem, response.status)
  }
  if (response.status === 204) return undefined as T
  const json = await response.json()
  return json.data
}

export const jogosService = {
  criar: (payload: SalvarJogoPayload, token?: string) =>
    request<JogoZeradoDTO>('/jogos', { method: 'POST', body: JSON.stringify(payload) }, token),
  atualizar: (id: number, payload: SalvarJogoPayload, token?: string) =>
    request<JogoZeradoDTO>(`/jogos/${id}`, { method: 'PUT', body: JSON.stringify(payload) }, token),
  excluir: (id: number, token?: string) => request<void>(`/jogos/${id}`, { method: 'DELETE' }, token),
  buscarIGDB: (termo: string, token?: string, signal?: AbortSignal) =>
    request<IGDBJogoSugestao[]>(`/igdb/jogos/busca?q=${encodeURIComponent(termo)}`, { method: 'GET', signal }, token),
  obterDetalhesIGDB: (id: number, token?: string) =>
    request<IGDBJogoSugestao>(`/jogos/igdb/${id}`, { method: 'GET' }, token),
}
