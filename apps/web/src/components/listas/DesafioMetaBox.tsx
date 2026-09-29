import { Trophy } from 'lucide-react'
import { format, parseISO } from 'date-fns'
import { ROTULOS_METABOX } from './listas.constants'
import { formatarFaltam } from './listas.utils'
import type { ListaDetalhada } from '@/types/listas'

export interface DesafioMetaBoxProps {
  lista: ListaDetalhada
}

export function DesafioMetaBox({ lista }: DesafioMetaBoxProps) {
  const progresso = lista.progresso
  if (!progresso) return null

  const { feitos, meta, percentual, concluido, concluido_em } = progresso
  const regraTipo = lista.regra?.tipo ?? 'manual'
  const rotulo = ROTULOS_METABOX[regraTipo] ?? 'jogos'
  const faltam = Math.max(0, meta - feitos)

  let dataConclusao = ''
  if (concluido && concluido_em) {
    try {
      dataConclusao = format(parseISO(concluido_em), 'dd/MM/yyyy')
    } catch {
      dataConclusao = concluido_em.slice(0, 10)
    }
  }

  return (
    <div
      className={`flex flex-col gap-3.5 rounded-[14px] p-5 sm:p-[20px_22px] transition-colors ${
        concluido
          ? 'border border-[var(--lista-metabox-concluido-border)] bg-[var(--lista-metabox-concluido-bg)]'
          : 'border border-[var(--lista-metabox-border)] bg-[var(--lista-metabox-bg)]'
      }`}
    >
      <div className="flex flex-wrap items-baseline gap-2.5">
        <span
          className={`text-[36px] sm:text-[44px] font-extrabold tracking-[-0.02em] leading-none tabular-nums ${
            concluido ? 'text-[var(--hall-ouro)]' : 'text-[var(--lista-text-bright)]'
          }`}
        >
          {feitos}
        </span>
        <span className="text-[20px] font-semibold text-[var(--lista-text-dimmer)]">
          / {meta}
        </span>
        <span className="ml-2 text-[14px] text-[var(--lista-text-secondary)]">
          {rotulo}
        </span>
        <span className="ml-auto text-[22px] font-bold text-[var(--lista-text-light)]">
          {percentual}%
        </span>
      </div>

      <div
        role="progressbar"
        aria-valuenow={percentual}
        aria-valuemin={0}
        aria-valuemax={100}
        className="h-2.5 w-full overflow-hidden rounded-[5px] bg-[var(--lista-progress-track)]"
      >
        <div
          style={{ width: `${Math.min(100, Math.max(0, percentual))}%` }}
          className={`h-full transition-all duration-300 ${
            concluido ? 'bg-[var(--hall-ouro)]' : 'bg-[var(--lista-progress-fill)]'
          }`}
        />
      </div>

      <div className="flex items-center gap-2">
        {concluido ? (
          <>
            <Trophy className="h-[18px] w-[18px] text-[var(--hall-ouro)] shrink-0" aria-hidden="true" />
            <span className="text-[14px] font-bold text-[var(--hall-ouro)]">
              Desafio concluído em {dataConclusao} — conquista desbloqueada
            </span>
          </>
        ) : (
          <span className="text-[13px] text-[var(--lista-text-muted)]">
            {formatarFaltam(faltam)}
          </span>
        )}
      </div>
    </div>
  )
}
