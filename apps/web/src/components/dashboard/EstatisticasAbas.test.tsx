import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { EstatisticasAbas } from './EstatisticasAbas'
import type { DashboardStore } from '@/stores/dashboardStore'

const storeBase = {
  resumo: null, abandonados: 0, jogoDoAno: null, jogosDaVida: [], jogosDaVidaTotal: 0, recentes: [], desafios: [], porAno: [], plataformas: [{ console: 'PC', total_jogos: 2, total_segundos: 7200, percentual_jogos: 100, percentual_segundos: 100 }], generos: [{ genero: 'RPG', total_jogos: 2, total_segundos: 7200, percentual_jogos: 100, percentual_segundos: 100 }], tipos: [], notas: { histograma: [{ nota: 8, total: 3 }, { nota: 10, total: 1 }], nota_media: 8.5, total_avaliados: 4 }, dificuldade: [], recordes: { mais_longo: null, mais_curto: null }, loading: { resumo: false, abandonados: false, jogoDoAno: false, jogosDaVida: false, recentes: false, desafios: false, porAno: false, plataformas: false, generos: false, tipos: false, notas: false, dificuldade: false, recordes: false }, errors: {}, carregado: true, vazio: false, carregarDashboard: vi.fn(), recarregarBloco: vi.fn(),
} as unknown as DashboardStore

describe('EstatisticasAbas', () => {
  it('renderiza seis abas, troca por clique e calcula a moda', () => {
    render(<EstatisticasAbas dashboard={storeBase} />)
    expect(screen.getAllByRole('tab')).toHaveLength(6)
    fireEvent.click(screen.getByRole('tab', { name: 'Notas' }))
    expect(screen.getByText('Nota mais comum')).toBeInTheDocument()
    expect(screen.getByText('8')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Horas' })).not.toBeInTheDocument()
  })

  it('troca por teclado e exibe o toggle apenas nas abas compatíveis', () => {
    render(<EstatisticasAbas dashboard={storeBase} />)
    const porAno = screen.getByRole('tab', { name: 'Por Ano' })
    porAno.focus()
    fireEvent.keyDown(porAno, { key: 'ArrowRight' })
    expect(screen.getByRole('tab', { name: 'Notas' })).toHaveAttribute('aria-selected', 'true')
    fireEvent.click(screen.getByRole('tab', { name: 'Plataformas' }))
    expect(screen.getByRole('button', { name: 'Horas' })).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Horas' }))
    expect(screen.getByRole('button', { name: 'Horas' })).toHaveClass('bg-[var(--accent)]')
  })

  it('mantém outras abas disponíveis quando uma delas falha', () => {
    const dashboard = { ...storeBase, errors: { notas: 'dashboard.falha' } } as DashboardStore
    render(<EstatisticasAbas dashboard={dashboard} />)
    fireEvent.click(screen.getByRole('tab', { name: 'Notas' }))
    expect(screen.getByRole('alert')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('tab', { name: 'Plataformas' }))
    expect(screen.getByText('PC')).toBeInTheDocument()
  })
})
