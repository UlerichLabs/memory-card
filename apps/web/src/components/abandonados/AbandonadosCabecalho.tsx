import { Plus } from 'lucide-react'

export interface AbandonadosCabecalhoProps {
  total: number
  onAbandonar: () => void
}

export function AbandonadosCabecalho({ total, onAbandonar }: AbandonadosCabecalhoProps) {
  return (
    <header className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold tracking-tight text-[var(--text-primary)] sm:text-3xl">
            Abandonados
          </h1>
          <span
            className={
              'rounded-full border border-[var(--abandonado-border)] ' +
              'bg-[var(--abandonado-bg)] px-2.5 py-0.5 text-xs font-semibold text-[var(--abandonado-text)]'
            }
          >
            {total} {total === 1 ? 'jogo' : 'jogos'}
          </span>
        </div>
        <p className="mt-1 text-sm text-[var(--text-secondary)]">
          Jogos que você começou e não terminou. Ficam fora das estatísticas, do Game do Ano e dos Games da Vida.
        </p>
      </div>
      <button
        type="button"
        onClick={onAbandonar}
        className={
          'inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-[var(--accent)] ' +
          'px-4 text-xs font-bold text-[var(--accent-foreground)] transition hover:opacity-90'
        }
      >
        <Plus className="h-4 w-4" aria-hidden="true" />
        <span>Abandonar jogo</span>
      </button>
    </header>
  )
}
