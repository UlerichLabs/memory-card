import { createContext, createElement, useContext, useState, type ReactNode } from 'react'
import {
  jogosService,
  type JogoZeradoDTO,
  type SalvarJogoPayload,
  type IGDBJogoSugestao,
} from '@/lib/services/jogosService'
import { AuthContext } from '@/store/authStore'

export interface JogosStore {
  jogos: JogoZeradoDTO[]
  isLoading: boolean
  error: string | null
  isModalOpen: boolean
  jogoEmEdicao: JogoZeradoDTO | null
  abrirModalRegistro: () => void
  abrirModalEdicao: (jogo: JogoZeradoDTO) => void
  fecharModal: () => void
  criarJogo: (payload: SalvarJogoPayload) => Promise<JogoZeradoDTO>
  atualizarJogo: (id: number, payload: SalvarJogoPayload) => Promise<JogoZeradoDTO>
  excluirJogo: (id: number) => Promise<void>
  buscarIGDB: (termo: string, signal?: AbortSignal) => Promise<IGDBJogoSugestao[]>
  obterDetalhesIGDB: (id: number) => Promise<IGDBJogoSugestao>
  setJogos: (jogos: JogoZeradoDTO[]) => void
  limparErro: () => void
}

export const JogosContext = createContext<JogosStore | null>(null)

export interface JogosProviderProps {
  children: ReactNode
  token?: string
  initialJogos?: JogoZeradoDTO[]
}

export function JogosProvider({ children, token, initialJogos = [] }: JogosProviderProps) {
  const auth = useContext(AuthContext)
  const effectiveToken = token ?? auth?.sessao?.access_token
  const [jogos, setJogos] = useState<JogoZeradoDTO[]>(initialJogos)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [jogoEmEdicao, setJogoEmEdicao] = useState<JogoZeradoDTO | null>(null)

  const store: JogosStore = {
    jogos, isLoading, error, isModalOpen, jogoEmEdicao, setJogos,
    limparErro: () => setError(null),
    abrirModalRegistro: () => { setJogoEmEdicao(null); setIsModalOpen(true) },
    abrirModalEdicao: (jogo) => { setJogoEmEdicao(jogo); setIsModalOpen(true) },
    fecharModal: () => { setIsModalOpen(false); setJogoEmEdicao(null) },
    async criarJogo(payload) {
      setIsLoading(true); setError(null)
      try {
        const criado = await jogosService.criar(payload, effectiveToken)
        setJogos((prev) => [criado, ...prev])
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
        setJogos((prev) => prev.map((item) => (item.id === id ? atualizado : item)))
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
        setJogos((prev) => prev.filter((item) => item.id !== id))
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao excluir jogo')
        throw err
      } finally { setIsLoading(false) }
    },
    buscarIGDB: (termo, signal) => jogosService.buscarIGDB(termo, effectiveToken, signal),
    obterDetalhesIGDB: (id) => jogosService.obterDetalhesIGDB(id, effectiveToken),
  }

  return createElement(JogosContext.Provider, { value: store }, children)
}

export function useJogosStore(): JogosStore {
  const store = useContext(JogosContext)
  if (!store) throw new Error('useJogosStore deve ser utilizado dentro de um JogosProvider')
  return store
}
