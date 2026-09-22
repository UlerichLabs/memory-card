import { Star, Clock, Calendar, Pencil, Trash2, Gamepad2, Award } from 'lucide-react'
import type { JogoZeradoDTO, Dificuldade } from '@/lib/services/jogosService'

export interface BibliotecaCardProps {
  jogo: JogoZeradoDTO
  onEditar: (jogo: JogoZeradoDTO) => void
  onExcluir: (jogo: JogoZeradoDTO) => void
}

const dificuldadeCores: Record<Dificuldade, string> = {
  C: 'text-[#52545C] border-[#52545C]/40',
  B: 'text-[#6B7280] border-[#6B7280]/40',
  A: 'text-[#4F7CFF] border-[#4F7CFF]/40',
  AA: 'text-[#E8C15C] border-[#E8C15C]/40',
  AAA: 'text-[#E05A4E] border-[#E05A4E]/40',
}

function formatarTempo(segundos: number): string {
  const h = Math.floor(segundos / 3600)
  const m = Math.floor((segundos % 3600) / 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  if (m > 0) return `${m}m`
  return `${segundos}s`
}

function formatarData(dataStr: string): string {
  const partes = dataStr.slice(0, 10).split('-')
  if (partes.length === 3) {
    return `${partes[2]}/${partes[1]}/${partes[0]}`
  }
  return dataStr
}

export function BibliotecaCard({ jogo, onEditar, onExcluir }: BibliotecaCardProps) {
  const corDificuldade = dificuldadeCores[jogo.dificuldade] || 'text-[var(--text-secondary)] border-[var(--border-subtle)]'

  return (
    <article className="group flex flex-col rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-3.5 transition-transform duration-200 hover:-translate-y-1">
      <div className="relative aspect-[3/4] w-full overflow-hidden rounded-[6px] border border-[var(--border)] bg-[var(--bg-surface-alt)]">
        {jogo.igdb_capa_url ? (
          <img
            src={jogo.igdb_capa_url}
            alt={jogo.nome}
            className="h-full w-full object-cover transition-opacity duration-200 group-hover:opacity-90"
            loading="lazy"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-[#1D1F25] to-[#2A2C33] text-[var(--text-muted)]">
            <Gamepad2 className="h-10 w-10 opacity-40" aria-hidden="true" />
          </div>
        )}

        <div className={`absolute left-1.5 top-1.5 rounded-full border bg-[var(--bg-surface-alt)]/90 px-2 py-0.5 text-[10px] font-bold ${corDificuldade}`}>
          Dif. {jogo.dificuldade}
        </div>

        <div className="absolute right-1.5 top-1.5 flex items-center gap-1 rounded bg-black/80 px-1.5 py-0.5 text-[11px] font-bold text-[var(--highlight-gold)]">
          <Star className="h-3 w-3 fill-[var(--highlight-gold)]" aria-hidden="true" />
          <span>{jogo.nota}</span>
        </div>

        {jogo.destaque && (
          <div className="absolute bottom-1.5 left-1.5 flex items-center gap-1 rounded bg-[var(--highlight-gold)]/20 border border-[var(--highlight-gold)]/40 px-1.5 py-0.5 text-[10px] font-bold text-[var(--highlight-gold)]">
            <Award className="h-3 w-3" aria-hidden="true" />
            <span>Destaque</span>
          </div>
        )}
      </div>

      <div className="mt-3 flex flex-1 flex-col justify-between space-y-2">
        <div>
          <h3 className="truncate text-[14px] font-bold text-[var(--text-primary)]" title={jogo.nome}>
            {jogo.nome}
          </h3>
          <div className="mt-1 flex flex-wrap items-center gap-1.5 text-[11px] text-[var(--text-secondary)]">
            <span className="rounded-[20px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2 py-0.5 font-medium">
              {jogo.console}
            </span>
            {jogo.genero && (
              <span className="rounded-[20px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2 py-0.5 text-[var(--text-muted)]">
                {jogo.genero}
              </span>
            )}
          </div>
        </div>

        <div className="space-y-1 border-t border-[var(--border)] pt-2 text-[11px] text-[var(--text-muted)]">
          {jogo.tempo_jogado > 0 && (
            <div className="flex items-center gap-1">
              <Clock className="h-3 w-3" aria-hidden="true" />
              <span>{formatarTempo(jogo.tempo_jogado)} jogados</span>
            </div>
          )}
          <div className="flex items-center gap-1">
            <Calendar className="h-3 w-3" aria-hidden="true" />
            <span>Finalizado em {formatarData(jogo.finalizado_em)}</span>
          </div>
          {jogo.condicao_zeramento && (
            <p className="line-clamp-2 text-[10.5px] italic text-[var(--text-secondary)]" title={jogo.condicao_zeramento}>
              "{jogo.condicao_zeramento}"
            </p>
          )}
        </div>

        <div className="flex items-center gap-2 pt-1">
          <button
            type="button"
            aria-label={`Editar ${jogo.nome}`}
            onClick={() => onEditar(jogo)}
            className="flex flex-1 items-center justify-center gap-1 rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2.5 py-1.5 text-[12px] font-medium text-[var(--text-primary)] transition-colors hover:border-[var(--accent)] hover:text-[var(--accent)]"
          >
            <Pencil className="h-3.5 w-3.5" aria-hidden="true" />
            <span>Editar</span>
          </button>
          <button
            type="button"
            aria-label={`Excluir ${jogo.nome}`}
            onClick={() => onExcluir(jogo)}
            className="flex items-center justify-center rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] p-1.5 text-[var(--text-muted)] transition-colors hover:border-[var(--danger)] hover:bg-[var(--danger)]/10 hover:text-[var(--danger)]"
          >
            <Trash2 className="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        </div>
      </div>
    </article>
  )
}
