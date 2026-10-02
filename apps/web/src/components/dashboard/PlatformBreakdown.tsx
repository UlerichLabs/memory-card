import { Monitor } from 'lucide-react'
import { useState } from 'react'
import type { DashboardRankingPlataforma } from '@/types/dashboard'
import { formatarHoras, formatarPercentual } from '@/lib/dashboardUtils'

export function PlatformBreakdown({ plataformas, modo: modoProp, onModoChange, mostrarControle = true, embutido = false }: { plataformas: DashboardRankingPlataforma[]; modo?: 'jogos' | 'horas'; onModoChange?: (modo: 'jogos' | 'horas') => void; mostrarControle?: boolean; embutido?: boolean }) {
  const [modoLocal, setModoLocal] = useState<'jogos' | 'horas'>('jogos')
  const modo = modoProp ?? modoLocal
  const alterarModo = (proximo: 'jogos' | 'horas') => { setModoLocal(proximo); onModoChange?.(proximo) }
  if (!plataformas.length) return null
  const itens = [...plataformas].sort((a, b) => (modo === 'jogos' ? b.percentual_jogos - a.percentual_jogos : b.percentual_segundos - a.percentual_segundos))
  const conteudo = <div className={embutido ? 'grid grid-cols-1 gap-x-12 gap-y-4 lg:grid-cols-2' : 'space-y-3'}>{itens.map((item) => { const percentual = modo === 'jogos' ? item.percentual_jogos : item.percentual_segundos; return <div key={item.console}><div className="flex justify-between text-[13px]"><span className="font-medium text-[var(--text-primary)]">{item.console}</span><span className="text-xs text-[var(--text-muted)]">{modo === 'jogos' ? item.total_jogos : formatarHoras(item.total_segundos)} · {formatarPercentual(percentual)}%</span></div><div className="mt-1.5 h-1.5 overflow-hidden rounded bg-[var(--border)]"><div className="h-full rounded bg-[var(--accent)]" style={{ width: `${percentual}%` }} role="progressbar" aria-valuenow={percentual} aria-valuemin={0} aria-valuemax={100} /></div></div>})}</div>
  if (embutido) return conteudo
  return <section aria-label="Distribuição por Plataforma" className="space-y-4 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><header className="flex items-center justify-between gap-2"><div className="flex items-center gap-2"><Monitor className="h-4 w-4 text-[var(--accent)]" aria-hidden="true" /><h2 className="text-base font-bold text-[var(--text-primary)]">Por Plataforma</h2></div>{mostrarControle && <div className="flex rounded-full border border-[var(--border-subtle)] text-[10px]"><button type="button" onClick={() => alterarModo('jogos')} className={`rounded-full px-2 py-1 ${modo === 'jogos' ? 'btn-primario btn-primario-ativo' : 'text-[var(--text-muted)]'}`}>Jogos</button><button type="button" onClick={() => alterarModo('horas')} className={`rounded-full px-2 py-1 ${modo === 'horas' ? 'btn-primario btn-primario-ativo' : 'text-[var(--text-muted)]'}`}>Horas</button></div>}</header>{conteudo}</section>
}
