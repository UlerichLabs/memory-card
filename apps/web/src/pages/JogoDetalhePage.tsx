import { useEffect, useRef, useState, useCallback } from 'react'
import { useParams, useNavigate, useLocation } from 'react-router-dom'
import { ArrowLeft, Pencil, Trash2 } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { useJogosStore } from '@/stores/jogosStore'
import { JogosApiError } from '@/lib/services/jogosService'
import { isoParaDataPt } from '@/lib/utils'
import type { JogoZeradoDTO } from '@/types/jogos'
import { ExcluirJogoDialog } from '@/components/jogos/ExcluirJogoDialog'
import { JogoDetalheCabecalho } from '@/components/jogos/detalhe/JogoDetalheCabecalho'
import { JogoDetalheColunaEsquerda } from '@/components/jogos/detalhe/JogoDetalheColunaEsquerda'
import { JogoDetalheStats } from '@/components/jogos/detalhe/JogoDetalheStats'
import { JogoDetalheReview } from '@/components/jogos/detalhe/JogoDetalheReview'
import { JogoDetalheSobre } from '@/components/jogos/detalhe/JogoDetalheSobre'
import { JogoDetalheSkeleton } from '@/components/jogos/detalhe/JogoDetalheSkeleton'

export function JogoDetalhePage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const { obterJogoPorId, abrirModalEdicao, excluirJogo, isModalOpen } = useJogosStore()

  const [jogo, setJogo] = useState<JogoZeradoDTO | null>(null)
  const [carregando, setCarregando] = useState(true)
  const [erroStatus, setErroStatus] = useState<'notFound' | 'network' | null>(null)
  const [excluindoAberto, setExcluindoAberto] = useState(false)
  const [isExcluindo, setIsExcluindo] = useState(false)

  const fromSearch = (location.state as { from?: string } | null)?.from || ''
  const handleVoltar = () => navigate(`/biblioteca${fromSearch}`)

  const carregarDetalhe = useCallback(async (signal?: AbortSignal) => {
    const numId = Number(id)
    if (!id || isNaN(numId) || numId <= 0) {
      setErroStatus('notFound')
      setCarregando(false)
      return
    }
    setCarregando(true)
    setErroStatus(null)
    try {
      const dados = await obterJogoPorId(numId, signal)
      if (!signal?.aborted) setJogo(dados)
    } catch (err) {
      if (signal?.aborted) return
      if (err instanceof JogosApiError && (err.status === 404 || err.status === 400 || err.codigo === 'jogos.not_found' || err.codigo === 'jogos.invalid_id')) {
        setErroStatus('notFound')
      } else {
        setErroStatus('network')
      }
    } finally {
      if (!signal?.aborted) setCarregando(false)
    }
  }, [id, obterJogoPorId])

  useEffect(() => {
    const ctrl = new AbortController()
    carregarDetalhe(ctrl.signal)
    return () => ctrl.abort()
  }, [carregarDetalhe])

  const prevModalRef = useRef(isModalOpen)
  useEffect(() => {
    if (prevModalRef.current && !isModalOpen) carregarDetalhe()
    prevModalRef.current = isModalOpen
  }, [isModalOpen, carregarDetalhe])

  useEffect(() => {
    if (jogo?.nome) document.title = jogo.nome
    return () => { document.title = 'Memory Card' }
  }, [jogo?.nome])

  async function handleConfirmarExclusao() {
    if (!jogo) return
    setIsExcluindo(true)
    try {
      await excluirJogo(jogo.id)
      navigate(`/biblioteca${fromSearch}`, { replace: true })
    } finally {
      setIsExcluindo(false)
    }
  }

  const dataCriacao = jogo ? isoParaDataPt(jogo.created_at) : ''
  const dataAtualizacao = jogo ? isoParaDataPt(jogo.updated_at) : ''
  const datasTexto = !dataAtualizacao || dataCriacao === dataAtualizacao
    ? `Criado em ${dataCriacao}`
    : `Criado em ${dataCriacao} · atualizado em ${dataAtualizacao}`

  return (
    <div className="min-h-svh w-full bg-[var(--bg-primary)] overflow-x-hidden">
      <Topbar />

      <div className="flex h-[56px] items-center justify-between border-b border-[var(--detalhe-mobile-header-border)] bg-[var(--detalhe-mobile-header-bg)] px-3 sm:hidden">
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={handleVoltar}
            aria-label="Voltar para a Biblioteca"
            className="flex h-[44px] w-[44px] items-center justify-center rounded-[8px] text-[var(--detalhe-link-voltar)] hover:text-[var(--text-primary)]"
          >
            <ArrowLeft className="h-5 w-5" aria-hidden="true" />
          </button>
          <span className="text-[15px] font-semibold text-[var(--text-primary)]">Biblioteca</span>
        </div>
        {jogo && (
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={() => abrirModalEdicao(jogo)}
              aria-label="Editar jogo"
              className="flex h-[44px] w-[44px] items-center justify-center rounded-[8px] text-[var(--detalhe-btn-editar-text)] hover:text-[var(--text-primary)]"
            >
              <Pencil className="h-4 w-4" aria-hidden="true" />
            </button>
            <button
              type="button"
              onClick={() => setExcluindoAberto(true)}
              aria-label="Excluir jogo"
              className="flex h-[44px] w-[44px] items-center justify-center rounded-[8px] text-[var(--detalhe-btn-excluir-text)] hover:opacity-80"
            >
              <Trash2 className="h-4 w-4" aria-hidden="true" />
            </button>
          </div>
        )}
      </div>

      <main className="mx-auto max-w-7xl px-4 py-5 pb-24 sm:px-10 sm:py-7 lg:py-12 md:pb-0">
        {carregando ? (
          <JogoDetalheSkeleton />
        ) : erroStatus === 'notFound' ? (
          <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Jogo não encontrado</h2>
            <p className="text-sm text-[var(--text-secondary)]">O registro solicitado não existe ou pertence a outro usuário.</p>
            <button
              type="button"
              onClick={handleVoltar}
              className="rounded-[10px] bg-[var(--accent)] px-4 py-2 text-sm font-bold text-[var(--accent-foreground)] transition-opacity hover:opacity-90"
            >
              Voltar para a Biblioteca
            </button>
          </div>
        ) : erroStatus === 'network' ? (
          <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Erro ao carregar registro</h2>
            <p className="text-sm text-[var(--text-secondary)]">Não foi possível carregar os detalhes do jogo. Verifique sua conexão.</p>
            <button
              type="button"
              onClick={() => carregarDetalhe()}
              className="rounded-[10px] bg-[var(--accent)] px-4 py-2 text-sm font-bold text-[var(--accent-foreground)] transition-opacity hover:opacity-90"
            >
              Tentar novamente
            </button>
          </div>
        ) : jogo ? (
          <div className="grid grid-cols-1 items-start gap-8 sm:gap-10 lg:grid-cols-[320px_minmax(0,1fr)] lg:gap-12">
            <JogoDetalheColunaEsquerda jogo={jogo} />

            <div className="flex flex-col gap-6">
              <JogoDetalheCabecalho
                jogo={jogo}
                onVoltar={handleVoltar}
                onEditar={() => abrirModalEdicao(jogo)}
                onExcluir={() => setExcluindoAberto(true)}
              />

              <JogoDetalheStats jogo={jogo} />

              <JogoDetalheReview
                review={jogo.review}
                onAdicionarReview={() => abrirModalEdicao(jogo)}
              />

              <JogoDetalheSobre descricao={jogo.igdb_descricao} />

              <div className="flex sm:hidden flex-col gap-0.5 pt-2 text-[12px] tabular-nums text-[var(--detalhe-text-meta)]">
                <p className="font-medium">{`Registro #${jogo.numero ?? jogo.id}`}</p>
                <p>{datasTexto}</p>
              </div>
            </div>
          </div>
        ) : null}
      </main>

      <ExcluirJogoDialog
        open={excluindoAberto}
        onOpenChange={setExcluindoAberto}
        jogoNome={jogo?.nome}
        isLoading={isExcluindo}
        onConfirm={handleConfirmarExclusao}
      />
    </div>
  )
}
