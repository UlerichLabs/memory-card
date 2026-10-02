import { useEffect, type ReactNode } from 'react'
import { useAuthStore } from '@/store/authStore'
import { useJogosStore } from '@/stores/jogosStore'
import { useDashboardStore } from '@/stores/dashboardStore'
import { Topbar } from '@/components/layout/Topbar'
import { Skeleton } from '@/components/ui/skeleton'
import { PerfilJogador } from '@/components/dashboard/PerfilJogador'
import { EliteDoJogador } from '@/components/dashboard/EliteDoJogador'
import { ActiveChallenges } from '@/components/dashboard/ActiveChallenges'
import { RecentlyCompleted } from '@/components/dashboard/RecentlyCompleted'
import { EstatisticasAbas } from '@/components/dashboard/EstatisticasAbas'
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
  return <div className="dashboard-page min-h-full overflow-x-hidden bg-[var(--bg-primary)] text-[var(--text-primary)]"><Topbar /><main className="mx-auto max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
    {dashboard.carregado && dashboard.vazio ? <><JogandoAgora /><DashboardVazio onRegistrar={registrar} /></> : <>
    <Bloco erro={erroTotais} carregando={dashboard.loading.resumo || dashboard.loading.abandonados} pronto={Boolean(dashboard.resumo)} retry={retryTotais}>{dashboard.resumo && <PerfilJogador nome={nomeUsuario} resumo={dashboard.resumo} totalAbandonados={dashboard.abandonados} />}</Bloco>
      <Bloco erro={dashboard.errors.jogoDoAno || dashboard.errors.jogosDaVida} carregando={dashboard.loading.jogoDoAno || dashboard.loading.jogosDaVida} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('jogoDoAno')}><EliteDoJogador jogoDoAno={dashboard.jogoDoAno} jogosDaVida={dashboard.jogosDaVida} jogosDaVidaTotal={dashboard.jogosDaVidaTotal} /></Bloco>
      <JogandoAgora />
      <EstatisticasAbas dashboard={dashboard} />
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,1fr)_384px]"><Bloco erro={dashboard.errors.recentes} carregando={dashboard.loading.recentes} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('recentes')}><RecentlyCompleted jogos={dashboard.recentes} /></Bloco><Bloco erro={dashboard.errors.desafios} carregando={dashboard.loading.desafios} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('desafios')}><ActiveChallenges desafios={dashboard.desafios} /></Bloco></div>
    </>}</main></div>
}
