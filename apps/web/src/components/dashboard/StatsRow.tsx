import { CircleSlash, Clock, Star, Trophy } from 'lucide-react'
import type { DashboardResumo } from '@/types/dashboard'
import { formatarData, formatarHoras } from '@/lib/dashboardUtils'

interface StatsRowProps { resumo: DashboardResumo; totalAbandonados: number }

export function StatsRow({ resumo, totalAbandonados }: StatsRowProps) {
  const ano = new Date().getFullYear()
  const media = formatarHoras(resumo.media_segundos_por_jogo)
  const jornada = resumo.anos_desde_primeiro > 0 ? `${resumo.anos_desde_primeiro} anos` : `${resumo.dias_desde_primeiro} dias`
  const subJornada = resumo.primeiro_zeramento_em
    ? `${resumo.dias_desde_primeiro.toLocaleString('pt-BR')} dias desde ${formatarData(resumo.primeiro_zeramento_em)}`
    : `${resumo.dias_desde_primeiro.toLocaleString('pt-BR')} dias`
  const cards = [
    { label: 'Jogos zerados', value: resumo.total_jogos, sub: `${resumo.jogos_no_ano_atual} em ${ano}`, icon: Trophy },
    { label: 'Horas jogadas', value: formatarHoras(resumo.total_segundos), sub: `Média de ${media} por jogo`, icon: Clock },
    { label: 'Jornada', value: jornada, sub: subJornada, icon: Clock },
    { label: 'Abandonados', value: totalAbandonados, sub: 'Descontinuados', icon: CircleSlash, abandoned: true },
    { label: 'Nota média', value: resumo.nota_media.toFixed(1), sub: 'Avaliação geral', icon: Star },
  ]
  return <section aria-label="Estatísticas Gerais" className="grid grid-cols-2 gap-3 lg:grid-cols-5">
    {cards.map(({ label, value, sub, icon: Icon, abandoned }) => <article key={label} className={`rounded-[10px] border p-4 ${abandoned ? 'border-[var(--abandonado-border)] bg-[var(--abandonado-bg)]' : 'border-[var(--border)] bg-[var(--bg-surface)]'}`}>
      <div className="flex items-center justify-between"><span className="text-[13px] font-medium text-[var(--text-secondary)]">{label}</span><Icon className={`h-4 w-4 ${abandoned ? 'text-[var(--abandonado-text)]' : 'text-[var(--text-faint)]'}`} aria-hidden="true" /></div>
      <p className="mt-3 text-2xl font-bold text-[var(--text-primary)]">{value}</p><p className="mt-0.5 text-[11px] text-[var(--text-muted)]">{sub}</p>
    </article>)}
  </section>
}
