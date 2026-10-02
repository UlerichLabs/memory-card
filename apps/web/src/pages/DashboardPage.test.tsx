import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { AbandonadosProvider } from '@/stores/abandonadosStore'
import { JogandoProvider } from '@/stores/jogandoStore'
import { DashboardProvider } from '@/stores/dashboardStore'
import { dashboardService } from '@/lib/services/dashboardService'
import { jogandoService } from '@/lib/services/jogandoService'
import { DashboardPage } from './DashboardPage'

vi.mock('@/lib/services/dashboardService', () => ({ dashboardService: {
  resumo: vi.fn(), abandonados: vi.fn(), jogoDoAno: vi.fn(), jogosDaVida: vi.fn(), recentes: vi.fn(), desafios: vi.fn(), porAno: vi.fn(), plataformas: vi.fn(), generos: vi.fn(), tipos: vi.fn(), notas: vi.fn(), dificuldade: vi.fn(), recordes: vi.fn(),
} }))
vi.mock('@/lib/services/jogandoService', () => ({ jogandoService: { listar: vi.fn(), criar: vi.fn(), remover: vi.fn() } }))

const services = vi.mocked(dashboardService)
const jogandoServices = vi.mocked(jogandoService)
const resumo = { total_jogos: 2, total_segundos: 10584000, media_segundos_por_jogo: 5292000, nota_media: 8.2, jogos_no_ano_atual: 1, primeiro_zeramento_em: '2020-01-01T00:00:00Z', dias_desde_primeiro: 1000, anos_desde_primeiro: 2 }
function preparar(jogos = 2) {
  services.resumo.mockResolvedValue({ ...resumo, total_jogos: jogos })
  services.abandonados.mockResolvedValue(3); services.jogoDoAno.mockResolvedValue([]); services.jogosDaVida.mockResolvedValue({ jogos: [], total: 0 }); services.recentes.mockResolvedValue([]); services.desafios.mockResolvedValue([]); services.porAno.mockResolvedValue([]); services.plataformas.mockResolvedValue([]); services.generos.mockResolvedValue([]); services.tipos.mockResolvedValue([]); services.notas.mockResolvedValue({ histograma: Array.from({ length: 11 }, (_, index) => ({ nota: index + 1, total: 0 })), nota_media: 0, total_avaliados: 0 }); services.dificuldade.mockResolvedValue([]); services.recordes.mockResolvedValue({ mais_longo: null, mais_curto: null })
  jogandoServices.listar.mockResolvedValue([])
}
function renderPage() { return render(<MemoryRouter><AuthProvider><JogosProvider><AbandonadosProvider><JogandoProvider><DashboardProvider><DashboardPage /></DashboardProvider></JogandoProvider></AbandonadosProvider></JogosProvider></AuthProvider></MemoryRouter>) }

describe('DashboardPage', () => {
  afterEach(() => vi.resetAllMocks())
  it('carrega os dados reais pelos services e mantém a navegação', async () => { preparar(); renderPage(); expect(await screen.findByText('Jogos zerados')).toBeInTheDocument(); expect(screen.getByRole('navigation', { name: 'Navegação Principal' })).toBeInTheDocument(); expect(services.resumo).toHaveBeenCalled() })
  it('mostra o estado vazio e o CTA quando não há jogos', async () => { preparar(0); renderPage(); expect(await screen.findByText('Seu dashboard começa no primeiro jogo zerado')).toBeInTheDocument(); expect(screen.getAllByRole('button', { name: '+ Registrar jogo' }).length).toBeGreaterThan(0) })
  it('posiciona perfil, elite e Jogando agora na ordem do dashboard', async () => {
    preparar()
    services.jogoDoAno.mockResolvedValue([{ ano: 2026, total_jogos: 1, game_do_ano: { id: 1, nome: 'Hades', console: 'PC', igdb_capa_url: '', nota: 10, usuario_id: 1, finalizado_em: '2026-01-01', tempo_jogado: 3600, dificuldade: 'A', destaque: true } }])
    renderPage()
    const totais = await screen.findByText('Jogos zerados')
    const jogando = screen.getByRole('heading', { name: 'Jogando agora' })
    const jogoDoAno = await screen.findByText('Jogo do Ano 2026')
    expect(totais.compareDocumentPosition(jogando) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(totais.compareDocumentPosition(jogoDoAno) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(jogoDoAno.compareDocumentPosition(jogando) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
  })
  it('mantém os outros blocos quando uma chamada falha e permite retry isolado', async () => {
    preparar()
    services.notas.mockRejectedValue(new Error('falha de notas'))
    renderPage()
    expect(await screen.findByText('Jogos zerados')).toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent('Não foi possível carregar este bloco.')
    services.notas.mockResolvedValue({ histograma: Array.from({ length: 11 }, (_, index) => ({ nota: index + 1, total: 0 })), nota_media: 0, total_avaliados: 0 })
    const chamadasResumo = services.resumo.mock.calls.length
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' })) })
    expect(await screen.findByText('Média 0.0')).toBeInTheDocument()
    expect(services.resumo).toHaveBeenCalledTimes(chamadasResumo)
  })

  it('não renderiza valores inválidos com dados mínimos', async () => {
    preparar()
    renderPage()
    await screen.findByText('Jogos zerados')
    expect(screen.queryByText(/NaN|undefined|null/)).not.toBeInTheDocument()
  })
})
