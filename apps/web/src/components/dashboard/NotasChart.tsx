import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import type { DashboardNotas } from '@/types/dashboard'

export function NotasChart({ notas, embutido = false }: { notas: DashboardNotas; embutido?: boolean }) {
  const conteudo = <><div className="h-48"><ResponsiveContainer width="100%" height="100%"><BarChart data={notas.histograma} aria-label="Histograma de notas"><CartesianGrid stroke="var(--border)" vertical={false} /><XAxis dataKey="nota" stroke="var(--text-muted)" fontSize={10} /><YAxis stroke="var(--text-muted)" fontSize={10} allowDecimals={false} /><Tooltip contentStyle={{ background: 'var(--bg-surface-alt)', border: '1px solid var(--border)' }} /><Bar dataKey="total" radius={[3, 3, 0, 0]}>{notas.histograma.map((item) => <Cell key={item.nota} fill={item.nota === 11 ? 'var(--highlight-gold)' : 'var(--accent)'} />)}</Bar></BarChart></ResponsiveContainer></div><p className="mt-2 text-[11px] text-[var(--text-muted)]">{notas.total_avaliados} jogos avaliados</p></>
  if (embutido) return conteudo
  return <section aria-label="Notas" className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><header className="mb-3 flex items-center justify-between"><h2 className="text-base font-bold text-[var(--text-primary)]">Notas</h2><span className="text-xs text-[var(--text-secondary)]">Média {notas.nota_media.toFixed(1)}</span></header>{conteudo}</section>
}
