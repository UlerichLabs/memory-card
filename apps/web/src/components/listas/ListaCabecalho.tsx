import { Filter, Pencil, Trash2 } from 'lucide-react'
import { useListasStore } from '@/stores/listasStore'
import type { ListaDetalhada } from '@/types/listas'

export interface ListaCabecalhoProps {
  lista: ListaDetalhada
}

export function ListaCabecalho({ lista }: ListaCabecalhoProps) {
  const { abrirModalEditar, abrirModalExcluir } = useListasStore()
  const isDesafio = lista.tipo === 'desafio'
  const origem = lista.origem
  const temRegraChip = isDesafio && origem

  let chipTexto = ''
  if (temRegraChip) {
    chipTexto = `${origem.tipo[0].toUpperCase()}${origem.tipo.slice(1)}: ${origem.nome}`
  }

  const actionButtons = (
    <div className="flex items-center gap-2">
      <button
        type="button"
        aria-label="Editar"
        onClick={() => abrirModalEditar(lista)}
        className="flex h-10 w-10 items-center justify-center rounded-[10px] border border-[var(--lista-btn-icon-border)] bg-[var(--lista-btn-icon-bg)] text-[var(--lista-btn-icon-text)] hover:opacity-90 focus:outline-none"
      >
        <Pencil className="h-4 w-4" />
      </button>
      <button
        type="button"
        aria-label="Excluir"
        onClick={() => abrirModalExcluir(lista)}
        className="flex h-10 w-10 items-center justify-center rounded-[10px] border border-[var(--lista-btn-delete-border)] bg-transparent text-[var(--lista-btn-delete-text)] hover:bg-[var(--danger)]/10 focus:outline-none"
      >
        <Trash2 className="h-4 w-4" />
      </button>
    </div>
  )

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between sm:hidden">
        <span
          className={`inline-flex rounded-full px-2.5 py-0.5 text-[12px] font-semibold ${
            isDesafio
              ? 'border border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)] text-[var(--lista-pill-desafio-text)]'
              : 'border border-[var(--lista-item-selected-border)] bg-[var(--hall-andamento-bg)] text-[var(--hall-andamento-text)]'
          }`}
        >
          {isDesafio ? 'Desafio' : 'Fila'}
        </span>
        {actionButtons}
      </div>

      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-2">
          <div className="hidden sm:flex">
            <span
              className={`inline-flex rounded-full px-2.5 py-0.5 text-[12px] font-semibold ${
                isDesafio
                  ? 'border border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)] text-[var(--lista-pill-desafio-text)]'
                  : 'border border-[var(--lista-item-selected-border)] bg-[var(--hall-andamento-bg)] text-[var(--hall-andamento-text)]'
              }`}
            >
              {isDesafio ? 'Desafio' : 'Fila'}
            </span>
          </div>

          <h1 className="text-[24px] font-bold text-[var(--text-primary)]">
            {lista.nome}
          </h1>

          {lista.descricao && (
            <p className="text-[14px] text-[var(--lista-text-secondary)]">
              {lista.descricao}
            </p>
          )}

          {temRegraChip && (
            <div className="mt-1 inline-flex h-7 w-fit items-center gap-1.5 rounded-full border border-[var(--lista-chip-border)] bg-[var(--lista-chip-bg)] px-2.5 text-[12px] text-[var(--lista-chip-text)]">
              <Filter className="h-[13px] w-[13px]" />
              <span>{chipTexto}</span>
            </div>
          )}
        </div>

        <div className="hidden sm:block">
          {actionButtons}
        </div>
      </div>
    </div>
  )
}
