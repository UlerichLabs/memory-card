import { apiRequest } from "@/lib/api";
import type {
  AdicionarItemPayload,
  AdicionarItensLotePayload,
  AdicionarItensLoteResposta,
  AtualizarListaPayload,
  CatalogoResposta,
  CriarListaPayload,
  FiltroCatalogo,
  IGDBFranquiaSugestao,
  ListaDetalhada,
  ListaItem,
  ListaResumo,
  OpcaoOrigem,
} from "@/types/listas";

function queryCatalogo(filtro: FiltroCatalogo): string {
  const params = new URLSearchParams({
    origem: filtro.origem,
    id: String(filtro.id),
  });
  if (filtro.genero_id) params.set("genero_id", String(filtro.genero_id));
  if (filtro.plataforma_id)
    params.set("plataforma_id", String(filtro.plataforma_id));
  if (filtro.busca?.trim()) params.set("busca", filtro.busca.trim());
  if (filtro.ordenar) params.set("ordenar", filtro.ordenar);
  if (filtro.pagina) params.set("pagina", String(filtro.pagina));
  if (filtro.por_pagina) params.set("por_pagina", String(filtro.por_pagina));
  if (filtro.somente_sugeridos !== undefined) {
    params.set("somente_sugeridos", String(filtro.somente_sugeridos));
  }
  return params.toString();
}

export const listasService = {
  listar: (token?: string, signal?: AbortSignal) =>
    apiRequest<ListaResumo[]>("/listas", { method: "GET", signal }, token),
  obterPorId: (id: number, token?: string, signal?: AbortSignal) =>
    apiRequest<ListaDetalhada>(
      `/listas/${id}`,
      { method: "GET", signal },
      token,
    ),
  criar: (payload: CriarListaPayload, token?: string) =>
    apiRequest<ListaDetalhada>(
      "/listas",
      { method: "POST", body: JSON.stringify(payload) },
      token,
    ),
  atualizar: (id: number, payload: AtualizarListaPayload, token?: string) =>
    apiRequest<ListaDetalhada>(
      `/listas/${id}`,
      { method: "PUT", body: JSON.stringify(payload) },
      token,
    ),
  excluir: (id: number, token?: string) =>
    apiRequest<void>(`/listas/${id}`, { method: "DELETE" }, token),
  adicionarItem: (
    listaId: number,
    payload: AdicionarItemPayload,
    token?: string,
  ) =>
    apiRequest<ListaItem>(
      `/listas/${listaId}/itens`,
      { method: "POST", body: JSON.stringify(payload) },
      token,
    ),
  adicionarItensLote: (
    listaId: number,
    payload: AdicionarItensLotePayload,
    token?: string,
  ) =>
    apiRequest<AdicionarItensLoteResposta>(
      `/listas/${listaId}/itens/lote`,
      { method: "POST", body: JSON.stringify(payload) },
      token,
    ),
  removerItem: (listaId: number, itemId: number, token?: string) =>
    apiRequest<void>(
      `/listas/${listaId}/itens/${itemId}`,
      { method: "DELETE" },
      token,
    ),
  reordenarItens: (listaId: number, itemIds: number[], token?: string) =>
    apiRequest<ListaItem[]>(
      `/listas/${listaId}/ordem`,
      { method: "PUT", body: JSON.stringify({ item_ids: itemIds }) },
      token,
    ),
  associarZeramento: (
    listaId: number,
    itemId: number,
    jogoZeradoId: number,
    token?: string,
  ) =>
    apiRequest<ListaItem>(
      `/listas/${listaId}/itens/${itemId}/zeramento`,
      { method: "PUT", body: JSON.stringify({ jogo_zerado_id: jogoZeradoId }) },
      token,
    ),
  buscarCatalogo: (
    filtro: FiltroCatalogo,
    token?: string,
    signal?: AbortSignal,
  ) =>
    apiRequest<CatalogoResposta>(
      `/catalogo/jogos?${queryCatalogo(filtro)}`,
      { method: "GET", signal },
      token,
    ),
  buscarFranquias: async (
    termo: string,
    token?: string,
    signal?: AbortSignal,
  ): Promise<IGDBFranquiaSugestao[]> => {
    try {
      return await apiRequest<IGDBFranquiaSugestao[]>(
        `/igdb/franquias/busca?q=${encodeURIComponent(termo)}`,
        { method: "GET", signal },
        token,
      );
    } catch (err: unknown) {
      if (
        signal?.aborted ||
        (err instanceof DOMException && err.name === "AbortError")
      )
        return [];
      throw err;
    }
  },
  listarPlataformas: async (token?: string, signal?: AbortSignal) => {
    const resposta = await apiRequest<Array<{ id: number; name: string }>>(
      "/igdb/plataformas",
      { method: "GET", signal },
      token,
    );
    return resposta.map((item) => ({ id: item.id, nome: item.name }));
  },
  listarGeneros: (token?: string, signal?: AbortSignal) =>
    apiRequest<OpcaoOrigem[]>(
      "/igdb/generos",
      { method: "GET", signal },
      token,
    ),
};

export { queryCatalogo };
