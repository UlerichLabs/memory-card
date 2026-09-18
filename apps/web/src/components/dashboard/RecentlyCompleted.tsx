import { Star, CheckCircle2, Clock, Calendar } from 'lucide-react'
import type { JogoZeradoRecente } from '@/mocks/dashboardData'

interface RecentlyCompletedProps {
  jogos: JogoZeradoRecente[]
}

export function RecentlyCompleted({ jogos }: RecentlyCompletedProps) {
  return (
    <section aria-label="Zerados Recentemente" className="space-y-3">
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CheckCircle2 className="h-4 w-4 text-[var(--accent)]" aria-hidden="true" />
          <h2 className="text-base font-bold text-[var(--text-primary)]">Zerados Recentemente</h2>
        </div>
        <span className="text-[11px] text-[var(--text-muted)]">Últimas conquistas</span>
      </header>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-2 md:grid-cols-4 sm:gap-4">
        {jogos.map((jogo) => (
          <article
            key={jogo.id}
            className="group flex flex-col transition-transform duration-200 hover:-translate-y-1"
          >
            <div className="relative aspect-[3/4] w-full overflow-hidden rounded-[6px] border border-[var(--border)] bg-[var(--bg-surface-alt)]">
              <img
                src={jogo.capaUrl}
                alt={jogo.titulo}
                className="h-full w-full object-cover transition-opacity duration-200 group-hover:opacity-90"
                loading="lazy"
              />
              {jogo.dificuldade && (
                <div className="absolute left-1.5 top-1.5 rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)]/90 px-2 py-0.5 text-[10px] font-semibold text-[var(--text-secondary)]">
                  Dif. {jogo.dificuldade}
                </div>
              )}
              <div className="absolute right-1.5 top-1.5 flex items-center gap-0.5 rounded bg-black/80 px-1.5 py-0.5 text-[11px] font-bold text-[var(--highlight-gold)]">
                <Star className="h-2.5 w-2.5 fill-[var(--highlight-gold)]" aria-hidden="true" />
                <span>{jogo.nota.toFixed(1)}</span>
              </div>
            </div>

            <div className="mt-2 space-y-1">
              <h3 className="truncate text-[13px] font-semibold text-[var(--text-primary)]" title={jogo.titulo}>
                {jogo.titulo}
              </h3>
              <div className="flex items-center justify-between text-[11px] text-[var(--text-muted)]">
                <span>{jogo.plataforma}</span>
                <span className="inline-flex items-center gap-0.5">
                  <Clock className="h-3 w-3" aria-hidden="true" />
                  {jogo.horasJogadas}h
                </span>
              </div>
              <div className="flex items-center gap-1 text-[10.5px] text-[var(--text-faint)]">
                <Calendar className="h-3 w-3" aria-hidden="true" />
                <span>Concluído em {jogo.dataFinalizacao}</span>
              </div>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
