import { useState, type KeyboardEvent, type ReactNode } from 'react'
import { BarChart3, ChartNoAxesColumn, Gauge, Layers3, ListChecks, Trophy } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import type { DashboardStore } from '@/stores/dashboardStore'
import { Skeleton } from '@/components/ui/skeleton'
import { DashboardSecaoErro } from './DashboardSecaoErro'
import { PorAnoChart } from './PorAnoChart'
import { NotasChart } from './NotasChart'
import { DificuldadeBreakdown } from './DificuldadeBreakdown'
import { PlatformBreakdown } from './PlatformBreakdown'
import { TopGenres } from './TopGenres'
import { Recordes } from './Recordes'

type Aba = 'porAno' | 'notas' | 'dificuldade' | 'plataformas' | 'generos' | 'recordes'
type Modo = 'jogos' | 'horas'
interface AbaDef { id: Aba; label: string; subtitulo: string; icon: LucideIcon }

const abas: AbaDef[] = [
  { id: 'porAno', label: 'Por Ano', subtitulo: 'Evolução dos seus zeramentos', icon: BarChart3 },
  { id: 'notas', label: 'Notas', subtitulo: 'Distribuição das suas avaliações', icon: ChartNoAxesColumn },
  { id: 'dificuldade', label: 'Dificuldade', subtitulo: 'Como você encara os desafios', icon: Gauge },
  { id: 'plataformas', label: 'Plataformas', subtitulo: 'Onde você mais joga', icon: Layers3 },
  { id: 'generos', label: 'Gêneros', subtitulo: 'Seus gêneros mais jogados', icon: ListChecks },
  { id: 'recordes', label: 'Recordes', subtitulo: 'Marcas da sua jornada', icon: Trophy },
]

function Bloco({ erro, carregando, pronto, retry, children }: { erro?: string; carregando: boolean; pronto: boolean; retry: () => void; children: ReactNode }) {
  if (erro) return <DashboardSecaoErro onRetry={retry} />
  if (carregando || !pronto) return <Skeleton className="h-48 w-full bg-[var(--bg-surface)]" />
  return <>{children}</>
}

function ToggleModo({ modo, onChange }: { modo: Modo; onChange: (modo: Modo) => void }) {
  return <div aria-label="Modo do gráfico" className="flex rounded border border-[var(--border-subtle)] text-[10px]"><button type="button" onClick={() => onChange('jogos')} className={`px-2 py-1 ${modo === 'jogos' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Jogos</button><button type="button" onClick={() => onChange('horas')} className={`px-2 py-1 ${modo === 'horas' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Horas</button></div>
}

function NotasResumo({ dashboard }: { dashboard: DashboardStore }) {
  if (!dashboard.notas) return null
  const comum = dashboard.notas.histograma.reduce((atual, item) => item.total > atual.total ? item : atual, { nota: 0, total: 0 })
  return <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_220px]"><NotasChart notas={dashboard.notas} /><div className="grid grid-cols-2 gap-3 lg:grid-cols-1"><article className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><p className="text-xs text-[var(--text-muted)]">Nota média</p><strong className="mt-2 block text-2xl font-bold tabular-nums text-[var(--text-primary)]">{dashboard.notas.nota_media.toFixed(1)}</strong></article><article className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><p className="text-xs text-[var(--text-muted)]">Nota mais comum</p><strong className="mt-2 block text-2xl font-bold tabular-nums text-[var(--highlight-gold)]">{comum.total ? comum.nota : '—'}</strong><span className="text-[11px] text-[var(--text-muted)]">{comum.total} {comum.total === 1 ? 'jogo' : 'jogos'}</span></article></div></div>
}

export function EstatisticasAbas({ dashboard }: { dashboard: DashboardStore }) {
  const [aba, setAba] = useState<Aba>('porAno')
  const [modo, setModo] = useState<Modo>('jogos')
  const ativa = abas.find((item) => item.id === aba) ?? abas[0]
  const indice = abas.findIndex((item) => item.id === aba)
  const selecionarComTeclado = (event: KeyboardEvent<HTMLButtonElement>) => {
    if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') return
    event.preventDefault()
    const proximo = event.key === 'ArrowRight' ? (indice + 1) % abas.length : (indice - 1 + abas.length) % abas.length
    setAba(abas[proximo].id)
    document.getElementById(`estatistica-tab-${abas[proximo].id}`)?.focus()
  }
  const painel = () => {
    if (aba === 'porAno') return <Bloco erro={dashboard.errors.porAno} carregando={dashboard.loading.porAno} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('porAno')}><PorAnoChart anos={dashboard.porAno} modo={modo} onModoChange={setModo} mostrarControle={false} /></Bloco>
    if (aba === 'notas') return <Bloco erro={dashboard.errors.notas} carregando={dashboard.loading.notas} pronto={Boolean(dashboard.notas)} retry={() => void dashboard.recarregarBloco('notas')}><NotasResumo dashboard={dashboard} /></Bloco>
    if (aba === 'dificuldade') return <Bloco erro={dashboard.errors.dificuldade} carregando={dashboard.loading.dificuldade} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('dificuldade')}><DificuldadeBreakdown dificuldade={dashboard.dificuldade} /></Bloco>
    if (aba === 'plataformas') return <Bloco erro={dashboard.errors.plataformas} carregando={dashboard.loading.plataformas} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('plataformas')}><PlatformBreakdown plataformas={dashboard.plataformas} modo={modo} onModoChange={setModo} mostrarControle={false} /></Bloco>
    if (aba === 'generos') return <Bloco erro={dashboard.errors.generos} carregando={dashboard.loading.generos} pronto={dashboard.carregado} retry={() => void dashboard.recarregarBloco('generos')}><TopGenres generos={dashboard.generos} tipos={dashboard.tipos} modo={modo} onModoChange={setModo} mostrarControle={false} /></Bloco>
    return <Bloco erro={dashboard.errors.recordes} carregando={dashboard.loading.recordes} pronto={Boolean(dashboard.recordes)} retry={() => void dashboard.recarregarBloco('recordes')}>{dashboard.recordes && <Recordes recordes={dashboard.recordes} compact />}</Bloco>
  }
  const temToggle = aba === 'porAno' || aba === 'plataformas' || aba === 'generos'
  return <section aria-label="Estatísticas" className="space-y-4 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-4 sm:p-5"><header className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><div><div className="flex items-center gap-2"><BarChart3 className="size-5 text-[var(--accent)]" aria-hidden="true" /><h2 className="text-base font-bold text-[var(--text-primary)]">Estatísticas</h2></div><p className="mt-1 text-xs text-[var(--text-muted)]">{ativa.subtitulo}</p></div>{temToggle && <ToggleModo modo={modo} onChange={setModo} />}</header><div role="tablist" aria-label="Estatísticas do dashboard" className="flex max-w-full gap-2 overflow-x-auto pb-1"><div className="flex min-w-max gap-2">{abas.map(({ id, label, icon: Icon }) => <button key={id} id={`estatistica-tab-${id}`} type="button" role="tab" aria-selected={aba === id} aria-controls={`estatistica-panel-${id}`} tabIndex={aba === id ? 0 : -1} onClick={() => setAba(id)} onKeyDown={selecionarComTeclado} className={`inline-flex h-8 items-center gap-1.5 rounded-full border px-3 text-xs font-semibold transition-colors ${aba === id ? 'border-[var(--accent)] bg-[var(--accent)] text-[var(--accent-foreground)]' : 'border-[var(--border-subtle)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)]'}`}><Icon className="size-3.5" aria-hidden="true" />{label}</button>)}</div></div><div id={`estatistica-panel-${aba}`} role="tabpanel" aria-labelledby={`estatistica-tab-${aba}`}>{painel()}</div></section>
}
