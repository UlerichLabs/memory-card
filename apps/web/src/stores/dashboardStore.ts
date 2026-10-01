import { createContext, createElement, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { AuthContext } from '@/store/authStore'
import { dashboardService } from '@/lib/services/dashboardService'
import type { DashboardAno, DashboardBloco, DashboardDados, DashboardDificuldade, DashboardNotas, DashboardRankingGenero, DashboardRankingPlataforma, DashboardRecordes, DashboardResumo, DashboardTipo } from '@/types/dashboard'
import type { JogoZeradoDTO } from '@/types/jogos'
import type { ListaResumo } from '@/types/listas'

type LoadingState = Record<DashboardBloco, boolean>
type ErrorState = Partial<Record<DashboardBloco, string>>

export interface DashboardStore extends DashboardDados {
  loading: LoadingState
  errors: ErrorState
  carregado: boolean
  vazio: boolean
  carregarDashboard: (signal?: AbortSignal) => Promise<void>
  recarregarBloco: (bloco: DashboardBloco) => Promise<void>
}

export const DashboardContext = createContext<DashboardStore | null>(null)

const blocos: DashboardBloco[] = ['resumo', 'abandonados', 'jogoDoAno', 'jogosDaVida', 'recentes', 'desafios', 'porAno', 'plataformas', 'generos', 'tipos', 'notas', 'dificuldade', 'recordes']
const vazioLoading = (): LoadingState => Object.fromEntries(blocos.map((bloco) => [bloco, false])) as LoadingState

export function DashboardProvider({ children, token }: { children: ReactNode; token?: string }) {
  const auth = useContext(AuthContext)
  const effectiveToken = token ?? auth?.sessao?.access_token
  const [resumo, setResumo] = useState<DashboardResumo | null>(null)
  const [abandonados, setAbandonados] = useState(0)
  const [jogoDoAno, setJogoDoAno] = useState<DashboardDados['jogoDoAno']>(null)
  const [jogosDaVida, setJogosDaVida] = useState<JogoZeradoDTO[]>([])
  const [recentes, setRecentes] = useState<JogoZeradoDTO[]>([])
  const [desafios, setDesafios] = useState<ListaResumo[]>([])
  const [porAno, setPorAno] = useState<DashboardAno[]>([])
  const [plataformas, setPlataformas] = useState<DashboardRankingPlataforma[]>([])
  const [generos, setGeneros] = useState<DashboardRankingGenero[]>([])
  const [tipos, setTipos] = useState<DashboardTipo[]>([])
  const [notas, setNotas] = useState<DashboardNotas | null>(null)
  const [dificuldade, setDificuldade] = useState<DashboardDificuldade[]>([])
  const [recordes, setRecordes] = useState<DashboardRecordes | null>(null)
  const [loading, setLoading] = useState<LoadingState>(vazioLoading)
  const [errors, setErrors] = useState<ErrorState>({})
  const [carregado, setCarregado] = useState(false)

  const setBloco = useCallback((bloco: DashboardBloco, active: boolean) => setLoading((atual) => ({ ...atual, [bloco]: active })), [])
  const setErro = useCallback((bloco: DashboardBloco, error: unknown) => setErrors((atual) => ({ ...atual, [bloco]: error instanceof Error ? error.message : 'Não foi possível carregar este bloco.' })), [])
  const limparErro = useCallback((bloco: DashboardBloco) => setErrors((atual) => { const proximo = { ...atual }; delete proximo[bloco]; return proximo }), [])

  const carregarBloco = useCallback(async (bloco: DashboardBloco, signal?: AbortSignal) => {
    setBloco(bloco, true)
    limparErro(bloco)
    try {
      if (bloco === 'resumo') setResumo(await dashboardService.resumo(effectiveToken, signal))
      if (bloco === 'abandonados') setAbandonados(await dashboardService.abandonados(effectiveToken, signal))
      if (bloco === 'jogoDoAno') {
        const anos = await dashboardService.jogoDoAno(effectiveToken, signal)
        const atual = [...anos].reverse().find((item) => item.game_do_ano)
        setJogoDoAno(atual?.game_do_ano ? { id: atual.game_do_ano.id, nome: atual.game_do_ano.nome, console: atual.game_do_ano.console, igdb_capa_url: atual.game_do_ano.igdb_capa_url ?? '', nota: atual.game_do_ano.nota, ano: atual.ano } : null)
      }
      if (bloco === 'jogosDaVida') setJogosDaVida(await dashboardService.jogosDaVida(effectiveToken, signal))
      if (bloco === 'recentes') setRecentes(await dashboardService.recentes(effectiveToken, signal))
      if (bloco === 'desafios') setDesafios(await dashboardService.desafios(effectiveToken, signal))
      if (bloco === 'porAno') setPorAno(await dashboardService.porAno(effectiveToken, signal))
      if (bloco === 'plataformas') setPlataformas(await dashboardService.plataformas(effectiveToken, signal))
      if (bloco === 'generos') {
        const dados = await dashboardService.generos(effectiveToken, signal)
        setGeneros(dados)
        if (dados[0]) setTipos(await dashboardService.tipos(dados[0].genero, effectiveToken, signal))
      }
      if (bloco === 'notas') setNotas(await dashboardService.notas(effectiveToken, signal))
      if (bloco === 'dificuldade') setDificuldade(await dashboardService.dificuldade(effectiveToken, signal))
      if (bloco === 'recordes') setRecordes(await dashboardService.recordes(effectiveToken, signal))
    } catch (error) {
      if (!signal?.aborted) setErro(bloco, error)
    } finally {
      if (!signal?.aborted) setBloco(bloco, false)
    }
  }, [effectiveToken, limparErro, setBloco, setErro])

  const carregarDashboard = useCallback(async (signal?: AbortSignal) => {
    setCarregado(false)
    await Promise.allSettled(blocos.map((bloco) => carregarBloco(bloco, signal)))
    if (!signal?.aborted) setCarregado(true)
  }, [carregarBloco])

  const recarregarBloco = useCallback((bloco: DashboardBloco) => carregarBloco(bloco), [carregarBloco])
  const value = useMemo<DashboardStore>(() => ({ resumo, abandonados, jogoDoAno, jogosDaVida, recentes, desafios, porAno, plataformas, generos, tipos, notas, dificuldade, recordes, loading, errors, carregado, vazio: resumo?.total_jogos === 0, carregarDashboard, recarregarBloco }), [resumo, abandonados, jogoDoAno, jogosDaVida, recentes, desafios, porAno, plataformas, generos, tipos, notas, dificuldade, recordes, loading, errors, carregado, carregarDashboard, recarregarBloco])
  return createElement(DashboardContext.Provider, { value }, children)
}

export function useDashboardStore(): DashboardStore {
  const store = useContext(DashboardContext)
  if (!store) throw new Error('useDashboardStore deve ser utilizado dentro de um DashboardProvider')
  return store
}
