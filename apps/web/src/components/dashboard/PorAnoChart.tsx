import { useState } from 'react'
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import type { DashboardAno } from '@/types/dashboard'

interface TickProps {
  x?: string | number
  y?: string | number
  payload?: { value: string | number }
}

export function PorAnoChart({ anos, modo: modoProp, onModoChange, mostrarControle = true }: { anos: DashboardAno[]; modo?: 'jogos' | 'horas'; onModoChange?: (modo: 'jogos' | 'horas') => void; mostrarControle?: boolean }) {
  const [modoLocal, setModoLocal] = useState<'jogos' | 'horas'>('jogos')
  const modo = modoProp ?? modoLocal
  const alterarModo = (proximo: 'jogos' | 'horas') => { setModoLocal(proximo); onModoChange?.(proximo) }
  if (!anos.length) return null
  const dados = anos.map((item) => ({
    ...item,
    valor: modo === 'jogos' ? item.total_jogos : Math.round(item.total_segundos / 3600),
  }))
  const renderTick = ({ x = 0, y = 0, payload }: TickProps) => {
    const item = anos.find((ano) => ano.ano === Number(payload?.value))
    const nome = item?.game_do_ano?.nome ?? 'sem destaque'
    return <g className="hidden sm:block" transform={`translate(${x},${y})`}>
      <title>{nome}</title>
      <text textAnchor="middle" y={0} fill="var(--text-muted)" fontSize={11}>{payload?.value}</text>
      <text textAnchor="middle" y={15} fill={item?.game_do_ano ? 'var(--text-secondary)' : 'var(--text-faint)'} fontSize={9}>{nome.length > 14 ? `${nome.slice(0, 14)}…` : nome}</text>
    </g>
  }
  return <section aria-label="Por Ano" className="space-y-4 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><header className="flex items-center justify-between"><h2 className="text-base font-bold text-[var(--text-primary)]">Por Ano</h2>{mostrarControle && <div className="flex rounded border border-[var(--border-subtle)] text-[10px]"><button type="button" onClick={() => alterarModo('jogos')} className={`px-2 py-1 ${modo === 'jogos' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Jogos</button><button type="button" onClick={() => alterarModo('horas')} className={`px-2 py-1 ${modo === 'horas' ? 'bg-[var(--accent)] text-[var(--accent-foreground)]' : 'text-[var(--text-muted)]'}`}>Horas</button></div>}</header><div className="overflow-x-auto"><div className="h-64 w-full min-w-[520px]"><ResponsiveContainer width="100%" height="100%"><BarChart data={dados} margin={{ bottom: 24 }} aria-label="Gráfico de jogos e horas por ano"><CartesianGrid stroke="var(--border)" vertical={false} /><XAxis dataKey="ano" stroke="var(--text-muted)" fontSize={11} tick={renderTick} /><YAxis stroke="var(--text-muted)" fontSize={11} allowDecimals={false} /><Tooltip contentStyle={{ background: 'var(--bg-surface-alt)', border: '1px solid var(--border)' }} formatter={(value) => [modo === 'jogos' ? value : `${value}h`, modo === 'jogos' ? 'Jogos' : 'Horas']} /><Bar dataKey="valor" radius={[4, 4, 0, 0]}>{dados.map((item) => <Cell key={item.ano} fill="var(--accent)" fillOpacity={item.ano === new Date().getFullYear() ? 1 : 0.75} />)}</Bar></BarChart></ResponsiveContainer></div></div><div className="grid grid-cols-2 gap-2 text-[10px] text-[var(--text-muted)]">{anos.map((item) => <span key={item.ano}>{item.ano}: {item.game_do_ano?.nome ?? 'sem destaque'}</span>)}</div></section>
}
