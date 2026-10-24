import { Pause, Pencil, Trash2 } from 'lucide-react'
import type { JogoAbandonado } from '@/types/abandonados'
import { formatarCapaIGDB } from '@/lib/utils'
import { obterIniciaisJogo } from '@/components/listas/listas.utils'
import { formatarDataBrasileira, formatarTempoAbandonado } from './abandonados.utils'
import { ConsoleBadge } from '@/components/jogos/ConsoleBadge'

export interface AbandonadoCardProps {
  jogo: JogoAbandonado
  onRetomar: (jogo: JogoAbandonado) => void
  onEditar: (jogo: JogoAbandonado) => void
  onExcluir: (jogo: JogoAbandonado) => void
}

export function AbandonadoCard({ jogo, onRetomar, onEditar, onExcluir }: AbandonadoCardProps) {
  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url ?? undefined, 't_cover_big')
  const iniciais = obterIniciaisJogo(jogo.nome)
  const tempoStr = formatarTempoAbandonado(jogo.tempo_jogado)
  const dataPt = formatarDataBrasileira(jogo.abandonado_em)

  return (
    <article className="group flex min-w-0 flex-col justify-between overflow-hidden rounded-xl border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] p-3 transition duration-150 hover:-translate-y-0.5 hover:border-[var(--biblioteca-control-border-hover)] sm:p-4">
      <div>
        <div className="relative aspect-[3/4] w-full overflow-hidden rounded-[12px] border border-[var(--biblioteca-card-cover-border)] bg-[var(--abandonado-cover)]">
          {capaUrl ? (
            <img src={capaUrl} alt={jogo.nome} className="h-full w-full object-cover transition duration-150 group-hover:opacity-90" loading="lazy" />
          ) : (
            <div className="flex h-full w-full items-center justify-center p-3 text-center text-[var(--abandonado-text)]">
              <span className="select-none text-xl font-bold">{iniciais}</span>
            </div>
          )}

          <div className="absolute left-1.5 top-1.5 z-10">
            <span className="inline-flex shrink-0 select-none items-center justify-center rounded-full border border-[var(--border-subtle)] bg-[var(--dif-pill-sobre-capa-bg)] px-1.5 py-0.5 text-[10px] font-semibold text-[var(--text-primary)] backdrop-blur-xs sm:px-2 sm:py-1 sm:text-[11px]">
              {tempoStr}
            </span>
          </div>

          <div
            aria-label="Jogo abandonado"
            className="absolute right-1.5 top-1.5 z-10 flex h-7 w-7 min-h-[28px] min-w-[28px] items-center justify-center rounded-[8px] border border-[var(--abandonado-border)] bg-[var(--abandonado-bg)] text-[var(--abandonado-text)] sm:h-8 sm:w-8 sm:min-h-[32px] sm:min-w-[32px]"
          >
            <Pause className="h-3.5 w-3.5 fill-current" />
          </div>
        </div>

        <div className="flex flex-col gap-0.5 pt-2.5">
          <h3
            className="line-clamp-2 min-h-[38px] text-[14px] font-semibold leading-[1.35] text-[var(--biblioteca-text-primary)]"
            title={jogo.nome}
          >
            {jogo.nome}
          </h3>
          <ConsoleBadge nome={jogo.console} />
          <p className="line-clamp-2 min-h-[32px] text-[12px] leading-tight tabular-nums text-[var(--biblioteca-card-meta)]">
            Abandonado em {dataPt}
          </p>
          <p
            className="mt-1 line-clamp-2 min-h-[32px] text-[12px] italic text-[var(--text-muted)]"
            title={jogo.motivo || undefined}
          >
            {jogo.motivo ? `"${jogo.motivo}"` : 'Sem motivo registrado'}
          </p>
        </div>
      </div>

      <div className="mt-4 flex flex-col gap-2 border-t border-[var(--border-subtle)] pt-3">
        <button
          type="button"
          onClick={() => onRetomar(jogo)}
          aria-label={`Retomar ${jogo.nome}`}
          className="flex min-h-[44px] w-full items-center justify-center rounded-lg border border-[var(--abandonado-border)] bg-[var(--abandonado-bg)] px-3 text-xs font-semibold text-[var(--abandonado-text)] transition hover:opacity-90"
        >
          <span>Retomar</span>
        </button>
        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={() => onEditar(jogo)}
            aria-label={`Editar ${jogo.nome}`}
            title="Editar"
            className="flex min-h-[44px] items-center justify-center gap-1.5 rounded-lg border border-[var(--border-subtle)] px-2 text-xs font-medium text-[var(--text-secondary)] transition hover:text-[var(--text-primary)]"
          >
            <Pencil className="size-3.5 shrink-0" aria-hidden="true" />
            <span>Editar</span>
          </button>
          <button
            type="button"
            onClick={() => onExcluir(jogo)}
            aria-label={`Excluir ${jogo.nome}`}
            title="Excluir"
            className="flex min-h-[44px] items-center justify-center gap-1.5 rounded-lg border border-[var(--border-subtle)] px-2 text-xs font-medium text-[var(--text-secondary)] transition hover:text-[var(--danger)]"
          >
            <Trash2 className="size-3.5 shrink-0" aria-hidden="true" />
            <span>Excluir</span>
          </button>
        </div>
      </div>
    </article>
  )
}
