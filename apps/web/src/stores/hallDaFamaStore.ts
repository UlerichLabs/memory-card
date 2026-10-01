import { createContext, createElement, useCallback, useContext, useState, type ReactNode } from 'react'
import {
  jogosService,
  type JogoZeradoDTO,
  type ResumoGameDoAnoItem,
  type DefinirGameDoAnoResposta,
} from '@/lib/services/jogosService'
import { AuthContext } from '@/store/authStore'

export interface HallDaFamaStore {
  resumo: ResumoGameDoAnoItem[]
  gamesDaVida: JogoZeradoDTO[]
  isLoading: boolean
  error: string | null
  carregarHallDaFama: (signal?: AbortSignal) => Promise<void>
  buscarJogosDoAno: (ano: number, signal?: AbortSignal) => Promise<JogoZeradoDTO[]>
  definirGameDoAno: (id: number) => Promise<DefinirGameDoAnoResposta>
  removerGameDoAno: (id: number) => Promise<void>
  limparErro: () => void
}

export const HallDaFamaContext = createContext<HallDaFamaStore | null>(null)

export interface HallDaFamaProviderProps {
  children: ReactNode
  token?: string
  initialResumo?: ResumoGameDoAnoItem[]
  initialGamesDaVida?: JogoZeradoDTO[]
}

export function HallDaFamaProvider({
  children,
  token,
  initialResumo = [],
  initialGamesDaVida = [],
}: HallDaFamaProviderProps) {
  const auth = useContext(AuthContext)
  const effectiveToken = token ?? auth?.sessao?.access_token
  const [resumo, setResumo] = useState<ResumoGameDoAnoItem[]>(initialResumo)
  const [gamesDaVida, setGamesDaVida] = useState<JogoZeradoDTO[]>(initialGamesDaVida)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const carregarHallDaFama = useCallback(async (signal?: AbortSignal) => {
    setIsLoading(true)
    setError(null)
    try {
      const [resumoResp, vidaResp] = await Promise.all([
        jogosService.obterResumoGameDoAno(effectiveToken, signal),
        jogosService.listar({ nota_min: 11, por_pagina: 100 }, effectiveToken, signal),
      ])
      if (!signal?.aborted) {
        setResumo(resumoResp)
        setGamesDaVida(vidaResp.data)
      }
    } catch (err) {
      if (!signal?.aborted) {
        setError(err instanceof Error ? err.message : 'Erro ao carregar Hall da Fama')
        throw err
      }
    } finally {
      if (!signal?.aborted) setIsLoading(false)
    }
  }, [effectiveToken])

  const buscarJogosDoAno = useCallback(async (ano: number, signal?: AbortSignal): Promise<JogoZeradoDTO[]> => {
    const res = await jogosService.listar(
      { ano, ordenar: 'nota', por_pagina: 100 },
      effectiveToken,
      signal
    )
    return res.data
  }, [effectiveToken])

  const definirGameDoAno = useCallback(async (id: number): Promise<DefinirGameDoAnoResposta> => {
    const resp = await jogosService.definirGameDoAno(id, effectiveToken)
    await carregarHallDaFama()
    return resp
  }, [effectiveToken, carregarHallDaFama])

  const removerGameDoAno = useCallback(async (id: number): Promise<void> => {
    await jogosService.removerGameDoAno(id, effectiveToken)
    await carregarHallDaFama()
  }, [effectiveToken, carregarHallDaFama])

  const store: HallDaFamaStore = {
    resumo,
    gamesDaVida,
    isLoading,
    error,
    carregarHallDaFama,
    buscarJogosDoAno,
    definirGameDoAno,
    removerGameDoAno,
    limparErro: () => setError(null),
  }

  return createElement(HallDaFamaContext.Provider, { value: store }, children)
}

export function useHallDaFamaStore(): HallDaFamaStore {
  const store = useContext(HallDaFamaContext)
  if (!store) throw new Error('useHallDaFamaStore deve ser utilizado dentro de um HallDaFamaProvider')
  return store
}
