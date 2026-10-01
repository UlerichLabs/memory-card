import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { PlatformBreakdown } from './PlatformBreakdown'
import { TopGenres } from './TopGenres'
import { NotasChart } from './NotasChart'
import { DificuldadeBreakdown } from './DificuldadeBreakdown'
import { Recordes } from './Recordes'
import { DashboardVazio } from './DashboardVazio'
import { DashboardSecaoErro } from './DashboardSecaoErro'

const plataforma = [{ console: 'PC', total_jogos: 2, total_segundos: 7200, percentual_jogos: 66.7, percentual_segundos: 50 }, { console: 'PS5', total_jogos: 1, total_segundos: 7200, percentual_jogos: 33.3, percentual_segundos: 50 }]
const genero = [{ genero: 'RPG', total_jogos: 2, total_segundos: 7200, percentual_jogos: 66.7, percentual_segundos: 50 }, { genero: 'Ação', total_jogos: 1, total_segundos: 7200, percentual_jogos: 33.3, percentual_segundos: 50 }]

describe('componentes do dashboard', () => {
  it('alterna Plataforma e Gêneros entre jogos e horas, incluindo chips', () => { render(<><PlatformBreakdown plataformas={plataforma} /><TopGenres generos={genero} tipos={[{ tipo: 'Principal', total_jogos: 2 }]} /></>); expect(screen.getAllByText('2 · 66.7%')).toHaveLength(2); fireEvent.click(screen.getAllByRole('button', { name: 'Horas' })[0]); expect(screen.getAllByText('2h · 50%')).toHaveLength(2); expect(screen.getByText('Principal 2')).toBeInTheDocument() })
  it('mantém histograma e dificuldade completos e não quebra com recordes nulos', () => { render(<><NotasChart notas={{ histograma: Array.from({ length: 11 }, (_, index) => ({ nota: index + 1, total: 0 })), nota_media: 8.2, total_avaliados: 0 }} /><DificuldadeBreakdown dificuldade={[]} /><Recordes recordes={{ mais_longo: null, mais_curto: null }} /></>); expect(screen.getByText('Média 8.2')).toBeInTheDocument(); expect(screen.getAllByText(/jogos ·/)).toHaveLength(5); expect(screen.getAllByText('Nenhum registro elegível.')).toHaveLength(2) })
  it('executa CTA vazio e retry de seção', () => { const registrar = vi.fn(); const retry = vi.fn(); render(<><DashboardVazio onRegistrar={registrar} /><DashboardSecaoErro onRetry={retry} /></>); fireEvent.click(screen.getByRole('button', { name: '+ Registrar jogo' })); fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' })); expect(registrar).toHaveBeenCalledOnce(); expect(retry).toHaveBeenCalledOnce() })
})
