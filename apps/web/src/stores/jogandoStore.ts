import { createContext, createElement, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { ApiError } from '@/lib/api'
import { jogandoService } from '@/lib/services/jogandoService'
import { AuthContext } from '@/store/authStore'
import type { CriarJogoEmAndamentoPayload, JogoEmAndamento } from '@/types/jogando'

export interface JogandoStore {
  jogos: JogoEmAndamento[]
  isLoading: boolean
  carregado: boolean
  error: string | null
  aviso: string | null
  isModalOpen: boolean
  carregar: (signal?: AbortSignal) => Promise<void>
  criar: (payload: CriarJogoEmAndamentoPayload) => Promise<JogoEmAndamento>
  remover: (id: number) => Promise<void>
  abrirModalIniciar: () => void
  fecharModal: () => void
  definirAviso: (aviso: string | null) => void
  limparAviso: () => void
}

export const JogandoContext = createContext<JogandoStore | null>(null)

function ordenarJogos(jogos: JogoEmAndamento[]) {
  return [...jogos].sort((a, b) => {
    const data = b.iniciado_em.slice(0, 10).localeCompare(a.iniciado_em.slice(0, 10))
    return data || b.id - a.id
  })
}

export function JogandoProvider({ children, token }: { children: ReactNode; token?: string }) {
  const auth = useContext(AuthContext)
  const effectiveToken = token ?? auth?.sessao?.access_token
  const [jogos, setJogos] = useState<JogoEmAndamento[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [carregado, setCarregado] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [aviso, setAviso] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)

  const carregar = useCallback(async (signal?: AbortSignal) => {
    setIsLoading(true)
    setError(null)
    try {
      const dados = await jogandoService.listar(effectiveToken, signal)
      if (!signal?.aborted) setJogos(dados)
    } catch (err) {
      if (!signal?.aborted) setError(err instanceof ApiError ? err.codigo : 'fallback')
    } finally {
      if (!signal?.aborted) {
        setIsLoading(false)
        setCarregado(true)
      }
    }
  }, [effectiveToken])

  const criar = useCallback(async (payload: CriarJogoEmAndamentoPayload) => {
    const criado = await jogandoService.criar(payload, effectiveToken)
    setJogos((atuais) => ordenarJogos([...atuais, criado]))
    return criado
  }, [effectiveToken])

  const remover = useCallback(async (id: number) => {
    try {
      await jogandoService.remover(id, effectiveToken)
      setJogos((atuais) => atuais.filter((jogo) => jogo.id !== id))
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        setJogos((atuais) => atuais.filter((jogo) => jogo.id !== id))
      }
      throw err
    }
  }, [effectiveToken])

  const value = useMemo<JogandoStore>(() => ({
    jogos, isLoading, carregado, error, aviso, isModalOpen,
    carregar, criar, remover,
    abrirModalIniciar: () => setIsModalOpen(true),
    fecharModal: () => setIsModalOpen(false),
    definirAviso: setAviso,
    limparAviso: () => setAviso(null),
  }), [jogos, isLoading, carregado, error, aviso, isModalOpen, carregar, criar, remover])

  return createElement(JogandoContext.Provider, { value }, children)
}

export function useJogandoStore(): JogandoStore {
  const store = useContext(JogandoContext)
  if (!store) throw new Error('useJogandoStore deve ser utilizado dentro de um JogandoProvider')
  return store
}
