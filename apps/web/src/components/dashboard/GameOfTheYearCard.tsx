import { Crown, Gamepad2, Star } from 'lucide-react'
import type { DashboardDados } from '@/types/dashboard'
import { iniciais } from '@/lib/dashboardUtils'

export function GameOfTheYearCard({ jogo }: { jogo: NonNullable<DashboardDados['jogoDoAno']> }) {
  return <section aria-label="Jogo do Ano" className="relative overflow-hidden rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-5"><div className="pointer-events-none absolute -right-16 -top-16 h-48 w-48 rounded-full bg-[var(--highlight-gold)] opacity-10 blur-3xl" /><div className="relative flex flex-col gap-5 sm:flex-row sm:items-center">
    <div className="relative aspect-[3/4] w-28 shrink-0 overflow-hidden rounded-[6px] border border-[var(--border-subtle)] sm:w-36">{jogo.igdb_capa_url ? <img src={jogo.igdb_capa_url} alt={jogo.nome} className="h-full w-full object-cover" /> : <div className="flex h-full items-center justify-center bg-[var(--bg-surface-alt)] text-2xl font-bold text-[var(--text-muted)]">{iniciais(jogo.nome)}</div>}<span className="absolute right-1.5 top-1.5 flex items-center gap-0.5 rounded bg-black/80 px-1.5 py-0.5 text-xs font-bold text-[var(--highlight-gold)]"><Star className="h-3 w-3 fill-current" aria-hidden="true" />{jogo.nota}</span></div>
    <div><span className="inline-flex items-center gap-1.5 rounded-full border border-[var(--highlight-gold)]/40 px-2.5 py-0.5 text-xs font-bold uppercase tracking-wider text-[var(--highlight-gold)]"><Crown className="h-3.5 w-3.5" aria-hidden="true" />Jogo do Ano {jogo.ano}</span><h2 className="mt-2 text-2xl font-bold text-[var(--text-primary)]">{jogo.nome}</h2><div className="mt-2 flex flex-wrap gap-3 text-xs text-[var(--text-secondary)]"><span className="inline-flex items-center gap-1"><Gamepad2 className="h-3.5 w-3.5" aria-hidden="true" />{jogo.console}</span></div></div>
  </div></section>
}
