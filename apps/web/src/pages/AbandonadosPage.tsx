import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { X } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { useAbandonadosStore } from '@/stores/abandonadosStore'
import { AbandonadosCabecalho } from '@/components/abandonados/AbandonadosCabecalho'
import { AbandonadosControles } from '@/components/abandonados/AbandonadosControles'
import { AbandonadosGrade } from '@/components/abandonados/AbandonadosGrade'
import { AbandonadosVazio } from '@/components/abandonados/AbandonadosVazio'
import { AbandonadosSkeleton } from '@/components/abandonados/AbandonadosSkeleton'
import { BibliotecaPaginacao } from '@/components/jogos/BibliotecaPaginacao'
import { ExcluirAbandonadoDialog } from '@/components/abandonados/ExcluirAbandonadoDialog'
import { useRetomarAbandonado } from '@/components/abandonados/useRetomarAbandonado'
import { AbandonadosApiError } from '@/lib/services/abandonadosService'
import { ERRO_JOGO_NAO_ENCONTRADO } from '@/components/abandonados/abandonados.constants'
import type { JogoAbandonado, OrdenarAbandonados } from '@/types/abandonados'

export function AbandonadosPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const {
    jogos, meta, filtros, totalGeral, isLoading, carregado, aviso, limparAviso,
    isExcluirModalOpen, jogoParaExcluir,
    abrirModalCriacao, abrirModalEdicao, abrirModalExcluir, fecharModalExcluir,
    carregarJogos, carregarFiltros, carregarTotal, excluirJogo,
  } = useAbandonadosStore()

  const { retomarJogo } = useRetomarAbandonado()
  const [isExcluindo, setIsExcluindo] = useState(false)
  const [erroExclusao, setErroExclusao] = useState<string | null>(null)
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

  function handleAbrirModalExcluir(jogo: JogoAbandonado) {
    setErroExclusao(null)
    abrirModalExcluir(jogo)
  }

  function handleFecharExcluir() {
    setErroExclusao(null)
    fecharModalExcluir()
  }

  async function handleConfirmarExcluir() {
    if (!jogoParaExcluir) return
    setIsExcluindo(true)
    setErroExclusao(null)
    try {
      await excluirJogo(jogoParaExcluir.id)
      handleFecharExcluir()
    } catch (err) {
      if (err instanceof AbandonadosApiError) {
        if (err.status === 404 || err.codigo === 'abandonados.nao_encontrado') {
          setErroExclusao(ERRO_JOGO_NAO_ENCONTRADO)
        } else {
          setErroExclusao(err.message || 'Não foi possível excluir o jogo. Tente novamente.')
        }
      } else if (err instanceof Error) {
        setErroExclusao(err.message)
      } else {
        setErroExclusao('Não foi possível excluir o jogo. Tente novamente.')
      }
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
        <AbandonadosCabecalho
          total={totalExibicao}
          onAbandonar={() => abrirModalCriacao()}
        />

        {aviso && (
          <div
            role="status"
            className={
              'flex items-center justify-between gap-3 rounded-lg border border-[var(--border)] ' +
              'bg-[var(--bg-surface)] px-4 py-3 text-sm text-[var(--text-primary)]'
            }
          >
            <span>{aviso}</span>
            <button
              type="button"
              onClick={limparAviso}
              aria-label="Fechar aviso"
              className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
            >
              <X className="h-4 w-4" aria-hidden="true" />
            </button>
          </div>
        )}

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

        {!carregado || isLoading ? (
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
              onRetomar={retomarJogo}
              onEditar={abrirModalEdicao}
              onExcluir={handleAbrirModalExcluir}
            />
            <BibliotecaPaginacao meta={meta} onMudarPagina={handleMudarPagina} />
          </div>
        )}
      </main>

      <ExcluirAbandonadoDialog
        open={isExcluirModalOpen}
        onOpenChange={(open) => !open && handleFecharExcluir()}
        onConfirm={handleConfirmarExcluir}
        jogoNome={jogoParaExcluir?.nome}
        isLoading={isExcluindo}
        error={erroExclusao}
      />
    </div>
  )
}
