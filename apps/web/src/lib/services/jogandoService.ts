import { ApiError } from '@/lib/api'
import type { CriarJogoEmAndamentoPayload, JogoEmAndamento, ListarJogandoResposta } from '@/types/jogando'

const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')

async function requestRaw<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (options.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const response = await fetch(`${baseUrl}/api/v1${path}`, { ...options, headers })
  if (!response.ok) {
    let codigo = 'fallback'
    let mensagem = ''
    try {
      const data = await response.json()
      codigo = typeof data?.error?.codigo === 'string' ? data.error.codigo : codigo
      mensagem = typeof data?.error?.mensagem === 'string' ? data.error.mensagem : ''
    } catch {}
    throw new ApiError(codigo, mensagem, response.status)
  }
  return response.status === 204 ? (undefined as T) : response.json()
}

export const jogandoService = {
  listar: async (token?: string, signal?: AbortSignal): Promise<JogoEmAndamento[]> => {
    const resposta = await requestRaw<ListarJogandoResposta>('/jogando', { method: 'GET', signal }, token)
    return resposta.data ?? []
  },
  criar: async (payload: CriarJogoEmAndamentoPayload, token?: string): Promise<JogoEmAndamento> => {
    const resposta = await requestRaw<{ data: JogoEmAndamento }>(
      '/jogando',
      { method: 'POST', body: JSON.stringify(payload) },
      token
    )
    return resposta.data
  },
  remover: (id: number, token?: string) =>
    requestRaw<void>(`/jogando/${id}`, { method: 'DELETE' }, token),
}
