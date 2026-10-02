import { useState } from 'react'
import { Star } from 'lucide-react'
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import type { DashboardAno } from '@/types/dashboard'

interface TickProps {
  x?: string | number
  y?: string | number
  payload?: { value: string | number }
}

export function PorAnoChart({ anos, modo: modoProp, onModoChange, mostrarControle = true, embutido = false }: { anos: DashboardAno[]; modo?: 'jogos' | 'horas'; onModoChange?: (modo: 'jogos' | 'horas') => void; mostrarControle?: boolean; embutido?: boolean }) {
  const [modoLocal, setModoLocal] = useState<'jogos' | 'horas'>('jogos')
  const modo = modoProp ?? modoLocal
  const alterarModo = (proximo: 'jogos' | 'horas') => { setModoLocal(proximo); onModoChange?.(proximo) }
  if (!anos.length) return null
  const dados = anos.map((item) => ({
    ...item,
    valor: modo === 'jogos' ? item.total_jogos : Math.round(item.total_segundos / 3600),
  }))
  const renderTick = ({ x = 0, y = 0, payload }: TickProps) => <g transform={`translate(${x},${y})`}>
    <text textAnchor="middle" y={12} fill="var(--text-muted)" fontSize={11}>{payload?.value}</text>
  </g>
  const conteudo = <>
    <div className="overflow-x-auto">
      <div className="h-[200px] min-w-[520px] sm:h-[240px]">
        <ResponsiveContainer width="100%" height="100%"><BarChart data={dados} margin={{ bottom: 8 }} aria-label="Gráfico de jogos e horas por ano"><CartesianGrid stroke="var(--border)" vertical={false} /><XAxis dataKey="ano" height={32} stroke="var(--text-muted)" fontSize={11} tick={renderTick} /><YAxis stroke="var(--text-muted)" fontSize={11} allowDecimals={false} /><Tooltip contentStyle={{ background: 'var(--bg-surface-alt)', border: '1px solid var(--border)' }} formatter={(value) => [modo === 'jogos' ? value : `${value}h`, modo === 'jogos' ? 'Jogos' : 'Horas']} /><Bar dataKey="valor" maxBarSize={42} radius={[4, 4, 0, 0]}>{dados.map((item) => <Cell key={item.ano} fill="var(--accent)" fillOpacity={item.ano === new Date().getFullYear() ? 1 : 0.75} />)}</Bar></BarChart></ResponsiveContainer>
      </div>
      <ul className="grid min-w-[520px] border-t border-[var(--border)] pt-2 text-[10px] text-[var(--text-muted)]" style={{ gridTemplateColumns: `repeat(${anos.length}, minmax(0, 1fr))` }}>
        {anos.map((item) => <li key={item.ano} className="min-w-0 px-1 text-center" aria-label={`${item.ano}: ${item.game_do_ano?.nome ?? 'sem destaque'}`}><span className="block text-[10px] text-[var(--text-faint)]">{item.ano}</span>{item.game_do_ano ? <span className="flex min-w-0 items-center justify-center gap-1"><Star className="size-3 shrink-0 text-[var(--highlight-gold)]" fill="currentColor" aria-hidden="true" /><span className="truncate" title={item.game_do_ano.nome}>{item.game_do_ano.nome}</span></span> : <span className="block truncate text-[var(--text-faint)]" title="sem destaque">— <span>sem destaque</span></span>}</li>)}
      </ul>
    </div>
  </>
  if (embutido) return conteudo
  return <section aria-label="Por Ano" className="space-y-4 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><header className="flex items-center justify-between"><h2 className="text-base font-bold text-[var(--text-primary)]">Por Ano</h2>{mostrarControle && <div className="flex rounded border border-[var(--border-subtle)] text-[10px]"><button type="button" onClick={() => alterarModo('jogos')} className={`px-2 py-1 ${modo === 'jogos' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Jogos</button><button type="button" onClick={() => alterarModo('horas')} className={`px-2 py-1 ${modo === 'horas' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Horas</button></div>}</header>{conteudo}</section>
}
