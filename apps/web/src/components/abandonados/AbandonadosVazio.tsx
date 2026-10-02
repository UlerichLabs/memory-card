import { Pause, RotateCcw, SearchX } from 'lucide-react'

export interface AbandonadosVazioProps {
  possuiFiltrosAtivos: boolean
  onLimparFiltros: () => void
}

export function AbandonadosVazio({
  possuiFiltrosAtivos,
  onLimparFiltros,
}: AbandonadosVazioProps) {
  if (possuiFiltrosAtivos) {
    return (
      <div
        className={
          'flex flex-col items-center justify-center rounded-xl border ' +
          'border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-12 text-center'
        }
      >
        <div
          className={
            'flex h-12 w-12 items-center justify-center rounded-full border ' +
            'border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] ' +
            'text-[var(--biblioteca-control-placeholder)]'
          }
        >
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
          className={
            'mt-4 inline-flex items-center gap-1.5 rounded-lg border ' +
            'border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] ' +
            'px-3.5 py-2 text-xs font-semibold text-[var(--biblioteca-text-secondary)] ' +
            'transition hover:text-[var(--text-primary)]'
          }
        >
          <RotateCcw className="h-3.5 w-3.5" aria-hidden="true" />
          <span>Limpar filtros</span>
        </button>
      </div>
    )
  }

  return (
    <div
      className={
        'flex flex-col items-center justify-center rounded-xl border ' +
        'border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-12 text-center'
      }
    >
      <div
        className={
          'flex h-14 w-14 items-center justify-center rounded-full border ' +
          'border-[var(--abandonado-border)] bg-[var(--abandonado-bg)] text-[var(--abandonado-text)]'
        }
      >
        <Pause className="h-7 w-7 fill-current" aria-hidden="true" />
      </div>
      <div className="mt-4 space-y-1">
        <h2 className="text-base font-bold text-[var(--text-primary)]">
          Nenhum jogo abandonado
        </h2>
        <p className="max-w-md text-xs text-[var(--text-muted)] leading-relaxed">
          Jogos que você começou e não terminou ficam aqui, fora das suas estatísticas.
          Você também pode abandonar um jogo direto de uma fila.
        </p>
      </div>
      <p className="mt-5 text-xs text-[var(--text-secondary)]">Use o botão Novo no topo e escolha Abandonar jogo.</p>
    </div>
  )
}
