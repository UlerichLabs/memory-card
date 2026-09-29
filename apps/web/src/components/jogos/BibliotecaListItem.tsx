import { useState, useRef, useEffect } from 'react'
import { MoreVertical, Pencil, Trash2, Gamepad2, Crown } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB, isoParaDataPt } from '@/lib/utils'
import { NotaBadge } from './NotaBadge'
import { DificuldadePill } from './DificuldadePill'

export interface BibliotecaListItemProps {
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

export function BibliotecaListItem({ jogo, onEditar, onExcluir, onDetalhes }: BibliotecaListItemProps) {
  const [menuAberto, setMenuAberto] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickFora(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuAberto(false)
      }
    }
    if (menuAberto) {
      document.addEventListener('mousedown', handleClickFora)
      return () => document.removeEventListener('mousedown', handleClickFora)
    }
  }, [menuAberto])

  const tempoFormatado = jogo.tempo_jogado > 0 ? formatarTempo(jogo.tempo_jogado) : ''
  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url, 't_cover_big')

  return (
    <div
      onClick={() => onDetalhes?.(jogo)}
      className="group flex cursor-pointer items-center justify-between gap-3 rounded-[8px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] px-3 py-2.5 transition-colors hover:bg-[var(--biblioteca-row-hover)]"
    >
      <div className="flex min-w-0 flex-1 items-center gap-3">
        {capaUrl ? (
          <img
            src={capaUrl}
            alt=""
            className="h-[58px] w-[44px] shrink-0 rounded-[4px] object-cover border border-[var(--biblioteca-card-cover-border)]"
          />
        ) : (
          <div className="flex h-[58px] w-[44px] shrink-0 items-center justify-center rounded-[4px] border border-[var(--biblioteca-card-cover-border)] bg-[var(--biblioteca-card-cover-bg)] text-[var(--detalhe-capa-icon-sem-arte)]">
            <Gamepad2 className="h-5 w-5" aria-hidden="true" />
          </div>
        )}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="truncate text-[13px] font-bold text-[var(--biblioteca-text-primary)]">
              {jogo.nome}
            </span>
            {jogo.destaque && (
              <span className="inline-flex items-center gap-1 rounded-[4px] bg-[var(--ouro-jogo-ano)] px-1.5 py-0.5 text-[10px] font-bold text-[var(--ouro-jogo-ano-text)]">
                <Crown className="h-3 w-3" aria-hidden="true" />
                <span>Jogo do ano</span>
              </span>
            )}
          </div>
          {jogo.genero && (
            <p className="hidden truncate text-xs text-[var(--biblioteca-text-muted)] lg:block">
              {jogo.genero}
            </p>
          )}
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-3 sm:gap-4 md:gap-6">
        <span className="rounded-[4px] border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] px-2 py-0.5 text-xs font-medium text-[var(--biblioteca-text-muted)]">
          {jogo.console}
        </span>
        {tempoFormatado && (
          <span className="hidden text-xs text-[var(--biblioteca-text-muted)] lg:inline-block">
            {tempoFormatado}
          </span>
        )}
        <span className="hidden text-xs text-[var(--biblioteca-text-muted)] sm:inline-block">
          {isoParaDataPt(jogo.finalizado_em)}
        </span>
        <DificuldadePill nivel={jogo.dificuldade} variante="neutra" />
        <NotaBadge nota={jogo.nota} tamanho="sm" />
      </div>

      <div
        ref={menuRef}
        onClick={(e) => e.stopPropagation()}
        className="relative shrink-0"
      >
        <button
          type="button"
          onClick={() => setMenuAberto((prev) => !prev)}
          aria-label={`Opções de ${jogo.nome}`}
          className="flex h-8 w-8 items-center justify-center rounded-[6px] text-[var(--biblioteca-text-faint)] transition-colors hover:bg-[var(--biblioteca-control-bg)] hover:text-[var(--biblioteca-text-primary)]"
        >
          <MoreVertical className="h-4 w-4" aria-hidden="true" />
        </button>

        {menuAberto && (
          <div className="absolute right-0 top-full z-20 mt-1 min-w-[120px] rounded-[8px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-control-bg)] p-1 shadow-lg">
            <button
              type="button"
              onClick={() => { setMenuAberto(false); onEditar(jogo) }}
              className="flex w-full items-center gap-2 rounded-[6px] px-2.5 py-1.5 text-xs font-medium text-[var(--biblioteca-text-primary)] hover:bg-[var(--biblioteca-card-bg)]"
            >
              <Pencil className="h-3.5 w-3.5" aria-hidden="true" />
              <span>Editar</span>
            </button>
            <button
              type="button"
              onClick={() => { setMenuAberto(false); onExcluir(jogo) }}
              className="flex w-full items-center gap-2 rounded-[6px] px-2.5 py-1.5 text-xs font-medium text-[var(--danger)] hover:bg-[var(--biblioteca-card-bg)]"
            >
              <Trash2 className="h-3.5 w-3.5" aria-hidden="true" />
              <span>Excluir</span>
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
