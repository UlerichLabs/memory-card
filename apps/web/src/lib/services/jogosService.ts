import type {
  JogoZeradoDTO, IGDBJogoSugestao, SalvarJogoPayload, ListarJogosParams,
  ListarJogosResposta, ListagemMeta, OpcoesFiltrosDTO, Dificuldade,
} from '@/types/jogos'

export type { Dificuldade, JogoZeradoDTO, IGDBJogoSugestao, SalvarJogoPayload, ListarJogosParams, ListarJogosResposta, ListagemMeta, OpcoesFiltrosDTO }

export class JogosApiError extends Error {
  readonly codigo: string; readonly status?: number
  constructor(codigo: string, message: string, status?: number) {
    super(message); this.name = 'JogosApiError'; this.codigo = codigo; this.status = status
  }
}

export const JOGOS_CAMPO_ERRO_MENSAGENS: Record<string, { campo: string; mensagem: string }> = {
  'jogos.nome_muito_longo': { campo: 'nome', mensagem: 'O nome deve ter no máximo 200 caracteres.' },
  'jogos.console_muito_longo': { campo: 'console', mensagem: 'A plataforma deve ter no máximo 100 caracteres.' },
  'jogos.genero_muito_longo': { campo: 'genero', mensagem: 'O gênero deve ter no máximo 150 caracteres.' },
  'jogos.tipo_muito_longo': { campo: 'tipo', mensagem: 'O tipo deve ter no máximo 50 caracteres.' },
  'jogos.review_muito_longo': { campo: 'review', mensagem: 'A review deve ter no máximo 5.000 caracteres.' },
}

export const JOGOS_ERRO_GENERICO = 'Não foi possível salvar o registro. Tente novamente.'
const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')

async function requestRaw<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`)
  if (options.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const response = await fetch(`${baseUrl}/api/v1${path}`, { ...options, headers })
  if (!response.ok) {
    let codigo = 'fallback', mensagem = ''
    try {
      const data = await response.json()
      if (typeof data?.error?.codigo === 'string') {
        codigo = data.error.codigo; mensagem = typeof data.error.mensagem === 'string' ? data.error.mensagem : ''
      }
    } catch {}
    throw new JogosApiError(codigo, mensagem, response.status)
  }
  return response.status === 204 ? (undefined as T) : response.json()
}

async function request<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  const json = await requestRaw<{ data: T }>(path, options, accessToken)
  return json.data
}

export function montarQueryString(params?: ListarJogosParams): string {
  if (!params) return ''
  const sp = new URLSearchParams()
  if (params.pagina && params.pagina > 1) sp.set('pagina', String(params.pagina))
  if (params.por_pagina && params.por_pagina !== 24) sp.set('por_pagina', String(params.por_pagina))
  if (params.busca?.trim()) sp.set('busca', params.busca.trim())
  if (params.console?.trim()) sp.set('console', params.console.trim())
  if (params.genero?.trim()) sp.set('genero', params.genero.trim())
  if (params.tipo?.trim()) sp.set('tipo', params.tipo.trim())
  if (params.nota_min !== undefined && params.nota_min !== null) sp.set('nota_min', String(params.nota_min))
  if (params.nota_max !== undefined && params.nota_max !== null) sp.set('nota_max', String(params.nota_max))
  if (params.ano !== undefined && params.ano !== null) sp.set('ano', String(params.ano))
  if (params.dificuldade) sp.set('dificuldade', params.dificuldade)
  return sp.toString()
}

function ehCancelado(err: unknown, sig?: AbortSignal): boolean {
  return Boolean(sig?.aborted || (err instanceof DOMException && err.name === 'AbortError') || (err instanceof JogosApiError && err.status === 499))
}

export const jogosService = {
  criar: (payload: SalvarJogoPayload, token?: string) => request<JogoZeradoDTO>('/jogos', { method: 'POST', body: JSON.stringify(payload) }, token),
  atualizar: (id: number, payload: SalvarJogoPayload, token?: string) => request<JogoZeradoDTO>(`/jogos/${id}`, { method: 'PUT', body: JSON.stringify(payload) }, token),
  excluir: (id: number, token?: string) => request<void>(`/jogos/${id}`, { method: 'DELETE' }, token),
  buscarIGDB: async (termo: string, token?: string, signal?: AbortSignal) => {
    try {
      return await request<IGDBJogoSugestao[]>(`/igdb/jogos/busca?q=${encodeURIComponent(termo)}`, { method: 'GET', signal }, token)
    } catch (err) {
      if (ehCancelado(err, signal)) return []
      throw err
    }
  },
  obterDetalhesIGDB: (id: number, token?: string) => request<IGDBJogoSugestao>(`/jogos/igdb/${id}`, { method: 'GET' }, token),
  listar: async (params?: ListarJogosParams, token?: string, signal?: AbortSignal): Promise<ListarJogosResposta> => {
    const qs = montarQueryString(params)
    try {
      return await requestRaw<ListarJogosResposta>(`/jogos${qs ? `?${qs}` : ''}`, { method: 'GET', signal }, token)
    } catch (err) {
      if (ehCancelado(err, signal)) return { data: [], meta: { pagina: params?.pagina ?? 1, por_pagina: params?.por_pagina ?? 24, total: 0, total_paginas: 0 } }
      throw err
    }
  },
  obterFiltros: async (token?: string, signal?: AbortSignal): Promise<OpcoesFiltrosDTO> => {
    try {
      return await request<OpcoesFiltrosDTO>('/jogos/filtros', { method: 'GET', signal }, token)
    } catch (err) {
      if (ehCancelado(err, signal)) return { consoles: [], generos: [], tipos: [], anos: [] }
      throw err
    }
  },
}
