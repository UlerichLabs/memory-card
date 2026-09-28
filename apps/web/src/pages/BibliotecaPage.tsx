import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { useJogosStore } from '@/stores/jogosStore'
import { BibliotecaFiltros } from '@/components/jogos/BibliotecaFiltros'
import { BibliotecaControles } from '@/components/jogos/BibliotecaControles'
import { BibliotecaGrade } from '@/components/jogos/BibliotecaGrade'
import { BibliotecaLista } from '@/components/jogos/BibliotecaLista'
import { BibliotecaPaginacao } from '@/components/jogos/BibliotecaPaginacao'
import { BibliotecaVazia } from '@/components/jogos/BibliotecaVazia'
import { BibliotecaSkeletons } from '@/components/jogos/BibliotecaSkeletons'
import { ExcluirJogoDialog } from '@/components/jogos/ExcluirJogoDialog'
import type { Dificuldade, JogoZeradoDTO, ListarJogosParams } from '@/types/jogos'

export function BibliotecaPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const {
    jogos,
    meta,
    filtros,
    isLoading,
    carregarJogos,
    carregarFiltros,
    limparBiblioteca,
    abrirModalRegistro,
    abrirModalEdicao,
    excluirJogo,
  } = useJogosStore()

  const [jogoParaExcluir, setJogoParaExcluir] = useState<JogoZeradoDTO | null>(null)
  const [isExcluindo, setIsExcluindo] = useState(false)
  const abortControllerRef = useRef<AbortController | null>(null)

  const busca = searchParams.get('busca') || ''
  const consoleVal = searchParams.get('console') || ''
  const generoVal = searchParams.get('genero') || ''
  const tipoVal = searchParams.get('tipo') || ''
  const anoVal = searchParams.get('ano') || ''
  const notaMinVal = Number(searchParams.get('nota_min')) || 1
  const notaMaxVal = Number(searchParams.get('nota_max')) || 11
  const dificuldadeVal = (searchParams.get('dificuldade') as Dificuldade) || undefined
  const pagina = Number(searchParams.get('pagina')) || 1
  const modo = (searchParams.get('modo') as 'grid' | 'list') || 'grid'

  useEffect(() => {
    carregarFiltros()
    return () => {
      abortControllerRef.current?.abort()
      limparBiblioteca()
    }
  }, [])

  useEffect(() => {
    abortControllerRef.current?.abort()
    const controller = new AbortController()
    abortControllerRef.current = controller

    const params: ListarJogosParams = { pagina, por_pagina: 24 }
    if (busca) params.busca = busca
    if (consoleVal) params.console = consoleVal
    if (generoVal) params.genero = generoVal
    if (tipoVal) params.tipo = tipoVal
    if (anoVal) params.ano = Number(anoVal)
    if (notaMinVal > 1) params.nota_min = notaMinVal
    if (notaMaxVal < 11) params.nota_max = notaMaxVal
    if (dificuldadeVal) params.dificuldade = dificuldadeVal

    carregarJogos(params, controller.signal).catch(() => {})
  }, [busca, consoleVal, generoVal, tipoVal, anoVal, notaMinVal, notaMaxVal, dificuldadeVal, pagina])

  const atualizarFiltros = (atualizacoes: Record<string, string | number | undefined>, resetPagina = true) => {
    const novos = new URLSearchParams(searchParams)
    if (resetPagina) novos.delete('pagina')
    Object.entries(atualizacoes).forEach(([k, v]) => {
      if (v === undefined || v === '' || (k === 'nota_min' && v === 1) || (k === 'nota_max' && v === 11) || (k === 'pagina' && v === 1)) {
        novos.delete(k)
      } else {
        novos.set(k, String(v))
      }
    })
    setSearchParams(novos)
  }

  const opcoesConsole = useMemo(() => [{ value: '', label: 'Todos os consoles' }, ...filtros.consoles.map((c) => ({ value: c, label: c }))], [filtros.consoles])
  const opcoesGenero = useMemo(() => [{ value: '', label: 'Todos os gêneros' }, ...filtros.generos.map((g) => ({ value: g, label: g }))], [filtros.generos])
  const opcoesTipo = useMemo(() => [{ value: '', label: 'Todos os tipos' }, ...filtros.tipos.map((t) => ({ value: t, label: t }))], [filtros.tipos])
  const opcoesAno = useMemo(() => [{ value: '', label: 'Todos os anos' }, ...filtros.anos.map((a) => ({ value: String(a), label: String(a) }))], [filtros.anos])

  const possuiFiltrosAtivos = Boolean(busca || consoleVal || generoVal || tipoVal || anoVal || notaMinVal > 1 || notaMaxVal < 11 || dificuldadeVal)

  async function handleConfirmarExclusao() {
    if (!jogoParaExcluir) return
    setIsExcluindo(true)
    try {
      await excluirJogo(jogoParaExcluir.id)
      setJogoParaExcluir(null)
    } finally {
      setIsExcluindo(false)
    }
  }

  return (
    <div className="min-h-svh bg-[var(--bg-primary)]">
      <Topbar />
      <main className="mx-auto max-w-7xl space-y-5 px-4 py-8 sm:px-6 lg:px-8">
        <header className="flex flex-col gap-1">
          <div className="flex items-center gap-3">
            <h1 className="text-[22px] font-bold text-[var(--biblioteca-text-primary)]">Biblioteca</h1>
            <span className="rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface)] px-2.5 py-0.5 text-xs font-semibold text-[var(--text-secondary)]">
              {meta.total} {meta.total === 1 ? 'jogo' : 'jogos'}
            </span>
          </div>
          <p className="text-[13px] text-[var(--biblioteca-text-muted)]">Seus registros de zeramentos concluídos.</p>
        </header>

        <BibliotecaFiltros
          busca={busca}
          consoleVal={consoleVal}
          generoVal={generoVal}
          tipoVal={tipoVal}
          anoVal={anoVal}
          notaMinVal={notaMinVal}
          notaMaxVal={notaMaxVal}
          dificuldadeVal={dificuldadeVal}
          opcoesConsole={opcoesConsole}
          opcoesGenero={opcoesGenero}
          opcoesTipo={opcoesTipo}
          opcoesAno={opcoesAno}
          onBuscaChange={(t) => atualizarFiltros({ busca: t })}
          onConsoleChange={(c) => atualizarFiltros({ console: c })}
          onGeneroChange={(g) => atualizarFiltros({ genero: g })}
          onTipoChange={(t) => atualizarFiltros({ tipo: t })}
          onAnoChange={(a) => atualizarFiltros({ ano: a })}
          onNotaChange={(min, max) => atualizarFiltros({ nota_min: min, nota_max: max })}
          onDificuldadeToggle={(d) => atualizarFiltros({ dificuldade: dificuldadeVal === d ? undefined : d })}
          onLimparFiltros={() => {
            const novos = new URLSearchParams()
            if (modo !== 'grid') novos.set('modo', modo)
            setSearchParams(novos)
          }}
        />

        {isLoading && jogos.length === 0 ? (
          <BibliotecaSkeletons modo={modo} />
        ) : jogos.length === 0 ? (
          <BibliotecaVazia
            possuiFiltrosAtivos={possuiFiltrosAtivos}
            onLimparFiltros={() => {
              const novos = new URLSearchParams()
              if (modo !== 'grid') novos.set('modo', modo)
              setSearchParams(novos)
            }}
            onRegistrarPrimeiroJogo={abrirModalRegistro}
          />
        ) : (
          <div className="space-y-4">
            <BibliotecaControles
              totalJogos={meta.total}
              modo={modo}
              onAlternarModo={(m) => atualizarFiltros({ modo: m === 'grid' ? undefined : m }, false)}
            />
            {modo === 'grid' ? (
              <BibliotecaGrade
                jogos={jogos}
                onEditar={abrirModalEdicao}
                onExcluir={setJogoParaExcluir}
                onDetalhes={(j) => navigate(`/biblioteca/${j.id}`)}
              />
            ) : (
              <BibliotecaLista
                jogos={jogos}
                onEditar={abrirModalEdicao}
                onExcluir={setJogoParaExcluir}
                onDetalhes={(j) => navigate(`/biblioteca/${j.id}`)}
              />
            )}
            <BibliotecaPaginacao
              meta={meta}
              onMudarPagina={(p) => atualizarFiltros({ pagina: p }, false)}
            />
          </div>
        )}
      </main>

      <ExcluirJogoDialog
        open={Boolean(jogoParaExcluir)}
        onOpenChange={(aberto) => !aberto && setJogoParaExcluir(null)}
        jogoNome={jogoParaExcluir?.nome}
        isLoading={isExcluindo}
        onConfirm={handleConfirmarExclusao}
      />
    </div>
  )
}
