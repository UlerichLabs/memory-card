import { CalendarDays } from 'lucide-react'
import { calcularIntensidade, type AtividadeAno, type PontoAtividade } from '@/mocks/dashboardData'

function obterClasseIntensidade(intensidade: number): string {
  switch (intensidade) {
    case 1:
      return 'bg-[var(--accent)]/30'
    case 2:
      return 'bg-[var(--accent)]/55'
    case 3:
      return 'bg-[var(--accent)]/80'
    case 4:
      return 'bg-[var(--accent)]'
    default:
      return 'bg-[var(--bg-surface-alt)]'
  }
}

interface ActivityHeatmapProps {
  atividade: AtividadeAno
}

export function ActivityHeatmap({ atividade }: ActivityHeatmapProps) {
  const meses = ['Jan', 'Fev', 'Mar', 'Abr', 'Mai', 'Jun', 'Jul', 'Ago', 'Set', 'Out', 'Nov', 'Dez']
  const semanas: PontoAtividade[][] = []

  for (let i = 0; i < atividade.pontos.length; i += 7) {
    semanas.push(atividade.pontos.slice(i, i + 7))
  }

  return (
    <section
      aria-label="Atividade do Ano"
      className="space-y-3 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4 sm:p-5"
    >
      <header className="flex flex-col justify-between gap-1 sm:flex-row sm:items-center">
        <div className="flex items-center gap-2">
          <CalendarDays className="h-4 w-4 text-[var(--accent)]" aria-hidden="true" />
          <h2 className="text-base font-bold text-[var(--text-primary)]">
            Atividade em {atividade.ano}
          </h2>
        </div>
        <p className="text-[12px] text-[var(--text-secondary)]">
          <span className="font-semibold text-[var(--text-primary)]">{atividade.totalHoras}h</span> em{' '}
          <span className="font-semibold text-[var(--text-primary)]">{atividade.diasAtivos}</span> dias
        </p>
      </header>

      <div className="space-y-2">
        <div className="flex justify-between text-[11px] text-[var(--text-muted)]">
          {meses.map((mes) => (
            <span key={mes}>{mes}</span>
          ))}
        </div>

        <div className="overflow-x-auto pb-1" tabIndex={0} aria-label="Tabela de atividade diária">
          <div className="flex min-w-[650px] gap-1">
            {semanas.map((semana, idx) => (
              <div key={idx} className="flex flex-col gap-1">
                {semana.map((ponto) => {
                  const intensidade = calcularIntensidade(ponto.horas)
                  return (
                    <div
                      key={ponto.data}
                      className={`h-2.5 w-2.5 rounded-[2px] ${obterClasseIntensidade(intensidade)}`}
                      title={`${ponto.data}: ${ponto.horas}h jogadas`}
                    />
                  )
                })}
              </div>
            ))}
          </div>
        </div>

        <div className="flex items-center justify-end gap-1.5 pt-1 text-[11px] text-[var(--text-muted)]">
          <span>Menos</span>
          {[0, 1, 2, 3, 4].map((nivel) => (
            <div
              key={nivel}
              className={`h-2.5 w-2.5 rounded-[2px] ${obterClasseIntensidade(nivel)}`}
            />
          ))}
          <span>Mais</span>
        </div>
      </div>
    </section>
  )
}
