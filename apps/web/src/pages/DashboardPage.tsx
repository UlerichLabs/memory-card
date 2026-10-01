import { useEffect, type ReactNode } from 'react'
import { useAuthStore } from '@/store/authStore'
import { useJogosStore } from '@/stores/jogosStore'
import { useDashboardStore } from '@/stores/dashboardStore'
import { Topbar } from '@/components/layout/Topbar'
import { Skeleton } from '@/components/ui/skeleton'
import { StatsRow } from '@/components/dashboard/StatsRow'
import { GameOfTheYearCard } from '@/components/dashboard/GameOfTheYearCard'
import { LifeGamesGrid } from '@/components/dashboard/LifeGamesGrid'
import { ActiveChallenges } from '@/components/dashboard/ActiveChallenges'
import { RecentlyCompleted } from '@/components/dashboard/RecentlyCompleted'
import { PlatformBreakdown } from '@/components/dashboard/PlatformBreakdown'
import { TopGenres } from '@/components/dashboard/TopGenres'
import { PorAnoChart } from '@/components/dashboard/PorAnoChart'
import { NotasChart } from '@/components/dashboard/NotasChart'
import { DificuldadeBreakdown } from '@/components/dashboard/DificuldadeBreakdown'
import { Recordes } from '@/components/dashboard/Recordes'
import { DashboardVazio } from '@/components/dashboard/DashboardVazio'
import { DashboardSecaoErro } from '@/components/dashboard/DashboardSecaoErro'
import { JogandoAgora } from '@/components/jogando/JogandoAgora'
import { useJogandoStore } from '@/stores/jogandoStore'

function Bloco({ erro, carregando, pronto, retry, children }: { erro?: string; carregando: boolean; pronto: boolean; retry: () => void; children: ReactNode }) {
  if (erro) return <DashboardSecaoErro onRetry={retry} />
  if (carregando || !pronto) return <Skeleton className="h-32 w-full bg-[var(--bg-surface)]" />
  return <>{children}</>
}

export function DashboardPage() {
  const { sessao } = useAuthStore()
  const { abrirModalRegistro } = useJogosStore()
  const dashboard = useDashboardStore()
  const { carregar: carregarJogando } = useJogandoStore()
  const { carregarDashboard } = dashboard
  const nomeUsuario = sessao?.usuario?.nome ?? 'Jogador'
  useEffect(() => { document.title = 'Dashboard'; const controller = new AbortController(); void carregarDashboard(controller.signal); void carregarJogando(controller.signal); return () => { controller.abort(); document.title = 'Memory Card' } }, [carregarDashboard, carregarJogando])

  const registrar = () => abrirModalRegistro({ onSalvo: () => dashboard.carregarDashboard() })
  const erroTotais = dashboard.errors.resumo ?? dashboard.errors.abandonados
  const retryTotais = dashboard.errors.resumo ? () => void dashboard.recarregarBloco('resumo') : () => void dashboard.recarregarBloco('abandonados')
  return <div className="dashboard-page min-h-svh overflow-x-hidden bg-[var(--bg-primary)] text-[var(--text-primary)]"><Topbar /><main className="mx-auto max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8"><p className="text-sm text-[var(--text-secondary)]">Bem-vindo de volta, <strong className="font-bold text-[var(--text-primary)]">{nomeUsuario}</strong></p>
    {dashboard.carregado && dashboard.vazio ? <><JogandoAgora /><DashboardVazio onRegistrar={registrar} /></> : <>
    <Bloco erro={erroTotais} carregando={dashboard.loading.resumo || dashboard.loading.abandonados} pronto={Boolean(dashboard.resumo)} retry={retryTotais}>{dashboard.resumo && <StatsRow resumo={dashboard.resumo} totalAbandonados={dashboard.abandonados} />}</Bloco>
      <JogandoAgora />
      {(dashboard.jogoDoAno || dashboard.errors.jogoDoAno || dashboard.loading.jogoDoAno || !dashboard.carregado) && <Bloco erro={dashboard.errors.jogoDoAno} carregando={dashboard.loading.jogoDoAno} pronto={Boolean(dashboard.jogoDoAno)} retry={() => void dashboard.recarregarBloco('jogoDoAno')}>{dashboard.jogoDoAno && <GameOfTheYearCard jogo={dashboard.jogoDoAno} />}</Bloco>}
      <Bloco erro={dashboard.errors.jogosDaVida} carregando={dashboard.loading.jogosDaVida} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('jogosDaVida')}><LifeGamesGrid jogos={dashboard.jogosDaVida} /></Bloco>
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-12"><div className="space-y-8 lg:col-span-7 xl:col-span-8">
        <Bloco erro={dashboard.errors.recentes} carregando={dashboard.loading.recentes} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('recentes')}><RecentlyCompleted jogos={dashboard.recentes} /></Bloco>
        <Bloco erro={dashboard.errors.porAno} carregando={dashboard.loading.porAno} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('porAno')}><PorAnoChart anos={dashboard.porAno} /></Bloco>
      </div><div className="space-y-8 lg:col-span-5 xl:col-span-4">
        <Bloco erro={dashboard.errors.desafios} carregando={dashboard.loading.desafios} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('desafios')}><ActiveChallenges desafios={dashboard.desafios} /></Bloco>
        <Bloco erro={dashboard.errors.plataformas} carregando={dashboard.loading.plataformas} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('plataformas')}><PlatformBreakdown plataformas={dashboard.plataformas} /></Bloco>
        <Bloco erro={dashboard.errors.generos} carregando={dashboard.loading.generos} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('generos')}><TopGenres generos={dashboard.generos} tipos={dashboard.tipos} /></Bloco>
      </div></div>
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-3"><Bloco erro={dashboard.errors.notas} carregando={dashboard.loading.notas} pronto={Boolean(dashboard.notas)} retry={() => void dashboard.recarregarBloco('notas')}>{dashboard.notas && <NotasChart notas={dashboard.notas} />}</Bloco><Bloco erro={dashboard.errors.dificuldade} carregando={dashboard.loading.dificuldade} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('dificuldade')}><DificuldadeBreakdown dificuldade={dashboard.dificuldade} /></Bloco><Bloco erro={dashboard.errors.recordes} carregando={dashboard.loading.recordes} pronto={Boolean(dashboard.recordes)} retry={() => void dashboard.recarregarBloco('recordes')}>{dashboard.recordes && <Recordes recordes={dashboard.recordes} />}</Bloco></div>
    </>}</main></div>
}
