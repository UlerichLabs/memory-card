import { Star } from 'lucide-react'
import type { DashboardResumo } from '@/types/dashboard'
import { formatarData, formatarHoras, iniciais } from '@/lib/dashboardUtils'

interface PerfilJogadorProps {
  nome: string
  resumo: DashboardResumo
  totalAbandonados: number
}

export function PerfilJogador({ nome, resumo, totalAbandonados }: PerfilJogadorProps) {
  const jornada = resumo.anos_desde_primeiro > 0 ? `${resumo.anos_desde_primeiro} anos` : `${resumo.dias_desde_primeiro} dias`
  const anoInicio = resumo.primeiro_zeramento_em?.slice(0, 4)
  const metricas = [
    { label: 'Jogos zerados', value: resumo.total_jogos, sub: typeof resumo.jogos_no_ano_atual === 'number' ? `${resumo.jogos_no_ano_atual} em ${new Date().getFullYear()}` : undefined },
    { label: 'Horas jogadas', value: formatarHoras(resumo.total_segundos), sub: typeof resumo.media_segundos_por_jogo === 'number' ? `Média de ${formatarHoras(resumo.media_segundos_por_jogo)} por jogo` : undefined },
    { label: 'Nota média', value: resumo.nota_media.toFixed(1), sub: 'Avaliação geral', nota: true },
    { label: 'Abandonados', value: totalAbandonados, sub: 'Descontinuados', abandoned: true },
    { label: 'Jornada', value: jornada, sub: resumo.primeiro_zeramento_em ? `Desde ${formatarData(resumo.primeiro_zeramento_em)}` : undefined },
  ]

  return (
    <section aria-label="Perfil do jogador" className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4 sm:p-5">
      <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-3">
          <div className="flex size-[60px] shrink-0 items-center justify-center rounded-full bg-[var(--accent)] text-xl font-bold text-[var(--accent-foreground)] sm:size-[72px] sm:text-2xl">{iniciais(nome)}</div>
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-wide text-[var(--text-muted)]">Perfil do jogador</p>
            <h1 className="text-[22px] font-bold text-[var(--text-primary)]">{nome}</h1>
            {anoInicio && <p className="text-xs text-[var(--text-secondary)]">Jogando desde {anoInicio}</p>}
          </div>
        </div>
        <div className="grid grid-cols-2 gap-y-5 border-t border-[var(--border)] pt-5 sm:grid-cols-5 lg:flex lg:flex-1 lg:justify-end lg:border-t-0 lg:pt-0">
          {metricas.map(({ label, value, sub, nota, abandoned }, index) => <div key={label} className={`flex min-w-0 flex-col gap-0.5 px-3 lg:min-w-[112px] lg:border-l lg:border-[var(--border)] ${index === 0 ? 'lg:border-l-0' : ''}`}><span className="text-xs text-[var(--text-muted)]">{label}</span><strong className={`flex items-center gap-1 text-[26px] font-bold leading-tight tracking-tight tabular-nums lg:text-[28px] ${abandoned ? 'text-[var(--abandonado-text)]' : 'text-[var(--text-primary)]'}`}>{nota && <Star className="size-4 fill-current text-[var(--highlight-gold)]" aria-hidden="true" />}{value}</strong>{sub && <span className="text-[11px] text-[var(--text-muted)]">{sub}</span>}</div>)}
        </div>
      </div>
    </section>
  )
}
