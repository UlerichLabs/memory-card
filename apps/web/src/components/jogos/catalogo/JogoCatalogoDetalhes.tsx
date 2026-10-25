import type { ReactNode } from 'react'
import { ArrowLeft, Gamepad2 } from 'lucide-react'
import { JogoDetalheSobre } from '@/components/jogos/detalhe/JogoDetalheSobre'
import { ConsoleBadge } from '@/components/jogos/ConsoleBadge'
import { formatarCapaIGDB } from '@/lib/utils'
import type { IGDBJogoSugestao } from '@/types/jogos'

export interface JogoCatalogoDetalhesProps {
  jogo: IGDBJogoSugestao
  onVoltar: () => void
  botaoAcao?: ReactNode
}

export function JogoCatalogoDetalhes({
  jogo,
  onVoltar,
  botaoAcao,
}: JogoCatalogoDetalhesProps) {
  const capaUrl = formatarCapaIGDB(jogo.cover?.url, 't_cover_big_2x')
  const anoLancamento = jogo.first_release_date
    ? new Date(jogo.first_release_date * 1000).getUTCFullYear()
    : null

  return (
    <div className="grid grid-cols-1 items-start gap-8 sm:gap-10 lg:grid-cols-[320px_minmax(0,1fr)] lg:gap-12">
      <aside className="flex flex-col items-center sm:items-start gap-3.5">
        <div className="relative aspect-[3/4] w-[200px] sm:w-[240px] lg:w-[320px] lg:h-[427px] overflow-hidden rounded-[14px] lg:rounded-[16px] border border-[var(--detalhe-capa-border)] bg-[var(--detalhe-capa-bg)]">
          {capaUrl ? (
            <img src={capaUrl} alt={jogo.name} className="h-full w-full object-cover" />
          ) : (
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
              type="button"
              onClick={onVoltar}
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
              {jogo.platforms?.map((plat) => (
                <ConsoleBadge key={plat.id} nome={plat.name} tamanho="md" />
              ))}
              {jogo.genres?.map((gen) => (
                <span
                  key={gen.id}
                  className="inline-flex h-[28px] sm:h-[30px] items-center rounded-full border border-[var(--detalhe-chip-genero-border)] bg-[var(--detalhe-chip-genero-bg)] px-3 text-[12px] sm:text-[13px] text-[var(--detalhe-chip-genero-text)]"
                >
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
  )
}
