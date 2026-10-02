import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { ListagemMeta } from '@/types/jogos'

export interface BibliotecaPaginacaoProps {
  meta: ListagemMeta
  onMudarPagina: (pagina: number) => void
}

function calcularPaginas(paginaAtual: number, totalPaginas: number): (number | string)[] {
  if (totalPaginas <= 7) {
    return Array.from({ length: totalPaginas }, (_, i) => i + 1)
  }
  if (paginaAtual <= 4) {
    return [1, 2, 3, 4, 5, '...', totalPaginas]
  }
  if (paginaAtual >= totalPaginas - 3) {
    return [
      1,
      '...',
      totalPaginas - 4,
      totalPaginas - 3,
      totalPaginas - 2,
      totalPaginas - 1,
      totalPaginas,
    ]
  }
  return [1, '...', paginaAtual - 1, paginaAtual, paginaAtual + 1, '...', totalPaginas]
}

export function BibliotecaPaginacao({ meta, onMudarPagina }: BibliotecaPaginacaoProps) {
  if (meta.total_paginas <= 1) return null

  const inicio = meta.total === 0 ? 0 : (meta.pagina - 1) * meta.por_pagina + 1
  const fim = Math.min(meta.pagina * meta.por_pagina, meta.total)
  const paginas = calcularPaginas(meta.pagina, meta.total_paginas)

  return (
    <nav
      aria-label="Paginação da biblioteca"
      className="mt-6 flex flex-col items-center justify-between gap-3 border-t border-[var(--biblioteca-divider)] pt-4 sm:flex-row"
    >
      <p className="text-xs text-[var(--biblioteca-text-muted)]">
        Mostrando <span className="font-semibold text-[var(--biblioteca-text-primary)]">{inicio}–{fim}</span> de{' '}
        <span className="font-semibold text-[var(--biblioteca-text-primary)]">{meta.total}</span>
      </p>

      <div className="flex items-center gap-1.5">
        <button
          type="button"
          disabled={meta.pagina <= 1}
          onClick={() => onMudarPagina(meta.pagina - 1)}
          aria-label="Página anterior"
          className="inline-flex h-8 items-center gap-1 rounded-[6px] border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] px-2.5 text-xs text-[var(--biblioteca-control-text)] transition-colors hover:border-[var(--biblioteca-control-border-hover)] disabled:cursor-not-allowed disabled:opacity-40"
        >
          <ChevronLeft className="h-3.5 w-3.5" aria-hidden="true" />
          <span>Anterior</span>
        </button>

        {paginas.map((item, idx) => {
          if (typeof item === 'string') {
            return (
              <span
                key={`ellipsis-${idx}`}
                className="px-2 text-xs text-[var(--biblioteca-control-placeholder)]"
                aria-hidden="true"
              >
                ...
              </span>
            )
          }

          const ativo = item === meta.pagina
          return (
            <button
              key={item}
              type="button"
              onClick={() => onMudarPagina(item)}
              aria-current={ativo ? 'page' : undefined}
              aria-label={`Página ${item}`}
              className={`h-8 min-w-[32px] rounded-[6px] px-2 text-xs font-semibold transition-colors ${
                ativo
                  ? 'btn-primario btn-primario-ativo text-white'
                  : 'border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] text-[var(--biblioteca-control-text)] hover:border-[var(--biblioteca-control-border-hover)]'
              }`}
            >
              {item}
            </button>
          )
        })}

        <button
          type="button"
          disabled={meta.pagina >= meta.total_paginas}
          onClick={() => onMudarPagina(meta.pagina + 1)}
          aria-label="Próxima página"
          className="inline-flex h-8 items-center gap-1 rounded-[6px] border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] px-2.5 text-xs text-[var(--biblioteca-control-text)] transition-colors hover:border-[var(--biblioteca-control-border-hover)] disabled:cursor-not-allowed disabled:opacity-40"
        >
          <span>Próxima</span>
          <ChevronRight className="h-3.5 w-3.5" aria-hidden="true" />
        </button>
      </div>
    </nav>
  )
}
