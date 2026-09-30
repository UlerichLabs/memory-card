import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical, Pause, Trophy, X } from 'lucide-react'
import { formatarCapaIGDB } from '@/lib/utils'
import { obterIniciaisJogo } from './listas.utils'
import type { ListaItem } from '@/types/listas'

export interface FilaItemProps {
  item: ListaItem
  posicaoExibicao: number
  isFirst: boolean
  onZerei: (item: ListaItem) => void
  onAbandonei?: (item: ListaItem) => void
  onRemover: (itemId: number) => void
}

export function FilaItem({ item, posicaoExibicao, isFirst, onZerei, onAbandonei, onRemover }: FilaItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: item.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    zIndex: isDragging ? 20 : undefined,
    opacity: isDragging ? 0.7 : undefined,
  }

  const capaUrl = formatarCapaIGDB(item.igdb_capa_url ?? undefined, 't_cover_big')
  const iniciais = obterIniciaisJogo(item.nome)
  const metaTexto = [item.console, item.ano_lancamento].filter(Boolean).join(' · ')

  return (
    <li
      ref={setNodeRef}
      style={style}
      className={`flex items-center gap-3 rounded-xl p-[8px_12px_8px_4px] transition-colors ${
        isFirst
          ? 'border border-[var(--lista-row-first-border)] bg-[var(--lista-row-first-bg)]'
          : 'border border-[var(--lista-row-border)] bg-[var(--lista-row-bg)]'
      }`}
    >
      <button
        type="button"
        aria-label={`Mover ${item.nome}`}
        {...attributes}
        {...listeners}
        className="flex h-9 w-7 shrink-0 cursor-grab items-center justify-center text-[var(--lista-row-dim)] hover:text-[var(--text-primary)] active:cursor-grabbing focus:outline-none"
      >
        <GripVertical className="h-4 w-4" />
      </button>

      <span className="hidden w-6 shrink-0 text-right text-[15px] font-bold tabular-nums text-[var(--lista-row-dim)] sm:block">
        {posicaoExibicao}
      </span>

      <div className="relative flex h-[53px] w-[40px] shrink-0 items-center justify-center overflow-hidden rounded-md border border-[var(--lista-card-zerado-border)] bg-[var(--lista-cover-bg)]">
        {capaUrl ? (
          <img src={capaUrl} alt="" className="h-full w-full object-cover" loading="lazy" />
        ) : (
          <span className="text-[12px] font-bold text-[var(--lista-initials-zerado)] select-none">
            {iniciais}
          </span>
        )}
      </div>

      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex items-center gap-2">
          <span className="truncate text-[14px] font-semibold text-[var(--text-primary)]">
            {item.nome}
          </span>
          {isFirst && (
            <span className="shrink-0 rounded-full border border-[var(--lista-item-selected-border)] bg-[var(--hall-andamento-bg)] px-2 py-0.5 text-[11px] font-semibold text-[var(--hall-andamento-text)]">
              Próximo
            </span>
          )}
        </div>
        {metaTexto && (
          <span className="truncate text-[12px] text-[var(--lista-text-muted)]">
            {metaTexto}
          </span>
        )}
      </div>

      <button
        type="button"
        aria-label="Zerei!"
        onClick={() => onZerei(item)}
        className="hidden sm:flex h-9 items-center rounded-lg border border-[var(--lista-btn-zerei-border)] bg-[var(--lista-btn-zerei-bg)] px-3.5 text-[13px] font-bold text-[var(--lista-btn-zerei-text)] hover:opacity-90 shrink-0"
      >
        Zerei!
      </button>
      <button
        type="button"
        aria-label="Zerei!"
        onClick={() => onZerei(item)}
        className="flex sm:hidden h-11 w-11 items-center justify-center rounded-lg border border-[var(--lista-btn-zerei-border)] bg-[var(--lista-btn-zerei-bg)] text-[var(--lista-btn-zerei-text)] hover:opacity-90 shrink-0"
      >
        <Trophy className="h-5 w-5" />
      </button>

      {onAbandonei && (
        <>
          <button
            type="button"
            aria-label="Abandonei"
            onClick={() => onAbandonei(item)}
            className={
              'hidden sm:flex h-9 items-center gap-1.5 rounded-lg border ' +
              'border-[var(--lista-btn-abandonei-border)] bg-[var(--lista-btn-abandonei-bg)] ' +
              'px-3 text-[13px] font-bold text-[var(--lista-btn-abandonei-text)] hover:opacity-90 shrink-0'
            }
          >
            <Pause className="h-3.5 w-3.5 fill-current" />
            <span>Abandonei</span>
          </button>
          <button
            type="button"
            aria-label="Abandonei"
            onClick={() => onAbandonei(item)}
            className={
              'flex sm:hidden h-11 w-11 items-center justify-center rounded-lg border ' +
              'border-[var(--lista-btn-abandonei-border)] bg-[var(--lista-btn-abandonei-bg)] ' +
              'text-[var(--lista-btn-abandonei-text)] hover:opacity-90 shrink-0'
            }
          >
            <Pause className="h-5 w-5 fill-current" />
          </button>
        </>
      )}

      <button
        type="button"
        aria-label={`Remover ${item.nome} da fila`}
        onClick={() => onRemover(item.id)}
        className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-[var(--lista-text-muted)] hover:text-[var(--danger)] focus:outline-none"
      >
        <X className="h-4 w-4" />
      </button>
    </li>
  )
}
