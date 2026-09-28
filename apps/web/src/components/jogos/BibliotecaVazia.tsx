import { Disc3, SearchX, Plus, RotateCcw } from 'lucide-react'

export interface BibliotecaVaziaProps {
  possuiFiltrosAtivos: boolean
  onLimparFiltros: () => void
  onRegistrarPrimeiroJogo: () => void
}

export function BibliotecaVazia({
  possuiFiltrosAtivos,
  onLimparFiltros,
  onRegistrarPrimeiroJogo,
}: BibliotecaVaziaProps) {
  if (possuiFiltrosAtivos) {
    return (
      <div className="flex flex-col items-center justify-center rounded-[10px] border border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-12 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-full border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] text-[var(--biblioteca-control-placeholder)]">
          <SearchX className="h-6 w-6" aria-hidden="true" />
        </div>
        <h2 className="mt-4 text-base font-bold text-[var(--biblioteca-text-primary)]">
          Nenhum jogo encontrado com os filtros aplicados
        </h2>
        <p className="mt-1 text-xs text-[var(--biblioteca-text-muted)]">
          Tente buscar com outros termos ou redefinir os filtros selecionados.
        </p>
        <button
          type="button"
          onClick={onLimparFiltros}
          className="mt-4 inline-flex items-center gap-1.5 rounded-[6px] border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] px-3 py-1.5 text-xs font-medium text-[var(--biblioteca-text-secondary)] transition-colors hover:border-[var(--biblioteca-control-border-hover)] hover:text-[var(--biblioteca-text-primary)]"
        >
          <RotateCcw className="h-3.5 w-3.5" aria-hidden="true" />
          <span>Limpar filtros</span>
        </button>
      </div>
    )
  }

  return (
    <div className="flex flex-col items-center justify-center rounded-[10px] border border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-12 text-center">
      <div className="flex h-14 w-14 items-center justify-center rounded-full border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] text-[var(--biblioteca-text-secondary)]">
        <Disc3 className="h-7 w-7 text-[var(--accent)]" aria-hidden="true" />
      </div>
      <div className="mt-4 space-y-1">
        <h2 className="text-base font-bold text-[var(--biblioteca-text-primary)]">
          Nenhum jogo registrado ainda
        </h2>
        <p className="max-w-sm text-xs text-[var(--biblioteca-text-muted)]">
          Sua biblioteca está vazia. Comece a registrar seus jogos zerados com nota, tempo jogado e dificuldade.
        </p>
      </div>
      <button
        type="button"
        onClick={onRegistrarPrimeiroJogo}
        className="mt-5 inline-flex items-center justify-center gap-2 rounded-[7px] bg-[var(--accent)] px-4 py-2 text-xs font-bold text-[var(--accent-foreground)] transition-opacity hover:opacity-90"
      >
        <Plus className="h-4 w-4" aria-hidden="true" />
        <span>Registrar primeiro jogo</span>
      </button>
    </div>
  )
}
