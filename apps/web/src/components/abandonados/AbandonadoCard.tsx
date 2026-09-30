import { Pause, Pencil, Trash2 } from 'lucide-react'
import type { JogoAbandonado } from '@/types/abandonados'
import { formatarCapaIGDB } from '@/lib/utils'
import { obterIniciaisJogo } from '@/components/listas/listas.utils'
import { formatarDataBrasileira, formatarTempoAbandonado } from './abandonados.utils'
import { useJogosStore } from '@/stores/jogosStore'
import { useAbandonadosStore } from '@/stores/abandonadosStore'

export interface AbandonadoCardProps {
  jogo: JogoAbandonado
  onEditar: (jogo: JogoAbandonado) => void
  onExcluir: (jogo: JogoAbandonado) => void
}

export function AbandonadoCard({ jogo, onEditar, onExcluir }: AbandonadoCardProps) {
  const { abrirModalRegistro } = useJogosStore()
  const { excluirJogo } = useAbandonadosStore()

  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url ?? undefined, 't_cover_big')
  const iniciais = obterIniciaisJogo(jogo.nome)
  const tempoStr = formatarTempoAbandonado(jogo.tempo_jogado)
  const dataPt = formatarDataBrasileira(jogo.abandonado_em)

  function handleRetomar() {
    abrirModalRegistro({
      valoresIniciais: {
        nome: jogo.nome, console: jogo.console, igdb_id: jogo.igdb_id ?? undefined,
        igdb_capa_url: jogo.igdb_capa_url ?? undefined, tempo_jogado: jogo.tempo_jogado,
      },
      aviso: `Retomando um jogo abandonado em ${dataPt}. Ao salvar o zeramento, ele sai da lista de Abandonados.`,
      textoSubmit: 'Salvar zeramento',
      onSalvo: async () => {
        try { await excluirJogo(jogo.id) }
        catch { alert('Jogo registrado, mas não foi possível remover da lista de abandonados.') }
      },
    })
  }

  return (
    <article
      className={
        'group flex flex-col justify-between overflow-hidden rounded-xl border ' +
        'border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] p-3 ' +
        'transition duration-150 hover:-translate-y-0.5'
      }
    >
      <div>
        <div
          className={
            'relative aspect-[3/4] w-full overflow-hidden rounded-[10px] ' +
            'border border-[var(--biblioteca-card-cover-border)] bg-[var(--abandonado-cover)]'
          }
        >
          {capaUrl ? (
            <img src={capaUrl} alt={jogo.nome} className="h-full w-full object-cover" loading="lazy" />
          ) : (
            <div className="flex h-full w-full items-center justify-center p-3 text-center">
              <span className="text-xl font-bold text-[var(--abandonado-text)] select-none">{iniciais}</span>
            </div>
          )}

          <div
            className={
              'absolute left-1.5 top-1.5 z-10 rounded-md bg-[var(--biblioteca-pill-bg)] ' +
              'px-2 py-0.5 text-[11px] font-semibold text-[var(--text-primary)] backdrop-blur-xs'
            }
          >
            {tempoStr}
          </div>

          <div
            aria-label="Jogo abandonado"
            className={
              'absolute right-1.5 top-1.5 z-10 flex size-7 items-center justify-center ' +
              'rounded-md border border-[var(--abandonado-border)] bg-[var(--abandonado-bg)] ' +
              'text-[var(--abandonado-text)]'
            }
          >
            <Pause className="size-3.5 fill-current" />
          </div>
        </div>

        <div className="flex flex-col gap-0.5 pt-2.5">
          <h3
            className={
              'line-clamp-2 min-h-[38px] text-[14px] font-semibold ' +
              'leading-[1.35] text-[var(--biblioteca-text-primary)]'
            }
            title={jogo.nome}
          >
            {jogo.nome}
          </h3>
          <p className="truncate text-[12px] text-[var(--biblioteca-card-plataforma)]">
            {jogo.console}
          </p>
          <p className="truncate text-[12px] tabular-nums text-[var(--biblioteca-card-meta)]">
            Abandonado em {dataPt}
          </p>
          <p
            className="mt-1 line-clamp-2 min-h-[32px] text-[12px] text-[var(--text-muted)] italic"
            title={jogo.motivo || undefined}
          >
            {jogo.motivo ? `"${jogo.motivo}"` : 'Sem motivo registrado'}
          </p>
        </div>
      </div>

      <div className="mt-3 flex items-center gap-1.5 border-t border-[var(--border-subtle)] pt-2.5">
        <button
          type="button"
          onClick={handleRetomar}
          aria-label={`Retomar ${jogo.nome}`}
          className={
            'flex min-h-[44px] flex-1 items-center justify-center rounded-lg border ' +
            'border-[var(--abandonado-border)] bg-[var(--abandonado-bg)] px-2.5 text-xs ' +
            'font-semibold text-[var(--abandonado-text)] transition hover:opacity-90'
          }
        >
          Retomar
        </button>
        <button
          type="button"
          onClick={() => onEditar(jogo)}
          aria-label={`Editar ${jogo.nome}`}
          title="Editar"
          className={
            'flex min-h-[44px] min-w-[44px] items-center justify-center rounded-lg ' +
            'border border-[var(--border-subtle)] text-[var(--text-secondary)] ' +
            'transition hover:text-[var(--text-primary)]'
          }
        >
          <Pencil className="size-4" />
        </button>
        <button
          type="button"
          onClick={() => onExcluir(jogo)}
          aria-label={`Excluir ${jogo.nome}`}
          title="Excluir"
          className={
            'flex min-h-[44px] min-w-[44px] items-center justify-center rounded-lg ' +
            'border border-[var(--border-subtle)] text-[var(--text-secondary)] ' +
            'transition hover:text-[var(--danger)]'
          }
        >
          <Trash2 className="size-4" />
        </button>
      </div>
    </article>
  )
}
