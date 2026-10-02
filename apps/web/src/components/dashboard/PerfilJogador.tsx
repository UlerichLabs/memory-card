import { Clock, CircleSlash, Gamepad2, Star, Trophy } from 'lucide-react'
import type { DashboardResumo } from '@/types/dashboard'
import { formatarHoras, iniciais } from '@/lib/dashboardUtils'

interface PerfilJogadorProps {
  nome: string
  resumo: DashboardResumo
  totalAbandonados: number
}

export function PerfilJogador({ nome, resumo, totalAbandonados }: PerfilJogadorProps) {
  const jornada = resumo.anos_desde_primeiro > 0 ? `${resumo.anos_desde_primeiro} anos` : `${resumo.dias_desde_primeiro} dias`
  const anoInicio = resumo.primeiro_zeramento_em?.slice(0, 4)
  const metricas = [
    { label: 'Jogos zerados', value: resumo.total_jogos, icon: Trophy },
    { label: 'Horas jogadas', value: formatarHoras(resumo.total_segundos), icon: Clock },
    { label: 'Nota média', value: resumo.nota_media.toFixed(1), icon: Star },
    { label: 'Abandonados', value: totalAbandonados, icon: CircleSlash, abandoned: true },
    { label: 'Jornada', value: jornada, icon: Gamepad2 },
  ]

  return (
    <section aria-label="Perfil do jogador" className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4 sm:p-5">
      <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-3">
          <div className="flex size-12 shrink-0 items-center justify-center rounded-full bg-[var(--accent)] text-base font-bold text-[var(--accent-foreground)]">{iniciais(nome)}</div>
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-wide text-[var(--text-muted)]">Perfil do jogador</p>
            <h1 className="text-lg font-bold text-[var(--text-primary)]">{nome}</h1>
            {anoInicio && <p className="text-xs text-[var(--text-secondary)]">Jogando desde {anoInicio}</p>}
          </div>
        </div>
        <div className="grid grid-cols-2 gap-y-4 sm:grid-cols-5 lg:flex lg:flex-1 lg:justify-end">
          {metricas.map(({ label, value, icon: Icon, abandoned }, index) => <div key={label} className={`flex items-center gap-2 px-3 lg:min-w-[112px] lg:flex-col lg:items-start lg:gap-1 lg:border-l lg:border-[var(--border)] ${index === 0 ? 'lg:border-l-0' : ''}`}><Icon className={`size-4 ${abandoned ? 'text-[var(--abandonado-text)]' : 'text-[var(--text-faint)]'}`} aria-hidden="true" /><span className="text-[11px] text-[var(--text-secondary)]">{label}</span><strong className={`text-sm font-bold ${abandoned ? 'text-[var(--abandonado-text)]' : 'text-[var(--text-primary)]'}`}>{value}</strong></div>)}
        </div>
      </div>
    </section>
  )
}
