import { LayoutGrid, List } from 'lucide-react'

export interface BibliotecaControlesProps {
  modo: 'grid' | 'list'
  onAlternarModo: (modo: 'grid' | 'list') => void
  totalJogos?: number
  className?: string
}

export function BibliotecaControles({
  modo,
  onAlternarModo,
  className = '',
}: BibliotecaControlesProps) {
  return (
    <div
      role="group"
      aria-label="Modo de visualização"
      className={`inline-flex items-center gap-1 rounded-[12px] border border-[var(--biblioteca-toggle-group-border)] bg-[var(--biblioteca-toggle-group-bg)] p-1 ${className}`.trim()}
    >
      <button
        type="button"
        onClick={() => onAlternarModo('grid')}
        aria-label="Visualização em grade"
        aria-pressed={modo === 'grid'}
        className={`flex h-9 w-9 items-center justify-center rounded-[8px] transition-colors ${
          modo === 'grid'
            ? 'bg-[var(--biblioteca-toggle-btn-active-bg)] text-[var(--biblioteca-toggle-btn-active-icon)]'
            : 'text-[var(--biblioteca-toggle-btn-inactive-icon)] hover:text-[var(--biblioteca-text-primary)]'
        }`}
      >
        <LayoutGrid className="h-4 w-4" aria-hidden="true" />
      </button>

      <button
        type="button"
        onClick={() => onAlternarModo('list')}
        aria-label="Visualização em lista"
        aria-pressed={modo === 'list'}
        className={`flex h-9 w-9 items-center justify-center rounded-[8px] transition-colors ${
          modo === 'list'
            ? 'bg-[var(--biblioteca-toggle-btn-active-bg)] text-[var(--biblioteca-toggle-btn-active-icon)]'
            : 'text-[var(--biblioteca-toggle-btn-inactive-icon)] hover:text-[var(--biblioteca-text-primary)]'
        }`}
      >
        <List className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  )
}
