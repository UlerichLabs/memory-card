import { createContext, createElement, useContext, useRef, useState, type ReactNode } from 'react'
import { abandonadosService } from '@/lib/services/abandonadosService'
import { AuthContext } from '@/store/authStore'
import { JogandoContext } from './jogandoStore'
import type {
  AbandonadosListagemMeta,
  JogoAbandonado,
  ListarAbandonadosParams,
  OpcoesFiltrosAbandonados,
  SalvarAbandonadoPayload,
} from '@/types/abandonados'

export interface OrigemFilaInfo {
  listaNome: string
  onSalvo?: () => Promise<void>
}

export interface ModalAbandonadoOpcoes {
  valoresIniciais?: Partial<JogoAbandonado>
  origemFila?: OrigemFilaInfo
  aviso?: string
  onSalvo?: () => Promise<void> | void
}

export interface AbandonadosStore {
  jogos: JogoAbandonado[]
  meta: AbandonadosListagemMeta
  filtros: OpcoesFiltrosAbandonados
  totalGeral: number
  isLoading: boolean
  carregado: boolean
  error: string | null
  aviso: string | null
  isModalOpen: boolean
  jogoEmEdicao: JogoAbandonado | null
  modalOpcoes: ModalAbandonadoOpcoes | null
  isExcluirModalOpen: boolean
  jogoParaExcluir: JogoAbandonado | null
  abrirModalCriacao: (opcoes?: ModalAbandonadoOpcoes) => void
  abrirModalEdicao: (jogo: JogoAbandonado) => void
  fecharModal: () => void
  abrirModalExcluir: (jogo: JogoAbandonado) => void
  fecharModalExcluir: () => void
  carregarJogos: (params?: ListarAbandonadosParams, signal?: AbortSignal) => Promise<void>
  carregarFiltros: (signal?: AbortSignal) => Promise<void>
  carregarTotal: (signal?: AbortSignal) => Promise<void>
  criarJogo: (payload: SalvarAbandonadoPayload) => Promise<JogoAbandonado>
  atualizarJogo: (id: number, payload: SalvarAbandonadoPayload) => Promise<JogoAbandonado>
  excluirJogo: (id: number) => Promise<void>
  limparErro: () => void
  definirAviso: (aviso: string | null) => void
  limparAviso: () => void
}

export const AbandonadosContext = createContext<AbandonadosStore | null>(null)

const META_PADRAO: AbandonadosListagemMeta = { pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 }
const FILTROS_PADRAO: OpcoesFiltrosAbandonados = { consoles: [] }

export function AbandonadosProvider({ children, token }: { children: ReactNode; token?: string }) {
  const auth = useContext(AuthContext)
  const jogando = useContext(JogandoContext)
  const effectiveToken = token ?? auth?.sessao?.access_token
  const [jogos, setJogos] = useState<JogoAbandonado[]>([])
  const [meta, setMeta] = useState<AbandonadosListagemMeta>(META_PADRAO)
  const [filtros, setFiltros] = useState<OpcoesFiltrosAbandonados>(FILTROS_PADRAO)
  const [totalGeral, setTotalGeral] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const [carregado, setCarregado] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [aviso, setAviso] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [jogoEmEdicao, setJogoEmEdicao] = useState<JogoAbandonado | null>(null)
  const [modalOpcoes, setModalOpcoes] = useState<ModalAbandonadoOpcoes | null>(null)
  const [isExcluirModalOpen, setIsExcluirModalOpen] = useState(false)
  const [jogoParaExcluir, setJogoParaExcluir] = useState<JogoAbandonado | null>(null)
  const ultimosParamsRef = useRef<ListarAbandonadosParams | undefined>(undefined)

  const carregarJogos = async (params?: ListarAbandonadosParams, signal?: AbortSignal) => {
    ultimosParamsRef.current = params
    setIsLoading(true)
    setError(null)
    try {
      const res = await abandonadosService.listar(params, effectiveToken, signal)
      if (!signal?.aborted) {
        setJogos(res.data)
        setMeta(res.meta)
      }
    } catch (err) {
      if (!signal?.aborted) {
        setError(err instanceof Error ? err.message : 'Erro ao carregar abandonados')
        throw err
      }
    } finally {
      if (!signal?.aborted) {
        setIsLoading(false)
        setCarregado(true)
      }
    }
  }

  const carregarFiltros = async (signal?: AbortSignal) => {
    try {
      const res = await abandonadosService.obterFiltros(effectiveToken, signal)
      if (!signal?.aborted) setFiltros(res)
    } catch (err) {
      if (!signal?.aborted) setError(err instanceof Error ? err.message : 'Erro ao carregar filtros')
    }
  }

  const carregarTotal = async (signal?: AbortSignal) => {
    try {
      const res = await abandonadosService.obterTotal(effectiveToken, signal)
      if (!signal?.aborted) setTotalGeral(res.total)
    } catch (err) {
      if (!signal?.aborted) setError(err instanceof Error ? err.message : 'Erro ao carregar total')
    }
  }

  const recarregarAposMutacao = async () => {
    await Promise.allSettled([carregarJogos(ultimosParamsRef.current), carregarFiltros(), carregarTotal()])
  }

  const store: AbandonadosStore = {
    jogos, meta, filtros, totalGeral, isLoading, carregado, error, aviso,
    isModalOpen, jogoEmEdicao, modalOpcoes, isExcluirModalOpen, jogoParaExcluir,
    limparErro: () => setError(null),
    definirAviso: (novoAviso) => setAviso(novoAviso),
    limparAviso: () => setAviso(null),
    abrirModalCriacao: (opcoes) => {
      setJogoEmEdicao(null); setModalOpcoes(opcoes ?? null); setIsModalOpen(true)
    },
    abrirModalEdicao: (jogo) => {
      setModalOpcoes(null); setJogoEmEdicao(jogo); setIsModalOpen(true)
    },
    fecharModal: () => {
      setIsModalOpen(false); setJogoEmEdicao(null); setModalOpcoes(null)
    },
    abrirModalExcluir: (jogo) => {
      setJogoParaExcluir(jogo); setIsExcluirModalOpen(true)
    },
    fecharModalExcluir: () => {
      setIsExcluirModalOpen(false); setJogoParaExcluir(null)
    },
    carregarJogos, carregarFiltros, carregarTotal,
    criarJogo: async (payload) => {
      setIsLoading(true); setError(null)
      try {
        const criado = await abandonadosService.criar(payload, effectiveToken)
        await recarregarAposMutacao()
        try { await jogando?.carregar() } catch {}
        return criado
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao criar abandono')
        throw err
      } finally { setIsLoading(false) }
    },
    atualizarJogo: async (id, payload) => {
      setIsLoading(true); setError(null)
      try {
        const atualizado = await abandonadosService.atualizar(id, payload, effectiveToken)
        await recarregarAposMutacao()
        return atualizado
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao atualizar abandono')
        throw err
      } finally { setIsLoading(false) }
    },
    excluirJogo: async (id) => {
      setIsLoading(true); setError(null)
      try {
        await abandonadosService.excluir(id, effectiveToken)
        await recarregarAposMutacao()
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao excluir abandono')
        throw err
      } finally { setIsLoading(false) }
    },
  }

  return createElement(AbandonadosContext.Provider, { value: store }, children)
}

export function useAbandonadosStore(): AbandonadosStore {
  const store = useContext(AbandonadosContext)
  if (!store) throw new Error('useAbandonadosStore deve ser utilizado dentro de um AbandonadosProvider')
  return store
}
