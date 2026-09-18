import type { CSSProperties } from 'react'
import { useAuthStore } from '@/store/authStore'
import { mockDashboardData } from '@/mocks/dashboardData'
import { Topbar } from '@/components/layout/Topbar'
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

      <Topbar />

      <main className="mx-auto max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
        <div>
          <p className="text-[14px] text-[var(--text-secondary)]">
            Bem-vindo de volta, <strong className="font-bold text-[var(--text-primary)]">{nomeUsuario}</strong>
          </p>
        </div>

        <StatsRow estatisticas={mockDashboardData.estatisticas} />
        <GameOfTheYearCard jogoDoAno={mockDashboardData.jogoDoAno} />
        <LifeGamesGrid jogos={mockDashboardData.jogosDaVida} />

        <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
          <div className="space-y-8 lg:col-span-7 xl:col-span-8">
            <RecentlyCompleted jogos={mockDashboardData.zeradosRecentemente} />
            <ActivityHeatmap atividade={mockDashboardData.atividadeAno} />
          </div>
          <div className="space-y-8 lg:col-span-5 xl:col-span-4 [&_section[aria-label='Desafios Ativos']_>_div]:lg:grid-cols-1">
            <ActiveChallenges desafios={mockDashboardData.desafiosAtivos} />
            <PlatformBreakdown plataformas={mockDashboardData.distribuicaoPlataformas} />
            <TopGenres generos={mockDashboardData.principaisGeneros} />
          </div>
        </div>
      </main>
    </div>
  )
}
