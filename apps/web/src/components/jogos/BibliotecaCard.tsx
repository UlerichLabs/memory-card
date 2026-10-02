import { useState, useRef, useEffect } from 'react'
import { Crown, MoreVertical, Pencil, Trash2, Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB, isoParaDataPt } from '@/lib/utils'
import { NotaBadge } from './NotaBadge'
import { DificuldadePill } from './DificuldadePill'
import { ConsoleBadge } from './ConsoleBadge'

export interface BibliotecaCardProps {
  jogo: JogoZeradoDTO
  onEditar: (jogo: JogoZeradoDTO) => void
  onExcluir: (jogo: JogoZeradoDTO) => void
  onDetalhes?: (jogo: JogoZeradoDTO) => void
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

  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url, 't_cover_big')
  const dataPt = isoParaDataPt(jogo.finalizado_em)
  const tempoStr = jogo.tempo_jogado > 0 ? formatarTempo(jogo.tempo_jogado) : ''
  const metaTexto = dataPt && tempoStr ? `${dataPt} · ${tempoStr}` : dataPt || tempoStr

  return (
    <article
      onClick={() => onDetalhes?.(jogo)}
      className="group flex min-w-0 cursor-pointer flex-col overflow-hidden rounded-xl border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] p-3 transition duration-150 hover:-translate-y-0.5 hover:border-[var(--biblioteca-control-border-hover)] sm:p-4"
    >
      <div className="relative aspect-[3/4] w-full overflow-hidden rounded-[12px] border border-[var(--biblioteca-card-cover-border)] bg-[var(--biblioteca-card-cover-bg)]">
        {capaUrl ? (
          <img src={capaUrl} alt={jogo.nome} className="h-full w-full object-cover transition duration-150 group-hover:opacity-90" loading="lazy" />
        ) : (
          <div className="flex h-full w-full flex-col items-center justify-center p-3 text-center text-[var(--biblioteca-control-placeholder)]">
            <Gamepad2 className="h-8 w-8 opacity-40" aria-hidden="true" />
            <span className="mt-2 line-clamp-2 text-xs font-semibold text-[var(--biblioteca-text-secondary)]">{jogo.nome}</span>
          </div>
        )}

        <div className="absolute left-1.5 top-1.5 z-10">
          <DificuldadePill nivel={jogo.dificuldade} variante="sobreCapa" className="text-[10px] py-0.5 px-1.5 sm:text-[11px] sm:px-2 sm:py-1" />
        </div>

        <div className="absolute right-1.5 top-1.5 z-10">
          <NotaBadge nota={jogo.nota} tamanho="sm" className="min-w-[28px] min-h-[28px] h-7 text-[13px] sm:min-w-[32px] sm:min-h-[32px] sm:h-8 sm:text-[14px]" />
        </div>

        {jogo.destaque && (
          <div className="absolute bottom-1.5 left-1.5 flex items-center gap-1 rounded-[4px] bg-[var(--ouro-jogo-ano)] px-2 py-1 text-[11px] font-bold text-[var(--ouro-jogo-ano-text)]">
            <Crown className="h-3 w-3" aria-hidden="true" />
            <span>Jogo do ano</span>
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

      <div className="flex min-w-0 flex-col gap-0.5 pt-2.5">
        <h3 className="line-clamp-2 min-h-[38px] text-[14px] font-semibold leading-[1.35] text-[var(--biblioteca-text-primary)]" title={jogo.nome}>
          {jogo.nome}
        </h3>
        <ConsoleBadge nome={jogo.console} variante="solido" />
        {metaTexto && (
          <p className="truncate text-[12px] tabular-nums text-[var(--biblioteca-card-meta)]">
            {metaTexto}
          </p>
        )}
      </div>
    </article>
  )
}
