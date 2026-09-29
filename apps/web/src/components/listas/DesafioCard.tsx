import { Link } from 'react-router-dom'
import { Check, X } from 'lucide-react'
import { NotaBadge } from '@/components/jogos/NotaBadge'
import { formatarCapaIGDB } from '@/lib/utils'
import { obterIniciaisJogo } from './listas.utils'
import type { ListaItem } from '@/types/listas'

export interface DesafioCardProps {
  item: ListaItem
  isManual?: boolean
  onRemover?: (id: number) => void
}

export function DesafioCard({ item, isManual = false, onRemover }: DesafioCardProps) {
  const { zerado, jogo_zerado, nome, console: consoleName, ano_lancamento, igdb_capa_url, origem } = item
  const capaUrl = formatarCapaIGDB(igdb_capa_url ?? undefined, 't_cover_big')
  const capaUrl2x = formatarCapaIGDB(igdb_capa_url ?? undefined, 't_cover_big_2x')
  const iniciais = obterIniciaisJogo(nome)

  let metaTexto = ''
  if (origem === 'regra' && jogo_zerado?.finalizado_em) {
    const anoZerado = jogo_zerado.finalizado_em.slice(0, 4)
    metaTexto = [consoleName, `zerado em ${anoZerado}`].filter(Boolean).join(' · ')
  } else {
    metaTexto = [ano_lancamento, consoleName].filter(Boolean).join(' · ')
  }

  const cardContent = (
    <div
      className={`group relative flex flex-col gap-2 rounded-[10px] ${
        zerado ? 'cursor-pointer' : 'cursor-default'
      }`}
    >
      <div
        className={`relative aspect-[3/4] w-full overflow-hidden rounded-[10px] bg-[var(--lista-cover-bg)] transition-transform duration-200 ${
          zerado
            ? 'border border-[var(--lista-card-zerado-border)] group-hover:-translate-y-1'
            : 'border border-dashed border-[var(--lista-card-pendente-border)] opacity-55'
        }`}
      >
        {capaUrl ? (
          <img
            src={capaUrl}
            srcSet={capaUrl2x ? `${capaUrl} 1x, ${capaUrl2x} 2x` : undefined}
            alt={nome}
            className="h-full w-full object-cover"
            loading="lazy"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <span
              className={`text-[26px] font-bold select-none ${
                zerado ? 'text-[var(--lista-initials-zerado)]' : 'text-[var(--lista-initials-pendente)]'
              }`}
            >
              {iniciais}
            </span>
          </div>
        )}

        {zerado && (
          <>
            <div className="absolute left-1.5 top-1.5 flex h-6 w-6 items-center justify-center rounded-full bg-[var(--lista-progress-fill)] text-[var(--ouro-jogo-ano-text)] shadow">
              <Check className="h-3.5 w-3.5 stroke-[3]" />
            </div>
            {jogo_zerado && (
              <div className="absolute right-1.5 top-1.5">
                <NotaBadge nota={jogo_zerado.nota} tamanho="sm" />
              </div>
            )}
          </>
        )}

        {!zerado && (
          <span className="absolute bottom-1.5 left-1.5 rounded-full border border-[var(--lista-pill-pendente-border)] bg-[var(--lista-pill-pendente-bg)] px-[7px] py-[3px] text-[10px] font-semibold text-[var(--lista-pill-pendente-text)]">
            Pendente
          </span>
        )}

        {isManual && onRemover && (
          <button
            type="button"
            aria-label={`Remover ${nome}`}
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              onRemover(item.id)
            }}
            className="absolute right-1.5 top-1.5 flex h-9 w-9 items-center justify-center rounded-lg bg-[var(--lista-cover-bg)]/80 text-[var(--lista-text-muted)] opacity-0 transition-opacity hover:text-[var(--danger)] focus:opacity-100 group-hover:opacity-100"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      <div className="flex flex-col gap-0.5">
        <span
          title={nome}
          className={`line-clamp-2 text-[13px] font-semibold leading-tight ${
            zerado ? 'text-[var(--lista-text-bright)]' : 'text-[var(--lista-text-muted)]'
          }`}
        >
          {nome}
        </span>
        {metaTexto && (
          <span className="truncate text-[11px] text-[var(--lista-text-dim)]">
            {metaTexto}
          </span>
        )}
      </div>
    </div>
  )

  if (zerado && jogo_zerado) {
    return (
      <Link to={`/biblioteca/${jogo_zerado.id}`} className="block focus:outline-none">
        {cardContent}
      </Link>
    )
  }

  return cardContent
}
