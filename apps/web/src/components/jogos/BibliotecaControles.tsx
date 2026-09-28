import { LayoutGrid, List } from 'lucide-react'

export interface BibliotecaControlesProps {
  totalJogos: number
  modo: 'grid' | 'list'
  onAlternarModo: (modo: 'grid' | 'list') => void
}

export function BibliotecaControles({
  totalJogos,
  modo,
  onAlternarModo,
}: BibliotecaControlesProps) {
  return (
    <div className="flex items-center justify-between">
      <p className="text-[13px] text-[var(--biblioteca-text-muted)]">
        <span className="font-bold text-[var(--biblioteca-text-primary)]">{totalJogos}</span>{' '}
        {totalJogos === 1 ? 'jogo encontrado' : 'jogos encontrados'}
      </p>

      <div className="flex items-center gap-1 rounded-[6px] border border-[var(--biblioteca-toggle-border)] bg-[var(--biblioteca-toggle-bg)] p-0.5">
        <button
          type="button"
          onClick={() => onAlternarModo('grid')}
          aria-label="Visualização em grade"
          aria-pressed={modo === 'grid'}
          className={`flex h-8 w-8 items-center justify-center rounded-[4px] transition-colors ${
            modo === 'grid'
              ? 'bg-[var(--biblioteca-toggle-active-bg)] text-[var(--biblioteca-toggle-active-text)]'
              : 'text-[var(--biblioteca-toggle-inactive-text)] hover:text-[var(--biblioteca-text-primary)]'
          }`}
        >
          <LayoutGrid className="h-4 w-4" aria-hidden="true" />
        </button>

        <button
          type="button"
          onClick={() => onAlternarModo('list')}
          aria-label="Visualização em lista"
          aria-pressed={modo === 'list'}
          className={`flex h-8 w-8 items-center justify-center rounded-[4px] transition-colors ${
            modo === 'list'
              ? 'bg-[var(--biblioteca-toggle-active-bg)] text-[var(--biblioteca-toggle-active-text)]'
              : 'text-[var(--biblioteca-toggle-inactive-text)] hover:text-[var(--biblioteca-text-primary)]'
          }`}
        >
          <List className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>
    </div>
  )
}
