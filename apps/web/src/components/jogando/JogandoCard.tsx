import { MoreHorizontal } from 'lucide-react'
import { Menu } from '@base-ui/react/menu'
import type { JogoEmAndamento } from '@/types/jogando'
import { formatarDataJogando, rotuloDiasDesdeInicio } from '@/lib/jogandoUtils'
import { iniciais } from '@/lib/dashboardUtils'

interface JogandoCardProps {
  jogo: JogoEmAndamento
  onZerei: () => void
  onAbandonei: () => void
  onRemover: () => void
}

export function JogandoCard({ jogo, onZerei, onAbandonei, onRemover }: JogandoCardProps) {
  return (
    <article className="flex min-w-0 gap-3 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-3">
      {jogo.igdb_capa_url ? <img src={jogo.igdb_capa_url} alt={`Capa de ${jogo.nome}`} className="h-24 w-[72px] shrink-0 rounded-md border border-[var(--border)] object-cover" /> : <div className="flex h-24 w-[72px] shrink-0 items-center justify-center rounded-md border border-[var(--border-subtle)] bg-[linear-gradient(150deg,var(--bg-surface-alt),var(--bg-primary))] text-sm font-bold text-[var(--text-secondary)]">{iniciais(jogo.nome)}</div>}
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex min-w-0 items-start justify-between gap-2">
          <h3 className="truncate text-sm font-semibold text-[var(--text-primary)]" title={jogo.nome}>{jogo.nome}</h3>
          <Menu.Root>
            <Menu.Trigger type="button" aria-label="Mais opções" className="shrink-0 rounded p-1 text-[var(--text-muted)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)]"><MoreHorizontal className="size-4" aria-hidden="true" /></Menu.Trigger>
            <Menu.Portal><Menu.Positioner side="bottom" align="end" className="z-50"><Menu.Popup className="w-32 rounded-lg border border-[var(--border)] bg-[var(--bg-surface)] p-1 shadow-xl"><Menu.Item onClick={onRemover} className="cursor-pointer rounded px-3 py-2 text-xs text-[var(--danger)] outline-none hover:bg-[var(--bg-surface-alt)]">Remover</Menu.Item></Menu.Popup></Menu.Positioner></Menu.Portal>
          </Menu.Root>
        </div>
        <p className="mt-1 text-[11px] text-[var(--text-secondary)]">Começou em {formatarDataJogando(jogo.iniciado_em)}</p>
        <p className="mt-0.5 text-[11px] text-[var(--text-muted)]">{rotuloDiasDesdeInicio(jogo.iniciado_em)}</p>
        <div className="mt-auto flex gap-2 pt-2"><button type="button" onClick={onZerei} className="rounded-[6px] bg-[var(--accent)] px-2.5 py-1 text-[11px] font-bold text-[var(--accent-foreground)]">Zerei!</button><button type="button" onClick={onAbandonei} className="rounded-[6px] border border-[var(--border-subtle)] px-2.5 py-1 text-[11px] font-semibold text-[var(--text-secondary)] hover:text-[var(--text-primary)]">Abandonei</button></div>
      </div>
    </article>
  )
}
