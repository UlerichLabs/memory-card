import { createContext, createElement, useCallback, useContext, useState, type ReactNode } from 'react'
import { AuthContext } from '@/store/authStore'
import { listasService } from '@/lib/services/listasService'
import { ordenarListas } from '@/components/listas/listas.utils'
import { resolverMensagemErro } from '@/components/listas/listas.constants'
import { ApiError } from '@/lib/api'
import type {
  ListaResumo,
  ListaDetalhada,
  ListaItem,
  CriarListaPayload,
  AtualizarListaPayload,
  AdicionarItemPayload,
  AdicionarItensLoteResposta,
  CriarListaItemPayload,
  FiltroAba,
  ListasStore,
} from '@/types/listas'

export type { ListasStore, FiltroAba }

export const ListasContext = createContext<ListasStore | null>(null)

export interface ListasProviderProps {
  children: ReactNode
  token?: string
  initialListas?: ListaResumo[]
  initialListaAberta?: ListaDetalhada | null
}

export function ListasProvider({ children, token, initialListas = [], initialListaAberta = null }: ListasProviderProps) {
  const auth = useContext(AuthContext)
  const effectiveToken = token ?? auth?.sessao?.access_token
  const [listas, setListas] = useState<ListaResumo[]>(ordenarListas(initialListas))
  const [listaAberta, setListaAberta] = useState<ListaDetalhada | null>(initialListaAberta)
  const [isLoading, setIsLoading] = useState(false)
  const [isLoadingDetalhe, setIsLoadingDetalhe] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [filtroAba, setFiltroAba] = useState<FiltroAba>('todos')
  const [isNovaListaOpen, setIsNovaListaOpen] = useState(false)
  const [listaEmEdicao, setListaEmEdicao] = useState<ListaResumo | null>(null)
  const [isExcluirListaOpen, setIsExcluirListaOpen] = useState(false)
  const [listaParaExcluir, setListaParaExcluir] = useState<ListaResumo | null>(null)

  const carregarListas = useCallback(async (signal?: AbortSignal): Promise<ListaResumo[]> => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await listasService.listar(effectiveToken, signal)
      const ordenadas = ordenarListas(data)
      if (!signal?.aborted) setListas(ordenadas)
      return ordenadas
    } catch (err: unknown) {
      if (!signal?.aborted) {
        const msg = err instanceof ApiError ? resolverMensagemErro(err.codigo) : 'Erro ao carregar listas'
        setError(msg)
        throw err
      }
      return []
    } finally {
      if (!signal?.aborted) setIsLoading(false)
    }
  }, [effectiveToken])

  const abrirLista = useCallback(async (id: number, signal?: AbortSignal): Promise<ListaDetalhada> => {
    setIsLoadingDetalhe(true)
    setError(null)
    setFiltroAba('todos')
    try {
      const detalhe = await listasService.obterPorId(id, effectiveToken, signal)
      if (!signal?.aborted) setListaAberta(detalhe)
      return detalhe
    } catch (err: unknown) {
      if (!signal?.aborted) {
        const msg = err instanceof ApiError ? resolverMensagemErro(err.codigo) : 'Erro ao abrir lista'
        setError(msg)
        throw err
      }
      throw err
    } finally {
      if (!signal?.aborted) setIsLoadingDetalhe(false)
    }
  }, [effectiveToken])

  const criarLista = useCallback(async (payload: CriarListaPayload): Promise<ListaDetalhada> => {
    setError(null)
    const criada = await listasService.criar(payload, effectiveToken)
    await carregarListas()
    setListaAberta(criada)
    return criada
  }, [effectiveToken, carregarListas])

  const atualizarLista = useCallback(async (id: number, payload: AtualizarListaPayload): Promise<ListaDetalhada> => {
    setError(null)
    const atualizada = await listasService.atualizar(id, payload, effectiveToken)
    await carregarListas()
    setListaAberta(atualizada)
    return atualizada
  }, [effectiveToken, carregarListas])

  const excluirLista = useCallback(async (id: number): Promise<void> => {
    setError(null)
    await listasService.excluir(id, effectiveToken)
    if (listaAberta?.id === id) setListaAberta(null)
    await carregarListas()
  }, [effectiveToken, listaAberta?.id, carregarListas])

  const recarregarAbertaEResumo = useCallback(async (listaId: number) => {
    await Promise.all([
      carregarListas(),
      listasService.obterPorId(listaId, effectiveToken).then(setListaAberta),
    ])
  }, [effectiveToken, carregarListas])

  const adicionarItem = useCallback(async (payload: AdicionarItemPayload): Promise<ListaItem> => {
    if (!listaAberta) throw new Error('Nenhuma lista aberta')
    setError(null)
    const item = await listasService.adicionarItem(listaAberta.id, payload, effectiveToken)
    await recarregarAbertaEResumo(listaAberta.id)
    return item
  }, [listaAberta, effectiveToken, recarregarAbertaEResumo])

  const removerItem = useCallback(async (itemId: number): Promise<void> => {
    if (!listaAberta) throw new Error('Nenhuma lista aberta')
    setError(null)
    await listasService.removerItem(listaAberta.id, itemId, effectiveToken)
    await recarregarAbertaEResumo(listaAberta.id)
  }, [listaAberta, effectiveToken, recarregarAbertaEResumo])

  const reordenarItens = useCallback(async (itemIds: number[]): Promise<void> => {
    if (!listaAberta) throw new Error('Nenhuma lista aberta')
    setError(null)
    const itensAnteriores = [...listaAberta.itens]
    const itemMap = new Map(listaAberta.itens.map((it) => [it.id, it]))
    const itensOtimistas = itemIds
      .map((id, idx) => {
        const it = itemMap.get(id)
        return it ? { ...it, posicao: idx + 1 } : null
      })
      .filter((it): it is ListaItem => it !== null)
    setListaAberta({ ...listaAberta, itens: itensOtimistas })
    try {
      const novosItens = await listasService.reordenarItens(listaAberta.id, itemIds, effectiveToken)
      setListaAberta((prev) => (prev ? { ...prev, itens: novosItens } : null))
    } catch (err: unknown) {
      setListaAberta((prev) => (prev ? { ...prev, itens: itensAnteriores } : null))
      const msg = err instanceof ApiError ? resolverMensagemErro(err.codigo) : 'Não foi possível reordenar os jogos.'
      setError(msg)
      throw err
    }
  }, [listaAberta, effectiveToken])

  const associarZeramento = useCallback(async (itemId: number, jogoZeradoId: number): Promise<void> => {
    if (!listaAberta) throw new Error('Nenhuma lista aberta')
    setError(null)
    await listasService.associarZeramento(listaAberta.id, itemId, jogoZeradoId, effectiveToken)
    await recarregarAbertaEResumo(listaAberta.id)
  }, [listaAberta, effectiveToken, recarregarAbertaEResumo])

  const adicionarItensLote = useCallback(async (itens: CriarListaItemPayload[]): Promise<AdicionarItensLoteResposta> => {
    if (!listaAberta) throw new Error('Nenhuma lista aberta')
    setError(null)
    const resp = await listasService.adicionarItensLote(listaAberta.id, { itens }, effectiveToken)
    await recarregarAbertaEResumo(listaAberta.id)
    return resp
  }, [listaAberta, effectiveToken, recarregarAbertaEResumo])

  const store: ListasStore = {
    listas,
    listaAberta,
    isLoading,
    isLoadingDetalhe,
    error,
    filtroAba,
    isNovaListaOpen,
    listaEmEdicao,
    isExcluirListaOpen,
    listaParaExcluir,
    setFiltroAba,
    limparErro: () => setError(null),
    abrirModalCriar: () => { setListaEmEdicao(null); setIsNovaListaOpen(true) },
    abrirModalEditar: (lista) => { setListaEmEdicao(lista); setIsNovaListaOpen(true) },
    fecharModalNovaLista: () => { setIsNovaListaOpen(false); setListaEmEdicao(null) },
    abrirModalExcluir: (lista) => { setListaParaExcluir(lista); setIsExcluirListaOpen(true) },
    fecharModalExcluir: () => { setIsExcluirListaOpen(false); setListaParaExcluir(null) },
    carregarListas,
    abrirLista,
    criarLista,
    atualizarLista,
    excluirLista,
    adicionarItem,
    removerItem,
    reordenarItens,
    associarZeramento,
    adicionarItensLote,
  }

  return createElement(ListasContext.Provider, { value: store }, children)
}

export function useListasStore(): ListasStore {
  const store = useContext(ListasContext)
  if (!store) throw new Error('useListasStore deve ser utilizado dentro de um ListasProvider')
  return store
}
