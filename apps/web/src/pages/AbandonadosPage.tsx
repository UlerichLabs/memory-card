import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { useAbandonadosStore } from '@/stores/abandonadosStore'
import { AbandonadosControles } from '@/components/abandonados/AbandonadosControles'
import { AbandonadosGrade } from '@/components/abandonados/AbandonadosGrade'
import { AbandonadosVazio } from '@/components/abandonados/AbandonadosVazio'
import { AbandonadosSkeleton } from '@/components/abandonados/AbandonadosSkeleton'
import { BibliotecaPaginacao } from '@/components/jogos/BibliotecaPaginacao'
import { ExcluirAbandonadoDialog } from '@/components/abandonados/ExcluirAbandonadoDialog'
import type { OrdenarAbandonados } from '@/types/abandonados'

export function AbandonadosPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const {
    jogos, meta, filtros, totalGeral, isLoading,
    isExcluirModalOpen, jogoParaExcluir,
    abrirModalCriacao, abrirModalEdicao, abrirModalExcluir, fecharModalExcluir,
    carregarJogos, carregarFiltros, carregarTotal, excluirJogo,
  } = useAbandonadosStore()

  const [isExcluindo, setIsExcluindo] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  const busca = searchParams.get('busca') ?? ''
  const consoleVal = searchParams.get('console') ?? ''
  const ordenarVal = (searchParams.get('ordenar') as OrdenarAbandonados) || 'recentes'
  const pagina = Math.max(1, parseInt(searchParams.get('pagina') || '1', 10))

  useEffect(() => {
    carregarFiltros()
    carregarTotal()
  }, [])

  useEffect(() => {
    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller
    carregarJogos({
      pagina,
      por_pagina: 12,
      busca: busca.trim() || undefined,
      console: consoleVal.trim() || undefined,
      ordenar: ordenarVal !== 'recentes' ? ordenarVal : undefined,
    }, controller.signal).catch(() => undefined)

    return () => controller.abort()
  }, [busca, consoleVal, ordenarVal, pagina])

  function atualizarFiltros(novos: Record<string, string | null>) {
    const sp = new URLSearchParams(searchParams)
    Object.entries(novos).forEach(([k, v]) => {
      if (v === null || v === '') sp.delete(k)
      else sp.set(k, v)
    })
    setSearchParams(sp, { replace: true })
  }

  function handleBuscaChange(termo: string) {
    atualizarFiltros({ busca: termo || null, pagina: null })
  }

  function handleConsoleChange(val: string) {
    atualizarFiltros({ console: val || null, pagina: null })
  }

  function handleOrdenarChange(val: OrdenarAbandonados) {
    atualizarFiltros({ ordenar: val !== 'recentes' ? val : null, pagina: null })
  }

  function handleLimparFiltros() {
    setSearchParams(new URLSearchParams(), { replace: true })
  }

  function handleMudarPagina(novaPagina: number) {
    atualizarFiltros({ pagina: novaPagina > 1 ? String(novaPagina) : null })
  }

  async function handleConfirmarExcluir() {
    if (!jogoParaExcluir) return
    setIsExcluindo(true)
    try {
      await excluirJogo(jogoParaExcluir.id)
      fecharModalExcluir()
    } finally {
      setIsExcluindo(false)
    }
  }

  const temFiltrosAtivos = busca.trim() !== '' || consoleVal !== '' || ordenarVal !== 'recentes'
  const totalExibicao = totalGeral || meta.total

  return (
    <div className="flex min-h-screen flex-col bg-[var(--bg-primary)]">
      <Topbar />
      <main className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-8 sm:px-6 lg:px-8">
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
                {totalExibicao} {totalExibicao === 1 ? 'jogo' : 'jogos'}
              </span>
            </div>
            <p className="mt-1 text-sm text-[var(--text-secondary)]">
              Jogos que você começou e não terminou. Ficam fora das estatísticas, do Game do Ano e dos Games da Vida.
            </p>
          </div>
          <button
            type="button"
            onClick={() => abrirModalCriacao()}
            className={
              'inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-[var(--accent)] ' +
              'px-4 text-xs font-bold text-[var(--accent-foreground)] transition hover:opacity-90'
            }
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            <span>Abandonar jogo</span>
          </button>
        </header>

        <AbandonadosControles
          busca={busca}
          consoleVal={consoleVal}
          ordenarVal={ordenarVal}
          opcoesConsole={filtros.consoles}
          onBuscaChange={handleBuscaChange}
          onConsoleChange={handleConsoleChange}
          onOrdenarChange={handleOrdenarChange}
          onLimparFiltros={handleLimparFiltros}
        />

        {isLoading ? (
          <AbandonadosSkeleton />
        ) : meta.total === 0 ? (
          <AbandonadosVazio
            possuiFiltrosAtivos={temFiltrosAtivos}
            onLimparFiltros={handleLimparFiltros}
            onAbandonarJogo={() => abrirModalCriacao()}
          />
        ) : (
          <div className="flex flex-col gap-6">
            <AbandonadosGrade
              jogos={jogos}
              onEditar={abrirModalEdicao}
              onExcluir={abrirModalExcluir}
            />
            <BibliotecaPaginacao meta={meta} onMudarPagina={handleMudarPagina} />
          </div>
        )}
      </main>

      <ExcluirAbandonadoDialog
        open={isExcluirModalOpen}
        onOpenChange={(open) => !open && fecharModalExcluir()}
        onConfirm={handleConfirmarExcluir}
        jogoNome={jogoParaExcluir?.nome}
        isLoading={isExcluindo}
      />
    </div>
  )
}
