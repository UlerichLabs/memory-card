import { useState, useRef, useEffect } from 'react'
import { Star, Award, MoreVertical, Pencil, Trash2, Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO, Dificuldade } from '@/types/jogos'
import { formatarCapaIGDB } from '@/lib/utils'

export interface BibliotecaCardProps {
  jogo: JogoZeradoDTO
  onEditar: (jogo: JogoZeradoDTO) => void
  onExcluir: (jogo: JogoZeradoDTO) => void
  onDetalhes?: (jogo: JogoZeradoDTO) => void
}

const DIFICULDADE_VAR: Record<Dificuldade, string> = {
  C: 'var(--difficulty-c)', B: 'var(--difficulty-b)', A: 'var(--difficulty-a)',
  AA: 'var(--difficulty-aa)', AAA: 'var(--difficulty-aaa)',
}

function formatarTempo(s: number): string {
  const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  return m > 0 ? `${m}m` : `${s}s`
}

export function BibliotecaCard({ jogo, onEditar, onExcluir, onDetalhes }: BibliotecaCardProps) {
  const [menuAberto, setMenuAberto] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickFora(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuAberto(false)
    }
    if (menuAberto) {
      document.addEventListener('mousedown', handleClickFora)
      return () => document.removeEventListener('mousedown', handleClickFora)
    }
  }, [menuAberto])

  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url)
  const ano = jogo.finalizado_em ? jogo.finalizado_em.slice(0, 4) : ''
  const tempoStr = jogo.tempo_jogado > 0 ? formatarTempo(jogo.tempo_jogado) : ''
  const metaTexto = ano && tempoStr ? `${ano} · ${tempoStr}` : ano || tempoStr
  const difCor = DIFICULDADE_VAR[jogo.dificuldade] || 'var(--biblioteca-text-muted)'

  return (
    <article
      onClick={() => onDetalhes?.(jogo)}
      className="group flex cursor-pointer flex-col overflow-hidden rounded-[10px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] transition duration-150 hover:-translate-y-0.5 hover:border-[var(--biblioteca-card-border-hover)]"
    >
      <div className="relative aspect-[3/4] w-full overflow-hidden bg-[var(--biblioteca-card-cover-bg)]">
        {capaUrl ? (
          <img src={capaUrl} alt={jogo.nome} className="h-full w-full object-cover transition duration-150 group-hover:opacity-90" loading="lazy" />
        ) : (
          <div className="flex h-full w-full flex-col items-center justify-center p-3 text-center text-[var(--biblioteca-control-placeholder)]">
            <Gamepad2 className="h-8 w-8 opacity-40" aria-hidden="true" />
            <span className="mt-2 line-clamp-2 text-xs font-semibold text-[var(--biblioteca-text-secondary)]">{jogo.nome}</span>
          </div>
        )}

        <div className="absolute left-1.5 top-1.5 max-w-[65%] truncate rounded-[4px] bg-[var(--biblioteca-pill-bg)] px-1.5 py-0.5 text-[10px] font-medium text-[var(--biblioteca-control-text)] backdrop-blur-xs">
          {jogo.console}
        </div>

        <div className="absolute right-1.5 top-1.5 flex items-center gap-1 rounded-[4px] bg-black/80 px-1.5 py-0.5 text-xs font-bold text-[var(--biblioteca-gold)]">
          <Star className="h-3 w-3 fill-[var(--biblioteca-gold)]" aria-hidden="true" />
          <span>{jogo.nota}</span>
        </div>

        {jogo.destaque && (
          <div className="absolute bottom-1.5 left-1.5 flex items-center gap-0.5 rounded-[4px] bg-[var(--biblioteca-gold)]/20 border border-[var(--biblioteca-gold)]/40 px-1.5 py-0.5 text-[10px] font-bold text-[var(--biblioteca-gold)]">
            <Award className="h-3 w-3" aria-hidden="true" />
            <span>Destaque</span>
          </div>
        )}

        <div ref={menuRef} onClick={(e) => e.stopPropagation()} className="absolute bottom-1.5 right-1.5 z-10">
          <button
            type="button"
            onClick={() => setMenuAberto((prev) => !prev)}
            aria-label={`Opções de ${jogo.nome}`}
            className="flex h-7 w-7 items-center justify-center rounded-[4px] bg-black/60 text-[var(--biblioteca-control-text)] backdrop-blur-xs transition hover:bg-black/80"
          >
            <MoreVertical className="h-3.5 w-3.5" aria-hidden="true" />
          </button>

          {menuAberto && (
            <div className="absolute bottom-full right-0 mb-1 min-w-[120px] rounded-[6px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-control-bg)] p-1 shadow-lg">
              <button
                type="button"
                onClick={() => { setMenuAberto(false); onEditar(jogo) }}
                aria-label={`Editar ${jogo.nome}`}
                className="flex w-full items-center gap-2 rounded-[4px] px-2 py-1.5 text-xs font-medium text-[var(--biblioteca-text-primary)] hover:bg-[var(--biblioteca-card-bg)]"
              >
                <Pencil className="h-3.5 w-3.5" aria-hidden="true" />
                <span>Editar</span>
              </button>
              <button
                type="button"
                onClick={() => { setMenuAberto(false); onExcluir(jogo) }}
                aria-label={`Excluir ${jogo.nome}`}
                className="flex w-full items-center gap-2 rounded-[4px] px-2 py-1.5 text-xs font-medium text-[var(--danger)] hover:bg-[var(--biblioteca-card-bg)]"
              >
                <Trash2 className="h-3.5 w-3.5" aria-hidden="true" />
                <span>Excluir</span>
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="flex flex-1 flex-col justify-between p-3">
        <div>
          <h3 className="truncate text-sm font-bold text-[var(--biblioteca-text-primary)]" title={jogo.nome}>{jogo.nome}</h3>
          {metaTexto && <p className="mt-0.5 text-[11px] text-[var(--biblioteca-text-muted)]">{metaTexto}</p>}
        </div>
        <div className="mt-2.5 flex items-center justify-between">
          <span style={{ borderColor: difCor, color: difCor }} className="rounded-[4px] border px-1.5 py-0.5 text-[10px] font-bold">
            Dif. {jogo.dificuldade}
          </span>
        </div>
      </div>
    </article>
  )
}
