import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { DashboardProvider } from '@/stores/dashboardStore'
import { dashboardService } from '@/lib/services/dashboardService'
import { DashboardPage } from './DashboardPage'

vi.mock('@/lib/services/dashboardService', () => ({ dashboardService: {
  resumo: vi.fn(), abandonados: vi.fn(), jogoDoAno: vi.fn(), jogosDaVida: vi.fn(), recentes: vi.fn(), desafios: vi.fn(), porAno: vi.fn(), plataformas: vi.fn(), generos: vi.fn(), tipos: vi.fn(), notas: vi.fn(), dificuldade: vi.fn(), recordes: vi.fn(),
} }))

const services = vi.mocked(dashboardService)
const resumo = { total_jogos: 2, total_segundos: 10584000, media_segundos_por_jogo: 5292000, nota_media: 8.2, jogos_no_ano_atual: 1, primeiro_zeramento_em: '2020-01-01T00:00:00Z', dias_desde_primeiro: 1000, anos_desde_primeiro: 2 }
function preparar(jogos = 2) {
  services.resumo.mockResolvedValue({ ...resumo, total_jogos: jogos })
  services.abandonados.mockResolvedValue(3); services.jogoDoAno.mockResolvedValue([]); services.jogosDaVida.mockResolvedValue([]); services.recentes.mockResolvedValue([]); services.desafios.mockResolvedValue([]); services.porAno.mockResolvedValue([]); services.plataformas.mockResolvedValue([]); services.generos.mockResolvedValue([]); services.tipos.mockResolvedValue([]); services.notas.mockResolvedValue({ histograma: Array.from({ length: 11 }, (_, index) => ({ nota: index + 1, total: 0 })), nota_media: 0, total_avaliados: 0 }); services.dificuldade.mockResolvedValue([]); services.recordes.mockResolvedValue({ mais_longo: null, mais_curto: null })
}
function renderPage() { return render(<MemoryRouter><AuthProvider><JogosProvider><DashboardProvider><DashboardPage /></DashboardProvider></JogosProvider></AuthProvider></MemoryRouter>) }

describe('DashboardPage', () => {
  afterEach(() => vi.resetAllMocks())
  it('carrega os dados reais pelos services e mantém a navegação', async () => { preparar(); renderPage(); expect(await screen.findByText('Jogos zerados')).toBeInTheDocument(); expect(screen.getByRole('navigation', { name: 'Navegação Principal' })).toBeInTheDocument(); expect(services.resumo).toHaveBeenCalled() })
  it('mostra o estado vazio e o CTA quando não há jogos', async () => { preparar(0); renderPage(); expect(await screen.findByText('Seu dashboard começa no primeiro jogo zerado')).toBeInTheDocument(); expect(screen.getAllByRole('button', { name: '+ Registrar jogo' }).length).toBeGreaterThan(0) })
})
