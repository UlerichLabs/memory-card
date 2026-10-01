import { act, renderHook, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { DashboardProvider, useDashboardStore } from './dashboardStore'
import { dashboardService } from '@/lib/services/dashboardService'
import { ApiError } from '@/lib/api'

vi.mock('@/lib/services/dashboardService')
const mocked = vi.mocked(dashboardService)
function preparar() {
  mocked.resumo.mockResolvedValue({ total_jogos: 1, total_segundos: 3600, media_segundos_por_jogo: 3600, nota_media: 8, jogos_no_ano_atual: 1, primeiro_zeramento_em: null, dias_desde_primeiro: 0, anos_desde_primeiro: 0 })
  mocked.abandonados.mockResolvedValue(0); mocked.jogoDoAno.mockResolvedValue([]); mocked.jogosDaVida.mockResolvedValue([]); mocked.recentes.mockResolvedValue([]); mocked.desafios.mockRejectedValue(new Error('falhou')); mocked.porAno.mockResolvedValue([]); mocked.plataformas.mockResolvedValue([]); mocked.generos.mockResolvedValue([]); mocked.tipos.mockResolvedValue([]); mocked.notas.mockResolvedValue({ histograma: [], nota_media: 0, total_avaliados: 0 }); mocked.dificuldade.mockResolvedValue([]); mocked.recordes.mockResolvedValue({ mais_longo: null, mais_curto: null })
}
describe('dashboardStore', () => {
  beforeEach(() => vi.resetAllMocks())

  it('carrega blocos em paralelo e preserva falha parcial', async () => { preparar(); mocked.desafios.mockRejectedValue(new ApiError('dashboard.falha', 'falhou')); const { result } = renderHook(() => useDashboardStore(), { wrapper: DashboardProvider }); await act(() => result.current.carregarDashboard()); await waitFor(() => expect(result.current.carregado).toBe(true)); expect(result.current.resumo?.total_jogos).toBe(1); expect(result.current.errors.desafios).toBe('dashboard.falha') })
  it('recarrega somente o bloco solicitado', async () => { preparar(); const { result } = renderHook(() => useDashboardStore(), { wrapper: DashboardProvider }); await act(() => result.current.carregarDashboard()); const chamadas = mocked.resumo.mock.calls.length; await act(() => result.current.recarregarBloco('resumo')); expect(mocked.resumo).toHaveBeenCalledTimes(chamadas + 1) })

  it.each([
    {
      nome: 'ordem decrescente da API',
      anos: [2026, 2025, 2009],
      destaques: [null, 'A', 'B'],
      esperado: 'A',
    },
    {
      nome: 'ano atual com destaque',
      anos: [2026, 2025],
      destaques: ['C', 'A'],
      esperado: 'C',
    },
    {
      nome: 'ordem crescente',
      anos: [2009, 2025, 2026],
      destaques: ['B', 'A', null],
      esperado: 'A',
    },
  ])('seleciona o destaque mais recente quando há jogos do ano em $nome', async ({ anos, destaques, esperado }) => {
    preparar()
    mocked.jogoDoAno.mockResolvedValue(anos.map((ano, index) => ({
      ano,
      total_jogos: 1,
      game_do_ano: destaques[index] ? { ...jogo, nome: destaques[index] } : null,
    })))
    const { result } = renderHook(() => useDashboardStore(), { wrapper: DashboardProvider })
    await act(() => result.current.carregarDashboard())
    await waitFor(() => expect(result.current.carregado).toBe(true))
    expect(result.current.jogoDoAno?.nome).toBe(esperado)
  })

  it('retorna null quando nenhum ano tem destaque', async () => {
    preparar()
    mocked.jogoDoAno.mockResolvedValue([
      { ano: 2026, total_jogos: 1, game_do_ano: null },
      { ano: 2025, total_jogos: 1, game_do_ano: null },
    ])
    const { result } = renderHook(() => useDashboardStore(), { wrapper: DashboardProvider })
    await act(() => result.current.carregarDashboard())
    await waitFor(() => expect(result.current.carregado).toBe(true))
    expect(result.current.jogoDoAno).toBeNull()
  })
})

const jogo = {
  id: 1,
  usuario_id: 1,
  nome: 'Game',
  console: 'PC',
  finalizado_em: '2026-01-01',
  tempo_jogado: 3600,
  nota: 10,
  dificuldade: 'A' as const,
  destaque: true,
}
