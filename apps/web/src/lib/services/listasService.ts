import { apiRequest } from '@/lib/api'
import type {
  ListaResumo,
  ListaDetalhada,
  ListaItem,
  CriarListaPayload,
  AtualizarListaPayload,
  AdicionarItemPayload,
  SincronizarResposta,
  IGDBFranquiaSugestao,
} from '@/types/listas'

export const listasService = {
  listar: (token?: string, signal?: AbortSignal) =>
    apiRequest<ListaResumo[]>('/listas', { method: 'GET', signal }, token),
  obterPorId: (id: number, token?: string, signal?: AbortSignal) =>
    apiRequest<ListaDetalhada>(`/listas/${id}`, { method: 'GET', signal }, token),
  criar: (payload: CriarListaPayload, token?: string) =>
    apiRequest<ListaDetalhada>('/listas', { method: 'POST', body: JSON.stringify(payload) }, token),
  atualizar: (id: number, payload: AtualizarListaPayload, token?: string) =>
    apiRequest<ListaDetalhada>(`/listas/${id}`, { method: 'PUT', body: JSON.stringify(payload) }, token),
  excluir: (id: number, token?: string) =>
    apiRequest<void>(`/listas/${id}`, { method: 'DELETE' }, token),
  adicionarItem: (listaId: number, payload: AdicionarItemPayload, token?: string) =>
    apiRequest<ListaItem>(`/listas/${listaId}/itens`, { method: 'POST', body: JSON.stringify(payload) }, token),
  removerItem: (listaId: number, itemId: number, token?: string) =>
    apiRequest<void>(`/listas/${listaId}/itens/${itemId}`, { method: 'DELETE' }, token),
  reordenarItens: (listaId: number, itemIds: number[], token?: string) =>
    apiRequest<ListaItem[]>(`/listas/${listaId}/ordem`, { method: 'PUT', body: JSON.stringify({ item_ids: itemIds }) }, token),
  associarZeramento: (listaId: number, itemId: number, jogoZeradoId: number, token?: string) =>
    apiRequest<ListaItem>(`/listas/${listaId}/itens/${itemId}/zeramento`, { method: 'PUT', body: JSON.stringify({ jogo_zerado_id: jogoZeradoId }) }, token),
  sincronizarFranquia: (listaId: number, token?: string) =>
    apiRequest<SincronizarResposta>(`/listas/${listaId}/sincronizar`, { method: 'POST' }, token),
  buscarFranquiasIGDB: async (termo: string, token?: string, signal?: AbortSignal): Promise<IGDBFranquiaSugestao[]> => {
    try {
      return await apiRequest<IGDBFranquiaSugestao[]>(`/igdb/franquias/busca?q=${encodeURIComponent(termo)}`, { method: 'GET', signal }, token)
    } catch (err: unknown) {
      if (signal?.aborted || (err instanceof DOMException && err.name === 'AbortError')) return []
      throw err
    }
  },
}
