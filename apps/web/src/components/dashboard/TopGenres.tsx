import { Compass } from 'lucide-react'
import type { GeneroEstatistica } from '@/mocks/dashboardData'

interface TopGenresProps {
  generos: GeneroEstatistica[]
}

export function TopGenres({ generos }: TopGenresProps) {
  return (
    <section
      aria-label="Gêneros Mais Jogados"
      className="space-y-4 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4 sm:p-5"
    >
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Compass className="h-4 w-4 text-[var(--accent)]" aria-hidden="true" />
          <h2 className="text-base font-bold text-[var(--text-primary)]">Gêneros Mais Jogados</h2>
        </div>
        <span className="text-[11px] text-[var(--text-muted)]">Preferências de gameplay</span>
      </header>

      <div className="space-y-3.5">
        {generos.map((item) => (
          <div key={item.genero} className="space-y-1.5">
            <div className="flex items-center justify-between text-[13px]">
              <span className="font-medium text-[var(--text-primary)]">{item.genero}</span>
              <div className="flex items-center gap-2">
                <span className="text-[11.5px] text-[var(--text-muted)]">
                  {item.totalJogos} jogos &bull; {item.horasJogadas.toLocaleString('pt-BR')}h
                </span>
                <span className="w-9 text-right font-semibold text-[var(--accent)]">
                  {item.percentual}%
                </span>
              </div>
            </div>

            <div className="h-1.5 w-full overflow-hidden rounded-[3px] bg-[var(--border)]">
              <div
                className="h-full rounded-[3px] bg-[var(--accent)] transition-all duration-300"
                style={{ width: `${item.percentual}%` }}
                role="progressbar"
                aria-valuenow={item.percentual}
                aria-valuemin={0}
                aria-valuemax={100}
                aria-label={`${item.genero}: ${item.percentual}%`}
              />
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
