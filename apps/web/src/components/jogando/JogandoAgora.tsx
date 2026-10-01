import { Play } from 'lucide-react'
import { ApiError } from '@/lib/api'
import { Skeleton } from '@/components/ui/skeleton'
import { DashboardSecaoErro } from '@/components/dashboard/DashboardSecaoErro'
import { useDashboardStore } from '@/stores/dashboardStore'
import { useJogosStore } from '@/stores/jogosStore'
import { useAbandonadosStore } from '@/stores/abandonadosStore'
import { useJogandoStore } from '@/stores/jogandoStore'
import { formatarDataJogando } from '@/lib/jogandoUtils'
import { JogandoCard } from './JogandoCard'
import { JogandoVazio } from './JogandoVazio'

const AVISO_ZERAMENTO = (nome: string) => `Zeramento salvo, mas não consegui tirar «${nome}» de Jogando agora. Remova pelo menu …`
const AVISO_ABANDONO = (nome: string) => `Abandono salvo, mas não consegui tirar «${nome}» de Jogando agora. Remova pelo menu …`
const AVISO_REMOCAO = 'Não foi possível remover este jogo de Jogando agora. Tente novamente.'

export function JogandoAgora() {
  const dashboard = useDashboardStore()
  const jogos = useJogosStore()
  const abandonados = useAbandonadosStore()
  const jogando = useJogandoStore()

  async function retirar(jogoId: number, aviso: string) {
    try {
      await jogando.remover(jogoId)
    } catch (error) {
      if (!(error instanceof ApiError && error.status === 404)) jogando.definirAviso(aviso)
    }
  }

  function abrirZeramento(jogo: typeof jogando.jogos[number]) {
    jogos.abrirModalRegistro({
      valoresIniciais: { nome: jogo.nome, igdb_id: jogo.igdb_id, igdb_capa_url: jogo.igdb_capa_url ?? '', iniciado_em: jogo.iniciado_em },
      aviso: `Você começou este jogo em ${formatarDataJogando(jogo.iniciado_em)}. Ao salvar o zeramento, ele sai de Jogando agora.`,
      textoSubmit: 'Salvar zeramento',
      onSalvo: async () => { await retirar(jogo.id, AVISO_ZERAMENTO(jogo.nome)); await dashboard.carregarDashboard() },
    })
  }

  function abrirAbandono(jogo: typeof jogando.jogos[number]) {
    abandonados.abrirModalCriacao({
      valoresIniciais: { nome: jogo.nome, igdb_id: jogo.igdb_id, igdb_capa_url: jogo.igdb_capa_url, iniciado_em: jogo.iniciado_em },
      aviso: `Você começou este jogo em ${formatarDataJogando(jogo.iniciado_em)}. Ao salvar o abandono, ele sai de Jogando agora.`,
      onSalvo: async () => { await retirar(jogo.id, AVISO_ABANDONO(jogo.nome)); await dashboard.carregarDashboard() },
    })
  }

  async function remover(jogo: typeof jogando.jogos[number]) {
    try {
      await jogando.remover(jogo.id)
    } catch (error) {
      if (!(error instanceof ApiError && error.status === 404)) jogando.definirAviso(AVISO_REMOCAO)
    }
  }

  if (jogando.error) return <DashboardSecaoErro onRetry={() => void jogando.carregar()} />
  if (!jogando.carregado) return <section aria-label="Jogando agora"><Skeleton className="h-28 w-full bg-[var(--bg-surface)]" /></section>
  return <section aria-label="Jogando agora" className="space-y-3">{jogando.aviso && <p role="alert" className="rounded-lg border border-[var(--abandonado-banner-border)] bg-[var(--abandonado-banner-bg)] px-3 py-2 text-xs text-[var(--abandonado-banner-text)]">{jogando.aviso}</p>}{jogando.jogos.length === 0 ? <JogandoVazio onIniciar={jogando.abrirModalIniciar} /> : <><div className="flex items-center gap-2"><Play className="size-4 text-[var(--accent)]" aria-hidden="true" /><h2 className="text-base font-bold text-[var(--text-primary)]">Jogando agora</h2><span className="rounded-full bg-[var(--bg-surface-alt)] px-2 py-0.5 text-[11px] font-semibold text-[var(--text-secondary)]">{jogando.jogos.length}</span></div><div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">{jogando.jogos.map((jogo) => <JogandoCard key={jogo.id} jogo={jogo} onZerei={() => abrirZeramento(jogo)} onAbandonei={() => abrirAbandono(jogo)} onRemover={() => void remover(jogo)} />)}</div></>}</section>
}
