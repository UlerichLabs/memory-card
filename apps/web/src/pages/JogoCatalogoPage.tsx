import { useParams, useNavigate, useLocation } from 'react-router-dom'
import { ArrowLeft, Check } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { JogoDetalheSkeleton } from '@/components/jogos/detalhe/JogoDetalheSkeleton'
import { JogoCatalogoDetalhes } from '@/components/jogos/catalogo/JogoCatalogoDetalhes'
import { JogarAgoraBotao } from '@/components/jogos/catalogo/JogarAgoraBotao'
import { useJogoCatalogo } from '@/hooks/useJogoCatalogo'
import { useDocumentTitle } from '@/lib/useDocumentTitle'

export function JogoCatalogoPage() {
  const { igdbId } = useParams<{ igdbId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const state = location.state as { voltarPara?: string } | null

  const {
    jogo,
    carregando,
    erroStatus,
    recarregar,
    jogarAgora,
    isIniciando,
    emAndamento,
    sucesso,
    erroAcao,
  } = useJogoCatalogo(igdbId)

  useDocumentTitle(jogo?.name ?? 'Jogo')

  const handleVoltar = () => {
    if (state?.voltarPara) {
      navigate(state.voltarPara)
      return
    }
    navigate('/listas')
  }

  const botaoAcao = (
    <JogarAgoraBotao
      onClick={jogarAgora}
      isIniciando={isIniciando}
      emAndamento={emAndamento}
    />
  )

  return (
    <div className="min-h-full w-full bg-[var(--bg-primary)] overflow-x-hidden">
      <Topbar />
      <div className="flex h-[56px] items-center justify-between border-b border-[var(--detalhe-mobile-header-border)] bg-[var(--detalhe-mobile-header-bg)] px-3 sm:hidden">
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={handleVoltar}
            aria-label="Voltar para o desafio"
            className="flex h-[44px] w-[44px] items-center justify-center rounded-[8px] text-[var(--detalhe-link-voltar)] hover:text-[var(--text-primary)]"
          >
            <ArrowLeft className="h-5 w-5" aria-hidden="true" />
          </button>
          <span className="text-[15px] font-semibold text-[var(--text-primary)]">Desafio</span>
        </div>
        {jogo && <div className="flex items-center gap-2">{botaoAcao}</div>}
      </div>

      <main className="mx-auto max-w-7xl px-4 py-5 sm:px-10 sm:py-7 lg:py-12">
        {carregando ? (
          <JogoDetalheSkeleton />
        ) : erroStatus === 'notFound' ? (
          <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Jogo não encontrado</h2>
            <p className="text-sm text-[var(--text-secondary)]">O jogo solicitado não foi encontrado no catálogo.</p>
            <button type="button" onClick={handleVoltar} className="btn-primario px-4 py-2 text-sm">
              Voltar para o desafio
            </button>
          </div>
        ) : erroStatus === 'network' ? (
          <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Erro ao carregar registro</h2>
            <p className="text-sm text-[var(--text-secondary)]">
              Não foi possível carregar as informações do jogo. Verifique sua conexão.
            </p>
            <button type="button" onClick={recarregar} className="btn-primario px-4 py-2 text-sm">
              Tentar novamente
            </button>
          </div>
        ) : jogo ? (
          <div className="flex flex-col gap-4">
            {sucesso && (
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
            <JogoCatalogoDetalhes jogo={jogo} onVoltar={handleVoltar} botaoAcao={botaoAcao} />
          </div>
        ) : null}
      </main>
    </div>
  )
}
