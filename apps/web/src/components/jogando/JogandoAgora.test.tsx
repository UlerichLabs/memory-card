import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { JogandoAgora } from './JogandoAgora'

const mocks = vi.hoisted(() => ({
  dashboard: { carregarDashboard: vi.fn() },
  jogos: { abrirModalRegistro: vi.fn() },
  abandonados: { abrirModalCriacao: vi.fn() },
  jogando: {
    jogos: [{ id: 1, nome: 'Hades', igdb_id: 10, igdb_capa_url: null, iniciado_em: '2026-09-20T00:00:00Z' }],
    isLoading: false,
    carregado: true,
    error: null,
    aviso: null,
    carregar: vi.fn(),
    remover: vi.fn().mockResolvedValue(undefined),
    definirAviso: vi.fn(),
    abrirModalIniciar: vi.fn(),
  },
}))

vi.mock('@/stores/dashboardStore', () => ({ useDashboardStore: () => mocks.dashboard }))
vi.mock('@/stores/jogosStore', () => ({ useJogosStore: () => mocks.jogos }))
vi.mock('@/stores/abandonadosStore', () => ({ useAbandonadosStore: () => mocks.abandonados }))
vi.mock('@/stores/jogandoStore', () => ({ useJogandoStore: () => mocks.jogando }))

describe('JogandoAgora', () => {
  it('renderiza contador, card e prepara o zeramento', () => {
    render(<JogandoAgora />)
    expect(screen.getByRole('heading', { name: 'Jogando agora' })).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Zerei!' }))
    expect(mocks.jogos.abrirModalRegistro).toHaveBeenCalledWith(expect.objectContaining({ textoSubmit: 'Salvar zeramento' }))
  })

  it('mostra vazio e abre iniciar jogo', () => {
    mocks.jogando.jogos = []
    render(<JogandoAgora />)
    fireEvent.click(screen.getByRole('button', { name: 'Iniciar um jogo →' }))
    expect(mocks.jogando.abrirModalIniciar).toHaveBeenCalledOnce()
  })
})
