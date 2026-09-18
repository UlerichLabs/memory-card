import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { Disc3, User } from 'lucide-react'
import { useAuthStore } from '@/store/authStore'
import { mockDashboardData } from '@/mocks/dashboardData'
import { StatsRow } from '@/components/dashboard/StatsRow'
import { GameOfTheYearCard } from '@/components/dashboard/GameOfTheYearCard'
import { LifeGamesGrid } from '@/components/dashboard/LifeGamesGrid'
import { ActiveChallenges } from '@/components/dashboard/ActiveChallenges'
import { ActivityHeatmap } from '@/components/dashboard/ActivityHeatmap'
import { RecentlyCompleted } from '@/components/dashboard/RecentlyCompleted'
import { PlatformBreakdown } from '@/components/dashboard/PlatformBreakdown'
import { TopGenres } from '@/components/dashboard/TopGenres'

const dashboardTheme = {
  '--bg-primary': '#15161A',
  '--bg-surface': '#1A1B20',
  '--bg-surface-alt': '#1D1F25',
  '--border': '#24262C',
  '--border-subtle': '#2A2C33',
  '--text-primary': '#EDEDED',
  '--text-secondary': '#9A9CA5',
  '--text-muted': '#6B6D76',
  '--text-faint': '#52545C',
  '--accent': '#4F7CFF',
  '--highlight-gold': '#E8C15C',
  '--background': '#15161A',
  '--foreground': '#EDEDED',
  '--card': '#1A1B20',
  '--card-foreground': '#EDEDED',
  '--secondary': '#1D1F25',
  '--secondary-foreground': '#9A9CA5',
  '--muted': '#1D1F25',
  '--muted-foreground': '#6B6D76',
  '--primary': '#4F7CFF',
  '--primary-foreground': '#0E0F12',
  fontFamily: 'Inter, sans-serif',
} as CSSProperties

export function DashboardPage() {
  const { sessao } = useAuthStore()
  const nomeUsuario = sessao?.usuario?.nome ?? 'Jogador'

  return (
    <div style={dashboardTheme} className="min-h-svh bg-[var(--bg-primary)] text-[var(--text-primary)]">
      <link
        rel="stylesheet"
        href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap"
      />

      <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--bg-surface)]/95 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-8">
            <Link to="/" className="flex items-center gap-2 font-bold tracking-tight text-[var(--text-primary)]">
              <Disc3 className="h-5 w-5 text-[var(--accent)]" aria-hidden="true" />
              <span>Memory Card</span>
            </Link>

            <nav aria-label="Navegação Principal" className="flex items-center gap-6 text-[13px]">
              <Link
                to="/"
                className="border-b-2 border-[var(--accent)] py-4 font-semibold text-[var(--accent)]"
              >
                Dashboard
              </Link>
              <Link
                to="/biblioteca"
                className="py-4 font-medium text-[var(--text-secondary)] transition-colors hover:text-[var(--text-primary)]"
              >
                Biblioteca
              </Link>
            </nav>
          </div>

          <div className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)]">
              <User className="h-3.5 w-3.5 text-[var(--text-secondary)]" aria-hidden="true" />
            </div>
            <span className="hidden text-[13px] font-medium text-[var(--text-secondary)] sm:inline">
              {nomeUsuario}
            </span>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
        <StatsRow estatisticas={mockDashboardData.estatisticas} />
        <GameOfTheYearCard jogoDoAno={mockDashboardData.jogoDoAno} />
        <LifeGamesGrid jogos={mockDashboardData.jogosDaVida} />
        <ActiveChallenges desafios={mockDashboardData.desafiosAtivos} />
        <ActivityHeatmap atividade={mockDashboardData.atividadeAno} />
        <RecentlyCompleted jogos={mockDashboardData.zeradosRecentemente} />
        <div className="grid grid-cols-1 gap-4 sm:gap-6 md:grid-cols-2">
          <PlatformBreakdown plataformas={mockDashboardData.distribuicaoPlataformas} />
          <TopGenres generos={mockDashboardData.principaisGeneros} />
        </div>
      </main>
    </div>
  )
}
