import { Trophy, Clock, Gamepad2, Calendar } from 'lucide-react'
import type { EstatisticasGerais } from '@/mocks/dashboardData'

interface StatsRowProps {
  estatisticas: EstatisticasGerais
}

export function StatsRow({ estatisticas }: StatsRowProps) {
  const cards = [
    {
      label: 'Jogos Zerados',
      valor: estatisticas.totalJogosZerados.toString(),
      subtexto: 'Total histórico',
      icone: Trophy,
    },
    {
      label: 'Horas Jogadas',
      valor: `${estatisticas.totalHorasJogadas.toLocaleString('pt-BR')}h`,
      subtexto: 'Tempo registrado',
      icone: Clock,
    },
    {
      label: 'Média por Jogo',
      valor: `${estatisticas.mediaHorasPorJogo.toFixed(1)}h`,
      subtexto: 'Duração média',
      icone: Gamepad2,
    },
    {
      label: 'Zerados este Ano',
      valor: estatisticas.jogosZeradosAnoAtual.toString(),
      subtexto: 'Ano vigente',
      icone: Calendar,
    },
  ]

  return (
    <section aria-label="Estatísticas Gerais" className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
      {cards.map((card) => {
        const Icone = card.icone
        return (
          <div
            key={card.label}
            className="flex flex-col justify-between rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"
          >
            <div className="flex items-center justify-between">
              <span className="text-[13px] font-medium text-[var(--text-secondary)]">{card.label}</span>
              <Icone className="h-4 w-4 text-[var(--text-faint)]" aria-hidden="true" />
            </div>
            <div className="mt-3">
              <p className="text-2xl font-bold tracking-tight text-[var(--text-primary)] sm:text-3xl">
                {card.valor}
              </p>
              <p className="mt-0.5 text-[11px] text-[var(--text-muted)]">{card.subtexto}</p>
            </div>
          </div>
        )
      })}
    </section>
  )
}
