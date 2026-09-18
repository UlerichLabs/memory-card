import { Crown, Star, Clock, Gamepad2, Calendar } from 'lucide-react'
import type { JogoDoAno } from '@/mocks/dashboardData'

interface GameOfTheYearCardProps {
  jogoDoAno: JogoDoAno
}

export function GameOfTheYearCard({ jogoDoAno }: GameOfTheYearCardProps) {
  const { ano, jogo, comentario } = jogoDoAno

  return (
    <section aria-label="Jogo do Ano" className="relative overflow-hidden rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <div
        className="pointer-events-none absolute -right-16 -top-16 h-48 w-48 rounded-full opacity-10 blur-3xl"
        style={{ backgroundColor: 'var(--highlight-gold)' }}
        aria-hidden="true"
      />
      <div className="relative z-10 flex flex-col gap-5 sm:flex-row sm:items-center">
        <div className="relative aspect-[3/4] w-28 shrink-0 overflow-hidden rounded-[6px] border border-[var(--border-subtle)] sm:w-36">
          <img
            src={jogo.capaUrl}
            alt={jogo.titulo}
            className="h-full w-full object-cover"
            loading="lazy"
          />
          <div className="absolute right-1.5 top-1.5 flex items-center gap-0.5 rounded bg-black/80 px-1.5 py-0.5 text-xs font-bold text-[var(--highlight-gold)]">
            <Star className="h-3 w-3 fill-[var(--highlight-gold)]" aria-hidden="true" />
            <span>{jogo.nota.toFixed(0)}</span>
          </div>
        </div>

        <div className="flex flex-1 flex-col justify-between space-y-3">
          <div>
            <div className="inline-flex items-center gap-1.5 rounded-full border border-[var(--highlight-gold)]/40 bg-[var(--highlight-gold)]/10 px-2.5 py-0.5 text-xs font-bold uppercase tracking-wider text-[var(--highlight-gold)]">
              <Crown className="h-3.5 w-3.5" aria-hidden="true" />
              <span>Jogo do Ano {ano}</span>
            </div>
            <h2 className="mt-2 text-xl font-bold tracking-tight text-[var(--text-primary)] sm:text-2xl">
              {jogo.titulo}
            </h2>
            <div className="mt-1 flex flex-wrap items-center gap-3 text-[12px] text-[var(--text-secondary)]">
              <span className="inline-flex items-center gap-1">
                <Gamepad2 className="h-3.5 w-3.5 text-[var(--text-faint)]" aria-hidden="true" />
                {jogo.plataforma}
              </span>
              <span className="inline-flex items-center gap-1">
                <Clock className="h-3.5 w-3.5 text-[var(--text-faint)]" aria-hidden="true" />
                {jogo.horasJogadas}h jogadas
              </span>
              <span className="inline-flex items-center gap-1">
                <Calendar className="h-3.5 w-3.5 text-[var(--text-faint)]" aria-hidden="true" />
                Lançamento {jogo.anoLancamento}
              </span>
            </div>
          </div>

          {comentario && (
            <blockquote className="border-l-2 border-[var(--highlight-gold)]/60 pl-3 text-[13px] italic leading-relaxed text-[var(--text-secondary)]">
              &ldquo;{comentario}&rdquo;
            </blockquote>
          )}
        </div>
      </div>
    </section>
  )
}
