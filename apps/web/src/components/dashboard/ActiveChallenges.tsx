import { Target, Calendar } from 'lucide-react'
import type { DesafioAtivo } from '@/mocks/dashboardData'

interface ActiveChallengesProps {
  desafios: DesafioAtivo[]
}

export function ActiveChallenges({ desafios }: ActiveChallengesProps) {
  const desafiosExibidos = desafios.slice(0, 3)

  return (
    <section aria-label="Desafios Ativos" className="space-y-3">
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Target className="h-4 w-4 text-[var(--accent)]" aria-hidden="true" />
          <h2 className="text-base font-bold text-[var(--text-primary)]">Desafios Ativos</h2>
        </div>
        <span className="text-[11px] text-[var(--text-muted)]">Máximo 3 em andamento</span>
      </header>

      <div className="grid grid-cols-1 gap-3 sm:gap-4 md:grid-cols-3">
        {desafiosExibidos.map((desafio) => {
          const percentual = Math.min(
            100,
            Math.round((desafio.progressoAtual / Math.max(desafio.progressoMeta, 1)) * 100)
          )

          return (
            <article
              key={desafio.id}
              className="flex flex-col justify-between rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"
            >
              <div className="space-y-1">
                <div className="flex items-center justify-between gap-2">
                  <h3 className="truncate text-[13.5px] font-bold text-[var(--text-primary)]">
                    {desafio.titulo}
                  </h3>
                  <span className="shrink-0 text-[12px] font-semibold text-[var(--accent)]">
                    {percentual}%
                  </span>
                </div>
                <p className="line-clamp-2 text-[12px] text-[var(--text-secondary)]">
                  {desafio.descricao}
                </p>
              </div>

              <div className="mt-4 space-y-2">
                <div className="h-1.5 w-full overflow-hidden rounded-[3px] bg-[var(--border)]">
                  <div
                    className="h-full rounded-[3px] bg-[var(--accent)] transition-all duration-300"
                    style={{ width: `${percentual}%` }}
                    role="progressbar"
                    aria-valuenow={desafio.progressoAtual}
                    aria-valuemin={0}
                    aria-valuemax={desafio.progressoMeta}
                  />
                </div>

                <div className="flex items-center justify-between text-[11px] text-[var(--text-muted)]">
                  <span>
                    {desafio.progressoAtual} / {desafio.progressoMeta} {desafio.unidade}
                  </span>
                  {desafio.prazo && (
                    <span className="inline-flex items-center gap-1">
                      <Calendar className="h-3 w-3" aria-hidden="true" />
                      {desafio.prazo}
                    </span>
                  )}
                </div>
              </div>
            </article>
          )
        })}
      </div>
    </section>
  )
}
