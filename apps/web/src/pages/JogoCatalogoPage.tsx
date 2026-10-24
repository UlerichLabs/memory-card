import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams, useLocation } from 'react-router-dom'
import { ArrowLeft, Check, Gamepad2 } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { JogoDetalheSkeleton } from '@/components/jogos/detalhe/JogoDetalheSkeleton'
import { JogoDetalheSobre } from '@/components/jogos/detalhe/JogoDetalheSobre'
import { ConsoleBadge } from '@/components/jogos/ConsoleBadge'
import { jogosService, JogosApiError } from '@/lib/services/jogosService'
import { useJogandoStore } from '@/stores/jogandoStore'
import { encontrarJogoEmAndamento, hojeIso } from '@/lib/jogandoUtils'
import { formatarCapaIGDB } from '@/lib/utils'
import { useDocumentTitle } from '@/lib/useDocumentTitle'
import { ApiError } from '@/lib/api'
import type { IGDBJogoSugestao } from '@/types/jogos'

const ERROS_JOGANDO: Record<string, string> = {
  'jogando.nome_obrigatorio': 'O nome do jogo é obrigatório.', 'jogando.nome_muito_longo': 'O nome deve ter no máximo 200 caracteres.',
  'jogando.iniciado_em_obrigatorio': 'A data de início é obrigatória.', 'jogando.iniciado_em_invalido': 'Data de início inválida.',
  'jogando.iniciado_em_futuro': 'A data de início não pode estar no futuro.', 'jogando.entrada_invalida': 'Dados de entrada inválidos.',
  'jogando.item_duplicado': 'Este jogo já está em andamento.',
}

function obterMensagemErro(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 409) return ERROS_JOGANDO['jogando.item_duplicado']
    if (err.status === 404) return 'Registro não encontrado.'
    if (err.codigo && ERROS_JOGANDO[err.codigo]) return ERROS_JOGANDO[err.codigo]
    if (err.status === 400) return 'Dados de entrada inválidos.'
  }
  return 'Não foi possível iniciar o jogo. Tente novamente.'
}

export function JogoCatalogoPage() {
  const { igdbId } = useParams<{ igdbId: string }>()
  const navigate = useNavigate(), location = useLocation()
  const { jogos, criar } = useJogandoStore()
  const [jogo, setJogo] = useState<IGDBJogoSugestao | null>(null)
  const [carregando, setCarregando] = useState(true)
  const [erroStatus, setErroStatus] = useState<'notFound' | 'network' | null>(null)
  const [isIniciando, setIsIniciando] = useState(false)
  const [sucessoInline, setSucessoInline] = useState(false)
  const [erroAcao, setErroAcao] = useState<string | null>(null)

  useDocumentTitle(jogo?.name ?? 'Jogo')
  const state = location.state as { voltarPara?: string } | null, handleVoltar = () => (state?.voltarPara ? navigate(state.voltarPara) : navigate('/listas'))

  const carregarJogo = useCallback(async (signal?: AbortSignal) => {
    const numId = Number(igdbId)
    if (!igdbId || isNaN(numId) || numId <= 0) {
      setErroStatus('notFound'); setCarregando(false); return
    }
    setCarregando(true); setErroStatus(null)
    try {
      const dados = await jogosService.obterDetalhesIGDB(numId)
      if (!signal?.aborted) setJogo(dados)
    } catch (err) {
      if (signal?.aborted) return
      const isNotFound = err instanceof JogosApiError && (err.status === 404 || err.status === 400 || err.codigo === 'jogos.not_found' || err.codigo === 'jogos.invalid_id')
      setErroStatus(isNotFound ? 'notFound' : 'network')
    } finally {
      if (!signal?.aborted) setCarregando(false)
    }
  }, [igdbId])

  useEffect(() => {
    const ctrl = new AbortController()
    carregarJogo(ctrl.signal)
    return () => ctrl.abort()
  }, [carregarJogo])

  const jaEmAndamento = Boolean(jogo && encontrarJogoEmAndamento(jogos, { igdbId: jogo.id, nome: jogo.name }))
  const emAndamento = jaEmAndamento || sucessoInline

  async function handleJogarAgora() {
    if (!jogo || isIniciando || emAndamento) return
    setIsIniciando(true); setErroAcao(null)
    try {
      await criar({
        nome: jogo.name, igdb_id: jogo.id,
        igdb_capa_url: formatarCapaIGDB(jogo.cover?.url) || null, iniciado_em: hojeIso(),
      })
      setSucessoInline(true)
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) { navigate('/login'); return }
      setErroAcao(obterMensagemErro(err))
    } finally {
      setIsIniciando(false)
    }
  }

  const capaUrl = formatarCapaIGDB(jogo?.cover?.url, 't_cover_big_2x')
  const anoLancamento = jogo?.first_release_date ? new Date(jogo.first_release_date * 1000).getUTCFullYear() : null, rotuloBotao = isIniciando ? 'Iniciando...' : emAndamento ? 'Em andamento' : 'Jogar agora'

  const botaoAcao = (
    <button
      type="button" onClick={handleJogarAgora} disabled={isIniciando || emAndamento}
      className={`btn-primario flex h-[40px] items-center gap-2 px-4 text-[14px] font-bold ${emAndamento ? 'btn-primario-disabled' : ''}`}
    >
      <Gamepad2 className="h-4 w-4" aria-hidden="true" />
      <span>{rotuloBotao}</span>
    </button>
  )

  return (
    <div className="min-h-full w-full bg-[var(--bg-primary)] overflow-x-hidden">
      <Topbar />
      <div className="flex h-[56px] items-center justify-between border-b border-[var(--detalhe-mobile-header-border)] bg-[var(--detalhe-mobile-header-bg)] px-3 sm:hidden">
        <div className="flex items-center gap-1">
          <button
            type="button" onClick={handleVoltar} aria-label="Voltar para o desafio"
            className="flex h-[44px] w-[44px] items-center justify-center rounded-[8px] text-[var(--detalhe-link-voltar)] hover:text-[var(--text-primary)]"
          >
            <ArrowLeft className="h-5 w-5" aria-hidden="true" />
          </button>
          <span className="text-[15px] font-semibold text-[var(--text-primary)]">Desafio</span>
        </div>
        {jogo && <div className="flex items-center gap-2">{botaoAcao}</div>}
      </div>

      <main className="mx-auto max-w-7xl px-4 py-5 sm:px-10 sm:py-7 lg:py-12">
        {carregando ? <JogoDetalheSkeleton /> : erroStatus === 'notFound' ? (
          <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Jogo não encontrado</h2>
            <p className="text-sm text-[var(--text-secondary)]">O jogo solicitado não foi encontrado no catálogo.</p>
            <button type="button" onClick={handleVoltar} className="btn-primario px-4 py-2 text-sm">Voltar para o desafio</button>
          </div>
        ) : erroStatus === 'network' ? (
          <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Erro ao carregar registro</h2>
            <p className="text-sm text-[var(--text-secondary)]">Não foi possível carregar as informações do jogo. Verifique sua conexão.</p>
            <button type="button" onClick={() => carregarJogo()} className="btn-primario px-4 py-2 text-sm">Tentar novamente</button>
          </div>
        ) : jogo ? (
          <div className="flex flex-col gap-4">
            {sucessoInline && (
              <div className="flex items-center gap-2 rounded-lg border border-[var(--border)] bg-[var(--bg-surface)] p-3 text-[13px] font-medium text-[var(--success)]">
                <Check className="h-4 w-4 shrink-0" aria-hidden="true" />
                <span>Adicionado aos jogos em andamento!</span>
              </div>
            )}
            {erroAcao && (
              <div className="flex items-center rounded-lg border border-[var(--danger)]/30 bg-[var(--bg-surface)] p-3 text-[13px] font-medium text-[var(--danger)]">
                <span>{erroAcao}</span>
              </div>
            )}
            <div className="grid grid-cols-1 items-start gap-8 sm:gap-10 lg:grid-cols-[320px_minmax(0,1fr)] lg:gap-12">
              <aside className="flex flex-col items-center sm:items-start gap-3.5">
                <div className="relative aspect-[3/4] w-[200px] sm:w-[240px] lg:w-[320px] lg:h-[427px] overflow-hidden rounded-[14px] lg:rounded-[16px] border border-[var(--detalhe-capa-border)] bg-[var(--detalhe-capa-bg)]">
                  {capaUrl ? <img src={capaUrl} alt={jogo.name} className="h-full w-full object-cover" /> : (
                    <div className="flex h-full w-full flex-col items-center justify-center p-6 text-[var(--detalhe-capa-icon-sem-arte)]">
                      <Gamepad2 className="h-12 w-12" aria-hidden="true" />
                    </div>
                  )}
                </div>
              </aside>

              <div className="flex min-w-0 flex-col gap-6">
                <div className="flex flex-col gap-6">
                  <div className="hidden items-center justify-between sm:flex">
                    <button
                      type="button" onClick={handleVoltar}
                      className="flex h-[40px] items-center gap-1.5 rounded-[10px] px-3 pl-2 text-[14px] font-medium text-[var(--detalhe-link-voltar)] transition-colors hover:text-[var(--text-primary)]"
                    >
                      <ArrowLeft className="h-[18px] w-[18px]" aria-hidden="true" />
                      <span>Voltar para o desafio</span>
                    </button>
                    {botaoAcao}
                  </div>

                  <div className="flex flex-col gap-3.5">
                    <h1 className="break-words text-[26px] font-bold tracking-[-0.015em] text-[var(--text-primary)] leading-[1.15] sm:text-[36px]">
                      {jogo.name}
                    </h1>

                    <div className="flex flex-wrap items-center gap-2">
                      {jogo.platforms?.map((plat) => <ConsoleBadge key={plat.id} nome={plat.name} tamanho="md" />)}
                      {jogo.genres?.map((gen) => (
                        <span key={gen.id} className="inline-flex h-[28px] sm:h-[30px] items-center rounded-full border border-[var(--detalhe-chip-genero-border)] bg-[var(--detalhe-chip-genero-bg)] px-3 text-[12px] sm:text-[13px] text-[var(--detalhe-chip-genero-text)]">
                          {gen.name}
                        </span>
                      ))}
                      {anoLancamento && (
                        <span className="inline-flex h-[28px] sm:h-[30px] items-center rounded-full border border-dashed border-[var(--detalhe-chip-tipo-border)] bg-transparent px-3 text-[12px] sm:text-[13px] text-[var(--detalhe-chip-tipo-text)]">
                          {anoLancamento}
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <JogoDetalheSobre descricao={jogo.summary} />
              </div>
            </div>
          </div>
        ) : null}
      </main>
    </div>
  )
}
