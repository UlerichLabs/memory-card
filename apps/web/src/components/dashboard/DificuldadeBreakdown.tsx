import type { DashboardDificuldade } from '@/types/dashboard'
import { formatarPercentual } from '@/lib/dashboardUtils'

const ordem: DashboardDificuldade['dificuldade'][] = ['C', 'B', 'A', 'AA', 'AAA']
export function DificuldadeBreakdown({ dificuldade, embutido = false }: { dificuldade: DashboardDificuldade[]; embutido?: boolean }) {
  const mapa = new Map(dificuldade.map((item) => [item.dificuldade, item]))
  const conteudo = <div className={embutido ? 'grid grid-cols-1 gap-x-12 gap-y-4 lg:grid-cols-2' : 'space-y-3'}>{ordem.map((nivel) => { const item = mapa.get(nivel) ?? { dificuldade: nivel, total_jogos: 0, percentual: 0 }; return <div key={nivel}><div className="flex justify-between text-xs"><span className="font-bold text-[var(--text-primary)]">{nivel}</span><span className="text-[var(--text-muted)]">{item.total_jogos} jogos · {formatarPercentual(item.percentual)}%</span></div><div className="mt-1.5 h-1.5 overflow-hidden rounded bg-[var(--border)]"><div className="h-full rounded" style={{ width: `${item.percentual}%`, backgroundColor: `var(--difficulty-${nivel.toLowerCase()})` }} /></div></div>})}</div>
  if (embutido) return conteudo
  return <section aria-label="Dificuldade" className="space-y-4 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><h2 className="text-base font-bold text-[var(--text-primary)]">Dificuldade</h2>{conteudo}</section>
}
