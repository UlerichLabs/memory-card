import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { PlatformBreakdown } from './PlatformBreakdown'
import { TopGenres } from './TopGenres'
import { NotasChart } from './NotasChart'
import { DificuldadeBreakdown } from './DificuldadeBreakdown'
import { Recordes } from './Recordes'
import { DashboardVazio } from './DashboardVazio'
import { DashboardSecaoErro } from './DashboardSecaoErro'
import { GameOfTheYearCard } from './GameOfTheYearCard'
import { PorAnoChart } from './PorAnoChart'

const plataforma = [{ console: 'PC', total_jogos: 2, total_segundos: 7200, percentual_jogos: 66.7, percentual_segundos: 50 }, { console: 'PS5', total_jogos: 1, total_segundos: 7200, percentual_jogos: 33.3, percentual_segundos: 50 }]
const genero = [{ genero: 'RPG', total_jogos: 2, total_segundos: 7200, percentual_jogos: 66.7, percentual_segundos: 50 }, { genero: 'Ação', total_jogos: 1, total_segundos: 7200, percentual_jogos: 33.3, percentual_segundos: 50 }]

describe('componentes do dashboard', () => {
  it('alterna Plataforma e Gêneros entre jogos e horas, incluindo chips', () => { render(<><PlatformBreakdown plataformas={plataforma} /><TopGenres generos={genero} tipos={[{ tipo: 'Principal', total_jogos: 2 }]} /></>); expect(screen.getAllByText('2 · 66.7%')).toHaveLength(2); fireEvent.click(screen.getAllByRole('button', { name: 'Horas' })[0]); expect(screen.getAllByText('2h · 50%')).toHaveLength(2); expect(screen.getByText('Principal 2')).toBeInTheDocument() })
  it('mantém histograma e dificuldade completos e não quebra com recordes nulos', () => { render(<><NotasChart notas={{ histograma: Array.from({ length: 11 }, (_, index) => ({ nota: index + 1, total: 0 })), nota_media: 8.2, total_avaliados: 0 }} /><DificuldadeBreakdown dificuldade={[]} /><Recordes recordes={{ mais_longo: null, mais_curto: null }} /></>); expect(screen.getByText('Média 8.2')).toBeInTheDocument(); expect(screen.getAllByText(/jogos ·/)).toHaveLength(5); expect(screen.getAllByText('Nenhum registro elegível.')).toHaveLength(2) })
  it('executa CTA vazio e retry de seção', () => { const registrar = vi.fn(); const retry = vi.fn(); render(<><DashboardVazio onRegistrar={registrar} /><DashboardSecaoErro onRetry={retry} /></>); fireEvent.click(screen.getByRole('button', { name: '+ Registrar jogo' })); fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' })); expect(registrar).toHaveBeenCalledOnce(); expect(retry).toHaveBeenCalledOnce() })
  it('exibe tempo no Jogo do Ano e alterna Por Ano', () => { render(<><GameOfTheYearCard jogo={{ id: 1, nome: 'Jogo', console: 'PC', igdb_capa_url: '', nota: 11, ano: 2026, tempo_jogado: 151200 }} /><PorAnoChart anos={[{ ano: 2025, total_jogos: 2, total_segundos: 7200, game_do_ano: null }, { ano: 2026, total_jogos: 1, total_segundos: 14400, game_do_ano: { id: 1, nome: 'Jogo', console: 'PC', igdb_capa_url: '', nota: 11 } }]} /></>); expect(screen.getByText('42h jogadas')).toBeInTheDocument(); expect(screen.getByText(/sem destaque/)).toBeInTheDocument(); expect(screen.queryByText('2026: Jogo')).not.toBeInTheDocument(); fireEvent.click(screen.getByRole('button', { name: 'Horas' })); expect(screen.getByText('Horas')).toBeInTheDocument() })
  it('renderiza estatísticas embutidas sem card ou cabeçalho próprio', () => { const { container } = render(<><PorAnoChart anos={[{ ano: 2025, total_jogos: 2, total_segundos: 7200, game_do_ano: null }]} embutido /><NotasChart notas={{ histograma: [{ nota: 1, total: 1 }], nota_media: 1, total_avaliados: 1 }} embutido /><DificuldadeBreakdown dificuldade={[]} embutido /><PlatformBreakdown plataformas={plataforma} embutido /><TopGenres generos={genero} embutido tipos={[]} /><Recordes recordes={{ mais_longo: null, mais_curto: null }} compact embutido /></>); expect(container.querySelectorAll('section')).toHaveLength(0); expect(screen.queryByRole('heading', { name: 'Por Ano' })).not.toBeInTheDocument(); expect(container.querySelector('.lg\\:grid-cols-2')).toBeInTheDocument() })
})
