import { useEffect, useState, useRef } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { ListasProvider, useListasStore } from '@/stores/listasStore'
import { ListasSidebar } from '@/components/listas/ListasSidebar'
import { ListaCabecalho } from '@/components/listas/ListaCabecalho'
import { DesafioMetaBox } from '@/components/listas/DesafioMetaBox'
import { DesafioGrade } from '@/components/listas/DesafioGrade'
import { FilaLista } from '@/components/listas/FilaLista'
import { NovaListaDialog } from '@/components/listas/NovaListaDialog'
import { ExcluirListaDialog } from '@/components/listas/ExcluirListaDialog'
import { ListasVazio } from '@/components/listas/ListasVazio'
import { ListasSkeleton } from '@/components/listas/ListasSkeleton'

function ListasPageContent() {
  const { id } = useParams<{ id?: string }>()
  const navigate = useNavigate()
  const {
    listas,
    listaAberta,
    isLoading,
    isLoadingDetalhe,
    carregarListas,
    abrirLista,
    abrirModalCriar,
  } = useListasStore()

  const [naoEncontrada, setNaoEncontrada] = useState(false)
  const [inicializado, setInicializado] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  const idInvalido = Boolean(id && isNaN(parseInt(id, 10)))
  const estaNaoEncontrada = naoEncontrada || idInvalido

  useEffect(() => {
    let cancelado = false
    carregarListas().then((ordenadas) => {
      if (cancelado) return
      setInicializado(true)
      if (!id && ordenadas.length > 0) {
        navigate(`/listas/${ordenadas[0].id}`, { replace: true })
      }
    }).catch(() => {
      if (!cancelado) setInicializado(true)
    })
    return () => { cancelado = true }
  }, [carregarListas, id, navigate])

  useEffect(() => {
    if (!id) return
    const listaId = parseInt(id, 10)
    if (isNaN(listaId)) return
    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller
    abrirLista(listaId, controller.signal)
      .then(() => setNaoEncontrada(false))
      .catch(() => {
        if (!controller.signal.aborted) {
          setNaoEncontrada(true)
        }
      })
    return () => { controller.abort() }
  }, [id, abrirLista])

  if (!inicializado && isLoading) {
    return (
      <div className="min-h-screen bg-[var(--bg-primary)]">
        <Topbar />
        <ListasSkeleton />
      </div>
    )
  }

  const handleSelectLista = (selectedId: number) => {
    navigate(`/listas/${selectedId}`)
  }

  const renderPainel = () => {
    if (estaNaoEncontrada) {
      return (
        <section className="flex flex-col items-center justify-center gap-3 rounded-2xl border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] p-12 text-center">
          <h2 className="text-[20px] font-bold text-[var(--text-primary)]">Lista não encontrada</h2>
          <p className="text-[14px] text-[var(--lista-text-secondary)]">A lista que você tentou acessar não existe ou foi removida.</p>
          <Link to="/listas" onClick={() => setNaoEncontrada(false)} className="text-[14px] font-semibold text-[var(--lista-icon-fila)] hover:underline">
            Voltar para Listas e Desafios
          </Link>
        </section>
      )
    }

    if (listas.length === 0) {
      return <ListasVazio onNovaLista={abrirModalCriar} />
    }

    if (!listaAberta) {
      return (
        <section className="flex flex-col gap-6 rounded-2xl border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] p-5 sm:p-7">
          {isLoadingDetalhe ? (
            <div className="h-64 animate-pulse rounded bg-[var(--lista-cover-bg)]" />
          ) : (
            <div className="py-12 text-center text-[14px] text-[var(--lista-text-muted)]">
              Selecione uma lista na barra lateral.
            </div>
          )}
        </section>
      )
    }

    return (
      <section className="flex flex-col gap-[22px] rounded-2xl border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] p-5 sm:p-7">
        <ListaCabecalho lista={listaAberta} />
        {listaAberta.tipo === 'desafio' ? (
          <>
            <DesafioMetaBox lista={listaAberta} />
            <DesafioGrade lista={listaAberta} />
          </>
        ) : (
          <FilaLista lista={listaAberta} />
        )}
      </section>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)]">
      <Topbar />
      <main className="mx-auto grid max-w-7xl grid-cols-1 items-start gap-8 px-4 py-8 sm:px-6 md:grid-cols-[300px_minmax(0,1fr)] lg:px-8 xl:p-[32px_40px_48px]">
        <ListasSidebar
          listas={listas}
          selectedId={listaAberta?.id ?? null}
          onSelect={handleSelectLista}
          onNovaLista={abrirModalCriar}
        />
        {renderPainel()}
      </main>

      <NovaListaDialog />
      <ExcluirListaDialog />
    </div>
  )
}

export function ListasPage() {
  return (
    <ListasProvider>
      <ListasPageContent />
    </ListasProvider>
  )
}
