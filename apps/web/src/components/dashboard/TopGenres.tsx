import { Compass } from 'lucide-react'
import { useState } from 'react'
import type { DashboardRankingGenero, DashboardTipo } from '@/types/dashboard'
import { formatarHoras, formatarPercentual } from '@/lib/dashboardUtils'

export function TopGenres({ generos, tipos, modo: modoProp, onModoChange, mostrarControle = true, embutido = false }: { generos: DashboardRankingGenero[]; tipos: DashboardTipo[]; modo?: 'jogos' | 'horas'; onModoChange?: (modo: 'jogos' | 'horas') => void; mostrarControle?: boolean; embutido?: boolean }) {
  const [modoLocal, setModoLocal] = useState<'jogos' | 'horas'>('jogos')
  const modo = modoProp ?? modoLocal
  const alterarModo = (proximo: 'jogos' | 'horas') => { setModoLocal(proximo); onModoChange?.(proximo) }
  if (!generos.length) return null
  const itens = [...generos].sort((a, b) => (modo === 'jogos' ? b.percentual_jogos - a.percentual_jogos : b.percentual_segundos - a.percentual_segundos))
  const conteudo = <div className={embutido ? 'grid grid-cols-1 gap-x-12 gap-y-4 lg:grid-cols-2' : 'space-y-3'}>{itens.map((item, index) => { const percentual = modo === 'jogos' ? item.percentual_jogos : item.percentual_segundos; return <div key={item.genero}><div className="flex justify-between text-[13px]"><span className="font-medium text-[var(--text-primary)]">{item.genero}</span><span className="text-xs text-[var(--text-muted)]">{modo === 'jogos' ? item.total_jogos : formatarHoras(item.total_segundos)} · {formatarPercentual(percentual)}%</span></div><div className="mt-1.5 h-1.5 overflow-hidden rounded bg-[var(--border)]"><div className="h-full rounded bg-[var(--accent)]" style={{ width: `${percentual}%` }} /></div>{index === 0 && tipos.length > 0 && <div className="mt-2 flex flex-wrap gap-1.5">{tipos.map((tipo) => <span key={tipo.tipo} className="rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2 py-0.5 text-[10px] text-[var(--text-secondary)]">{tipo.tipo} {tipo.total_jogos}</span>)}</div>}</div>})}</div>
  if (embutido) return conteudo
  return <section aria-label="Gêneros Mais Jogados" className="space-y-4 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><header className="flex items-center justify-between gap-2"><div className="flex items-center gap-2"><Compass className="h-4 w-4 text-[var(--accent)]" aria-hidden="true" /><h2 className="text-base font-bold text-[var(--text-primary)]">Gêneros Mais Jogados</h2></div>{mostrarControle && <div className="flex rounded border border-[var(--border-subtle)] text-[10px]"><button type="button" onClick={() => alterarModo('jogos')} className={`px-2 py-1 ${modo === 'jogos' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Jogos</button><button type="button" onClick={() => alterarModo('horas')} className={`px-2 py-1 ${modo === 'horas' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Horas</button></div>}</header>{conteudo}</section>
}
