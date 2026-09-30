import type {
  JogoAbandonado,
  ListarAbandonadosParams,
  ListarAbandonadosResposta,
  OpcoesFiltrosAbandonados,
  SalvarAbandonadoPayload,
  TotalAbandonadosResposta,
} from '@/types/abandonados'
import { ehCancelado, montarQueryStringAbandonados } from '@/components/abandonados/abandonados.utils'

export class AbandonadosApiError extends Error {
  readonly codigo: string
  readonly status?: number
  constructor(codigo: string, message: string, status?: number) {
    super(message)
    this.name = 'AbandonadosApiError'
    this.codigo = codigo
    this.status = status
  }
}

const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')

async function requestRaw<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
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
    throw new AbandonadosApiError(codigo, mensagem, response.status)
  }
  return response.status === 204 ? (undefined as T) : response.json()
}

async function request<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  const json = await requestRaw<{ data: T }>(path, options, accessToken)
  return json.data
}

export const abandonadosService = {
  criar: (payload: SalvarAbandonadoPayload, token?: string) =>
    request<JogoAbandonado>('/jogos-abandonados', { method: 'POST', body: JSON.stringify(payload) }, token),
  atualizar: (id: number, payload: SalvarAbandonadoPayload, token?: string) =>
    request<JogoAbandonado>(`/jogos-abandonados/${id}`, { method: 'PUT', body: JSON.stringify(payload) }, token),
  excluir: (id: number, token?: string) =>
    requestRaw<void>(`/jogos-abandonados/${id}`, { method: 'DELETE' }, token),
  obterPorId: (id: number, token?: string, signal?: AbortSignal) =>
    request<JogoAbandonado>(`/jogos-abandonados/${id}`, { method: 'GET', signal }, token),
  listar: async (
    params?: ListarAbandonadosParams,
    token?: string,
    signal?: AbortSignal
  ): Promise<ListarAbandonadosResposta> => {
    const qs = montarQueryStringAbandonados(params)
    try {
      return await requestRaw<ListarAbandonadosResposta>(
        `/jogos-abandonados${qs ? `?${qs}` : ''}`,
        { method: 'GET', signal },
        token
      )
    } catch (err) {
      if (ehCancelado(err, signal)) {
        return {
          data: [],
          meta: { pagina: params?.pagina ?? 1, por_pagina: params?.por_pagina ?? 12, total: 0, total_paginas: 0 },
        }
      }
      throw err
    }
  },
  obterFiltros: async (token?: string, signal?: AbortSignal): Promise<OpcoesFiltrosAbandonados> => {
    try {
      return await request<OpcoesFiltrosAbandonados>('/jogos-abandonados/filtros', { method: 'GET', signal }, token)
    } catch (err) {
      if (ehCancelado(err, signal)) return { consoles: [] }
      throw err
    }
  },
  obterTotal: async (token?: string, signal?: AbortSignal): Promise<TotalAbandonadosResposta> => {
    try {
      return await request<TotalAbandonadosResposta>('/jogos-abandonados/total', { method: 'GET', signal }, token)
    } catch (err) {
      if (ehCancelado(err, signal)) return { total: 0 }
      throw err
    }
  },
}
