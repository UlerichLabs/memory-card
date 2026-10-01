import { useEffect, useState } from 'react'
import { Search, X } from 'lucide-react'
import { BibliotecaSelect, type BibliotecaSelectOption } from '@/components/jogos/BibliotecaSelect'
import type { OrdenarAbandonados } from '@/types/abandonados'

const OPCOES_ORDENAR: BibliotecaSelectOption[] = [
  { value: 'recentes', label: 'Mais recentes' },
  { value: 'antigos', label: 'Mais antigos' },
  { value: 'nome', label: 'Nome (A-Z)' },
  { value: 'tempo', label: 'Mais tempo jogado' },
]

export interface AbandonadosControlesProps {
  busca: string
  consoleVal: string
  ordenarVal: OrdenarAbandonados
  opcoesConsole: string[]
  onBuscaChange: (termo: string) => void
  onConsoleChange: (console: string) => void
  onOrdenarChange: (ordenar: OrdenarAbandonados) => void
  onLimparFiltros: () => void
}

export function AbandonadosControles({
  busca,
  consoleVal,
  ordenarVal,
  opcoesConsole,
  onBuscaChange,
  onConsoleChange,
  onOrdenarChange,
  onLimparFiltros,
}: AbandonadosControlesProps) {
  const [buscaLocal, setBuscaLocal] = useState(busca)

  useEffect(() => {
    setBuscaLocal(busca)
  }, [busca])

  useEffect(() => {
    const timer = setTimeout(() => {
      if (buscaLocal.trim() !== busca.trim()) {
        onBuscaChange(buscaLocal.trim())
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [buscaLocal, busca, onBuscaChange])

  const consoleOptions: BibliotecaSelectOption[] = [
    { value: '', label: 'Todas as plataformas' },
    ...opcoesConsole.map((c) => ({ value: c, label: c })),
  ]

  const temFiltrosAtivos = busca.trim() !== '' || consoleVal !== '' || ordenarVal !== 'recentes'

  return (
    <div
      className={
        'flex flex-col gap-3 rounded-xl border border-[var(--biblioteca-panel-border)] ' +
        'bg-[var(--biblioteca-panel-bg)] p-4 sm:flex-row sm:items-center'
      }
    >
      <div className="relative flex min-w-0 flex-1 items-center">
        <Search
          className="absolute left-3.5 h-4 w-4 text-[var(--biblioteca-control-placeholder)]"
          aria-hidden="true"
        />
        <input
          type="search"
          value={buscaLocal}
          onChange={(e) => setBuscaLocal(e.target.value)}
          placeholder="Buscar pelo nome do jogo…"
          aria-label="Buscar jogos abandonados"
          className={
            'h-11 w-full rounded-lg border border-[var(--biblioteca-input-border)] ' +
            'bg-[var(--biblioteca-input-bg)] pl-10 pr-9 text-sm text-[var(--biblioteca-input-text)] ' +
            'placeholder:text-[var(--biblioteca-control-placeholder)] transition ' +
            'hover:border-[var(--biblioteca-control-border-hover)] ' +
            'focus:border-[var(--biblioteca-input-active-border)] focus:outline-none'
          }
        />
        {buscaLocal && (
          <button
            type="button"
            onClick={() => {
              setBuscaLocal('')
              onBuscaChange('')
            }}
            aria-label="Limpar busca"
            className="absolute right-3 text-[var(--biblioteca-control-placeholder)] hover:text-[var(--text-primary)]"
          >
            <X className="h-4 w-4" aria-hidden="true" />
          </button>
        )}
      </div>

      <div className="flex flex-wrap items-center gap-2.5 sm:flex-nowrap">
        <div className="w-full min-w-[180px] sm:w-[200px]">
          <BibliotecaSelect
            value={consoleVal}
            onChange={onConsoleChange}
            options={consoleOptions}
            placeholder="Plataforma"
            ariaLabel="Filtrar por plataforma"
          />
        </div>

        <div className="w-full min-w-[180px] sm:w-[190px]">
          <BibliotecaSelect
            value={ordenarVal}
            onChange={(val) => onOrdenarChange(val as OrdenarAbandonados)}
            options={OPCOES_ORDENAR}
            placeholder="Ordenar por"
            ariaLabel="Ordenar jogos abandonados"
          />
        </div>

        {temFiltrosAtivos && (
          <button
            type="button"
            onClick={onLimparFiltros}
            aria-label="Limpar todos os filtros"
            className={
              'flex h-11 shrink-0 items-center justify-center rounded-lg border ' +
              'border-[var(--biblioteca-btn-limpar-border)] px-3 text-xs font-semibold ' +
              'text-[var(--biblioteca-btn-limpar-text)] transition hover:border-[var(--border)] ' +
              'hover:text-[var(--text-primary)]'
            }
          >
            Limpar filtros
          </button>
        )}
      </div>
    </div>
  )
}
