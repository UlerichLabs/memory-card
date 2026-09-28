import { useState, useRef, useEffect } from 'react'
import { MoreVertical, Pencil, Trash2, Gamepad2, Star, Award } from 'lucide-react'
import type { JogoZeradoDTO, Dificuldade } from '@/types/jogos'
import { formatarCapaIGDB, isoParaDataPt } from '@/lib/utils'

export interface BibliotecaListItemProps {
  jogo: JogoZeradoDTO
  onEditar: (jogo: JogoZeradoDTO) => void
  onExcluir: (jogo: JogoZeradoDTO) => void
  onDetalhes?: (jogo: JogoZeradoDTO) => void
}

const DIFICULDADE_VAR: Record<Dificuldade, string> = {
  C: 'var(--difficulty-c)',
  B: 'var(--difficulty-b)',
  A: 'var(--difficulty-a)',
  AA: 'var(--difficulty-aa)',
  AAA: 'var(--difficulty-aaa)',
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

  const horas = Math.floor(jogo.tempo_jogado / 3600)
  const minutos = Math.floor((jogo.tempo_jogado % 3600) / 60)
  const tempoFormatado = `${horas}h ${minutos}m`
  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url)
  const difCor = DIFICULDADE_VAR[jogo.dificuldade] || 'var(--biblioteca-text-muted)'

  return (
    <div
      onClick={() => onDetalhes?.(jogo)}
      className="group flex h-14 cursor-pointer items-center justify-between gap-3 rounded-[8px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] px-3 transition-colors hover:bg-[var(--biblioteca-row-hover)]"
    >
      <div className="flex min-w-0 flex-1 items-center gap-3">
        {capaUrl ? (
          <img
            src={capaUrl}
            alt=""
            className="h-[42px] w-[32px] shrink-0 rounded-[4px] object-cover border border-[var(--biblioteca-card-border)]"
          />
        ) : (
          <div className="flex h-[42px] w-[32px] shrink-0 items-center justify-center rounded-[4px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-control-bg)] text-[var(--biblioteca-text-faint)]">
            <Gamepad2 className="h-4 w-4" aria-hidden="true" />
          </div>
        )}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="truncate text-[13px] font-bold text-[var(--biblioteca-text-primary)]">
              {jogo.nome}
            </span>
            {jogo.destaque && (
              <span className="inline-flex items-center gap-0.5 rounded-[4px] bg-[var(--biblioteca-gold)]/10 px-1.5 py-0.5 text-[10px] font-bold text-[var(--biblioteca-gold)]">
                <Award className="h-3 w-3" aria-hidden="true" />
                <span>Destaque</span>
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
        <span className="hidden text-xs text-[var(--biblioteca-text-muted)] lg:inline-block">
          {tempoFormatado}
        </span>
        <span className="hidden text-xs text-[var(--biblioteca-text-muted)] sm:inline-block">
          {isoParaDataPt(jogo.finalizado_em)}
        </span>
        <span
          style={{ borderColor: difCor, color: difCor }}
          className="rounded-[4px] border px-1.5 py-0.5 text-[10px] font-bold"
        >
          {jogo.dificuldade}
        </span>
        <div className="flex items-center gap-1 rounded-[4px] bg-[var(--biblioteca-control-bg)] px-1.5 py-0.5 text-xs font-bold text-[var(--biblioteca-gold)]">
          <Star className="h-3 w-3 fill-[var(--biblioteca-gold)]" aria-hidden="true" />
          <span>Nota {jogo.nota}</span>
        </div>
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
