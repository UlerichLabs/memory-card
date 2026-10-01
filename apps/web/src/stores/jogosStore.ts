import { createContext, createElement, useContext, useRef, useState, type ReactNode } from 'react'
import {
  jogosService,
  type JogoZeradoDTO,
  type SalvarJogoPayload,
  type IGDBJogoSugestao,
  type ListarJogosParams,
  type ListagemMeta,
  type OpcoesFiltrosDTO,
} from '@/lib/services/jogosService'
import { AuthContext } from '@/store/authStore'

export interface ModalRegistroOpcoes {
  valoresIniciais?: Partial<JogoZeradoDTO>
  aviso?: string
  textoSubmit?: string
  onSalvo?: (jogo: JogoZeradoDTO) => void | Promise<void>
}

export interface JogosStore {
  jogos: JogoZeradoDTO[]
  meta: ListagemMeta
  filtros: OpcoesFiltrosDTO
  isLoading: boolean
  error: string | null
  isModalOpen: boolean
  jogoEmEdicao: JogoZeradoDTO | null
  modalRegistroOpcoes: ModalRegistroOpcoes | null
  abrirModalRegistro: (opcoes?: ModalRegistroOpcoes) => void
  abrirModalEdicao: (jogo: JogoZeradoDTO) => void
  fecharModal: () => void
  carregarJogos: (params?: ListarJogosParams, signal?: AbortSignal) => Promise<void>
  carregarFiltros: (signal?: AbortSignal) => Promise<void>
  limparBiblioteca: () => void
  criarJogo: (payload: SalvarJogoPayload) => Promise<JogoZeradoDTO>
  atualizarJogo: (id: number, payload: SalvarJogoPayload) => Promise<JogoZeradoDTO>
  excluirJogo: (id: number) => Promise<void>
  buscarIGDB: (termo: string, signal?: AbortSignal) => Promise<IGDBJogoSugestao[]>
  obterDetalhesIGDB: (id: number) => Promise<IGDBJogoSugestao>
  obterJogoPorId: (id: number, signal?: AbortSignal) => Promise<JogoZeradoDTO>
  setJogos: (jogos: JogoZeradoDTO[]) => void
  limparErro: () => void
}

export const JogosContext = createContext<JogosStore | null>(null)

export interface JogosProviderProps {
  children: ReactNode
  token?: string
  initialJogos?: JogoZeradoDTO[]
  initialMeta?: ListagemMeta
  initialFiltros?: OpcoesFiltrosDTO
}

const META_PADRAO: ListagemMeta = { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 }
const FILTROS_PADRAO: OpcoesFiltrosDTO = { consoles: [], generos: [], tipos: [], anos: [] }

export function JogosProvider({
  children,
  token,
  initialJogos = [],
  initialMeta,
  initialFiltros = FILTROS_PADRAO,
}: JogosProviderProps) {
  const auth = useContext(AuthContext)
  const effectiveToken = token ?? auth?.sessao?.access_token
  const [jogos, setJogos] = useState<JogoZeradoDTO[]>(initialJogos)
  const [meta, setMeta] = useState<ListagemMeta>(initialMeta ?? { ...META_PADRAO, total: initialJogos.length })
  const [filtros, setFiltros] = useState<OpcoesFiltrosDTO>(initialFiltros)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [jogoEmEdicao, setJogoEmEdicao] = useState<JogoZeradoDTO | null>(null)
  const [modalRegistroOpcoes, setModalRegistroOpcoes] = useState<ModalRegistroOpcoes | null>(null)
  const ultimosParamsRef = useRef<ListarJogosParams | undefined>(undefined)

  const carregarJogos = async (params?: ListarJogosParams, signal?: AbortSignal) => {
    ultimosParamsRef.current = params
    setIsLoading(true)
    setError(null)
    try {
      const res = await jogosService.listar(params, effectiveToken, signal)
      if (!signal?.aborted) {
        setJogos(res.data)
        setMeta(res.meta)
      }
    } catch (err) {
      if (!signal?.aborted) {
        setError(err instanceof Error ? err.message : 'Erro ao carregar jogos')
        throw err
      }
    } finally {
      if (!signal?.aborted) setIsLoading(false)
    }
  }

  const carregarFiltros = async (signal?: AbortSignal) => {
    try {
      const res = await jogosService.obterFiltros(effectiveToken, signal)
      if (!signal?.aborted) setFiltros(res)
    } catch (err) {
      if (!signal?.aborted) {
        setError(err instanceof Error ? err.message : 'Erro ao carregar filtros')
      }
    }
  }

  const recarregarAposMutacao = async () => {
    await Promise.allSettled([
      carregarJogos(ultimosParamsRef.current),
      carregarFiltros(),
    ])
  }

  const store: JogosStore = {
    jogos, meta, filtros, isLoading, error, isModalOpen, jogoEmEdicao, modalRegistroOpcoes, setJogos,
    limparErro: () => setError(null),
    abrirModalRegistro: (opcoes) => {
      setJogoEmEdicao(null)
      setModalRegistroOpcoes(opcoes ?? null)
      setIsModalOpen(true)
    },
    abrirModalEdicao: (jogo) => {
      setModalRegistroOpcoes(null)
      setJogoEmEdicao(jogo)
      setIsModalOpen(true)
    },
    fecharModal: () => {
      setIsModalOpen(false)
      setJogoEmEdicao(null)
      setModalRegistroOpcoes(null)
    },
    carregarJogos,
    carregarFiltros,
    limparBiblioteca: () => {
      setJogos([])
      setMeta(META_PADRAO)
      setError(null)
    },
    async criarJogo(payload) {
      setIsLoading(true); setError(null)
      try {
        const criado = await jogosService.criar(payload, effectiveToken)
        await recarregarAposMutacao()
        return criado
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao criar jogo')
        throw err
      } finally { setIsLoading(false) }
    },
    async atualizarJogo(id, payload) {
      setIsLoading(true); setError(null)
      try {
        const atualizado = await jogosService.atualizar(id, payload, effectiveToken)
        await recarregarAposMutacao()
        return atualizado
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao atualizar jogo')
        throw err
      } finally { setIsLoading(false) }
    },
    async excluirJogo(id) {
      setIsLoading(true); setError(null)
      try {
        await jogosService.excluir(id, effectiveToken)
        await recarregarAposMutacao()
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao excluir jogo')
        throw err
      } finally { setIsLoading(false) }
    },
    buscarIGDB: (termo, signal) => jogosService.buscarIGDB(termo, effectiveToken, signal),
    obterDetalhesIGDB: (id) => jogosService.obterDetalhesIGDB(id, effectiveToken),
    obterJogoPorId: (id, signal) => jogosService.obterPorId(id, effectiveToken, signal),
  }

  return createElement(JogosContext.Provider, { value: store }, children)
}

export function useJogosStore(): JogosStore {
  const store = useContext(JogosContext)
  if (!store) throw new Error('useJogosStore deve ser utilizado dentro de um JogosProvider')
  return store
}
