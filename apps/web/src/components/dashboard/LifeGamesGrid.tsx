import { Star } from 'lucide-react'
import type { JogoDaVida } from '@/mocks/dashboardData'

interface LifeGamesGridProps {
  jogos: JogoDaVida[]
}

export function LifeGamesGrid({ jogos }: LifeGamesGridProps) {
  return (
    <section aria-label="5 Jogos da Vida" className="space-y-3">
      <header className="flex items-center justify-between">
        <h2 className="text-base font-bold text-[var(--text-primary)]">5 Jogos da Vida</h2>
        <span className="text-[11px] text-[var(--text-muted)]">Favoritos de todos os tempos</span>
      </header>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5 sm:gap-4">
        {jogos.map(({ posicao, jogo }) => (
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
              <div className="absolute left-1.5 top-1.5 rounded bg-black/80 px-1.5 py-0.5 text-[11px] font-bold text-[var(--text-primary)]">
                #{posicao}
              </div>
              <div className="absolute right-1.5 top-1.5 flex items-center gap-0.5 rounded bg-black/80 px-1.5 py-0.5 text-[11px] font-bold text-[var(--highlight-gold)]">
                <Star className="h-2.5 w-2.5 fill-[var(--highlight-gold)]" aria-hidden="true" />
                <span>{jogo.nota.toFixed(0)}</span>
              </div>
            </div>

            <div className="mt-2 space-y-0.5">
              <h3 className="truncate text-[13px] font-medium text-[var(--text-primary)]" title={jogo.titulo}>
                {jogo.titulo}
              </h3>
              <p className="text-[11px] text-[var(--text-muted)]">
                {jogo.plataforma} &bull; {jogo.anoLancamento}
              </p>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
