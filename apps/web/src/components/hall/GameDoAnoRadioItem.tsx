import type { KeyboardEvent } from 'react'
import { Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB, isoParaDataPt } from '@/lib/utils'
import { NotaBadge } from '@/components/jogos/NotaBadge'

export interface GameDoAnoRadioItemProps {
  jogo: JogoZeradoDTO
  selecionado: boolean
  tabIndex: number
  onSelect: () => void
  onKeyDown: (e: KeyboardEvent<HTMLDivElement>) => void
  itemRef?: (el: HTMLDivElement | null) => void
}

export function GameDoAnoRadioItem({
  jogo,
  selecionado,
  tabIndex,
  onSelect,
  onKeyDown,
  itemRef,
}: GameDoAnoRadioItemProps) {
  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url, 't_cover_big')
  const dataPt = isoParaDataPt(jogo.finalizado_em)
  const metaTexto = dataPt ? `${jogo.console} · finalizado em ${dataPt}` : jogo.console

  return (
    <div
      ref={itemRef}
      role="radio"
      aria-checked={selecionado}
      tabIndex={tabIndex}
      onClick={onSelect}
      onKeyDown={onKeyDown}
      className={`flex cursor-pointer items-center justify-between gap-[14px] rounded-[12px] p-[8px_14px_8px_12px] transition-colors outline-none focus-visible:ring-2 focus-visible:ring-[var(--hall-ouro)] ${
        selecionado
          ? 'border border-[var(--hall-ouro)] bg-[var(--hall-radio-item-active-bg)]'
          : 'border border-[var(--hall-radio-item-border)] bg-[var(--hall-radio-item-bg)] hover:border-[var(--hall-ouro)]/50'
      }`}
    >
      <div className="flex min-w-0 flex-1 items-center gap-[14px]">
        <div
          className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 ${
            selecionado
              ? 'border-[var(--hall-ouro)]'
              : 'border-[var(--hall-radio-circle-border)]'
          }`}
          aria-hidden="true"
        >
          {selecionado && (
            <div className="h-2.5 w-2.5 rounded-full bg-[var(--hall-ouro)]" />
          )}
        </div>

        <div className="h-[53px] w-[40px] shrink-0 overflow-hidden rounded-[6px] bg-[var(--bg-surface-alt)]">
          {capaUrl ? (
            <img
              src={capaUrl}
              alt=""
              className="h-full w-full object-cover"
              loading="lazy"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-[var(--text-muted)]">
              <Gamepad2 className="h-4 w-4 opacity-40" aria-hidden="true" />
            </div>
          )}
        </div>

        <div className="flex min-w-0 flex-1 flex-col gap-0.5">
          <span
            className="truncate text-[14px] font-semibold text-[var(--text-primary)]"
            title={jogo.nome}
          >
            {jogo.nome}
          </span>
          <span className="truncate text-[12px] text-[var(--hall-muted)]">
            {metaTexto}
          </span>
        </div>
      </div>

      <div className="shrink-0 pl-2">
        <NotaBadge nota={jogo.nota} tamanho="sm" />
      </div>
    </div>
  )
}
