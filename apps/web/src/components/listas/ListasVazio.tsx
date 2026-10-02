import { Trophy, Plus } from 'lucide-react'

export interface ListasVazioProps {
  onNovaLista: () => void
}

export function ListasVazio({ onNovaLista }: ListasVazioProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-2xl border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] px-6 py-16 text-center">
      <div className="flex h-14 w-14 items-center justify-center rounded-full bg-[var(--lista-cover-bg)] text-[var(--hall-ouro)]">
        <Trophy className="h-7 w-7" />
      </div>
      <div className="flex max-w-md flex-col gap-1.5">
        <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
          Nenhuma lista ou desafio ainda
        </h2>
        <p className="text-[14px] text-[var(--lista-text-secondary)]">
          Crie uma fila do que jogar em seguida ou um desafio como zerar uma franquia inteira.
        </p>
      </div>
      <button
        type="button"
        onClick={onNovaLista}
        className="mt-2 flex h-11 items-center gap-2 rounded-[10px] bg-[var(--accent)] px-5 text-[14px] font-semibold text-[var(--accent-foreground)] transition-opacity hover:opacity-90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]"
      >
        <Plus className="h-4 w-4" />
        <span>Nova lista ou desafio</span>
      </button>
    </div>
  )
}
