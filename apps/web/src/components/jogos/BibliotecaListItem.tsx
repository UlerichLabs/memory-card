import { useState, useRef, useEffect } from 'react'
import { MoreVertical, Pencil, Trash2, Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/lib/services/jogosService'
import { formatarCapaIGDB, isoParaDataPt } from '@/lib/utils'

export interface BibliotecaListItemProps {
  jogo: JogoZeradoDTO
  onEditar: (jogo: JogoZeradoDTO) => void
  onExcluir: (jogo: JogoZeradoDTO) => void
}

export function BibliotecaListItem({ jogo, onEditar, onExcluir }: BibliotecaListItemProps) {
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

  return (
    <div className="flex items-center justify-between gap-3 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-2.5 transition-colors hover:border-[var(--border)]">
      <div className="flex items-center gap-3 min-w-0 flex-1">
        {capaUrl ? (
          <img src={capaUrl} alt="" className="h-14 w-10 shrink-0 rounded object-cover border border-[var(--border-subtle)]" />
        ) : (
          <div className="flex h-14 w-10 shrink-0 items-center justify-center rounded border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] text-[var(--text-faint)]">
            <Gamepad2 className="h-4 w-4" aria-hidden="true" />
          </div>
        )}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate text-sm font-semibold text-[var(--text-primary)]">{jogo.nome}</span>
            {jogo.destaque && (
              <span className="rounded bg-[var(--highlight-gold)]/10 px-1.5 py-0.5 text-[10px] font-bold text-[var(--highlight-gold)]">
                Destaque
              </span>
            )}
          </div>
          {jogo.genero && <p className="truncate text-xs text-[var(--text-muted)]">{jogo.genero}</p>}
        </div>
      </div>

      <div className="hidden sm:flex items-center gap-6 shrink-0">
        <span className="rounded bg-[var(--bg-surface-alt)] px-2 py-0.5 text-xs text-[var(--text-secondary)] font-medium border border-[var(--border-subtle)]">
          {jogo.console}
        </span>
        <span className="text-xs font-semibold text-[var(--accent)] min-w-[50px] text-right">
          Nota {jogo.nota}
        </span>
        <span className="text-xs text-[var(--text-secondary)] min-w-[60px] text-right">
          {tempoFormatado}
        </span>
        <span className="text-xs text-[var(--text-muted)] min-w-[80px] text-right">
          {isoParaDataPt(jogo.finalizado_em)}
        </span>
      </div>

      <div ref={menuRef} className="relative shrink-0">
        <button
          type="button"
          onClick={() => setMenuAberto((prev) => !prev)}
          aria-label={`Opções de ${jogo.nome}`}
          className="flex h-8 w-8 items-center justify-center rounded-md text-[var(--text-muted)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)] transition-colors"
        >
          <MoreVertical className="h-4 w-4" aria-hidden="true" />
        </button>

        {menuAberto && (
          <div className="absolute right-0 top-full z-20 mt-1 min-w-[130px] rounded-lg border border-[var(--border)] bg-[var(--bg-surface-alt)] p-1 shadow-lg">
            <button
              type="button"
              onClick={() => { setMenuAberto(false); onEditar(jogo) }}
              className="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-xs text-[var(--text-primary)] hover:bg-[var(--bg-surface)] transition-colors"
            >
              <Pencil className="h-3.5 w-3.5" aria-hidden="true" />
              <span>Editar</span>
            </button>
            <button
              type="button"
              onClick={() => { setMenuAberto(false); onExcluir(jogo) }}
              className="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-xs text-[var(--danger)] hover:bg-[var(--bg-surface)] transition-colors"
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
