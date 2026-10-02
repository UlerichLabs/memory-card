import { Monitor } from 'lucide-react'
import { useState } from 'react'
import type { DashboardRankingPlataforma } from '@/types/dashboard'
import { formatarHoras, formatarPercentual } from '@/lib/dashboardUtils'

export function PlatformBreakdown({ plataformas, modo: modoProp, onModoChange, mostrarControle = true }: { plataformas: DashboardRankingPlataforma[]; modo?: 'jogos' | 'horas'; onModoChange?: (modo: 'jogos' | 'horas') => void; mostrarControle?: boolean }) {
  const [modoLocal, setModoLocal] = useState<'jogos' | 'horas'>('jogos')
  const modo = modoProp ?? modoLocal
  const alterarModo = (proximo: 'jogos' | 'horas') => { setModoLocal(proximo); onModoChange?.(proximo) }
  if (!plataformas.length) return null
  const itens = [...plataformas].sort((a, b) => (modo === 'jogos' ? b.percentual_jogos - a.percentual_jogos : b.percentual_segundos - a.percentual_segundos))
  return <section aria-label="Distribuição por Plataforma" className="space-y-4 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><header className="flex items-center justify-between gap-2"><div className="flex items-center gap-2"><Monitor className="h-4 w-4 text-[var(--accent)]" aria-hidden="true" /><h2 className="text-base font-bold text-[var(--text-primary)]">Por Plataforma</h2></div>{mostrarControle && <div className="flex rounded border border-[var(--border-subtle)] text-[10px]"><button type="button" onClick={() => alterarModo('jogos')} className={`px-2 py-1 ${modo === 'jogos' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Jogos</button><button type="button" onClick={() => alterarModo('horas')} className={`px-2 py-1 ${modo === 'horas' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Horas</button></div>}</header><div className="space-y-3">{itens.map((item) => { const percentual = modo === 'jogos' ? item.percentual_jogos : item.percentual_segundos; return <div key={item.console}><div className="flex justify-between text-[13px]"><span className="font-medium text-[var(--text-primary)]">{item.console}</span><span className="text-xs text-[var(--text-muted)]">{modo === 'jogos' ? item.total_jogos : formatarHoras(item.total_segundos)} · {formatarPercentual(percentual)}%</span></div><div className="mt-1.5 h-1.5 overflow-hidden rounded bg-[var(--border)]"><div className="h-full rounded bg-[var(--accent)]" style={{ width: `${percentual}%` }} role="progressbar" aria-valuenow={percentual} aria-valuemin={0} aria-valuemax={100} /></div></div>})}</div></section>
}
