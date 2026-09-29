import { Link } from 'react-router-dom'
import { Crown, Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB } from '@/lib/utils'
import { NotaBadge } from '@/components/jogos/NotaBadge'

export interface GameDoAnoCardProps {
  ano: number
  totalJogos: number
  jogo: JogoZeradoDTO
  onTrocar: (ano: number, jogo: JogoZeradoDTO) => void
}

export function GameDoAnoCard({ ano, totalJogos, jogo, onTrocar }: GameDoAnoCardProps) {
  const anoAtual = new Date().getFullYear()
  const ehAnoAtual = ano === anoAtual
  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url, 't_cover_big')
  const textoJogos = totalJogos === 1 ? '1 jogo zerado no ano' : `${totalJogos} jogos zerados no ano`

  return (
    <article className="flex flex-col justify-between gap-[14px] rounded-[16px] border border-[var(--hall-card-border)] bg-[var(--hall-card-bg)] p-[18px]">
      <div className="flex items-center justify-between">
        <span className="text-[32px] font-extrabold tracking-[-0.02em] tabular-nums text-[var(--hall-ano-ouro)]">
          {ano}
        </span>
        {ehAnoAtual && (
          <span className="rounded-full border border-[var(--hall-andamento-border)] bg-[var(--hall-andamento-bg)] px-[10px] py-[4px] text-[12px] font-semibold text-[var(--hall-andamento-text)]">
            Em andamento
          </span>
        )}
      </div>

      <Link
        to={`/biblioteca/${jogo.id}`}
        aria-label={`Ver detalhes de ${jogo.nome}`}
        className="group flex items-start gap-[14px] transition-transform hover:-translate-y-0.5"
      >
        <div className="relative h-[144px] w-[108px] shrink-0 overflow-visible">
          <div className="h-full w-full overflow-hidden rounded-[10px] border border-[var(--hall-ouro)] bg-[var(--bg-surface-alt)] shadow-[0_0_16px_1px_rgba(240,190,80,0.22)]">
            {capaUrl ? (
              <img
                src={capaUrl}
                alt={jogo.nome}
                className="h-full w-full object-cover transition-opacity group-hover:opacity-90"
                loading="lazy"
              />
            ) : (
              <div className="flex h-full w-full flex-col items-center justify-center p-2 text-center text-[var(--text-muted)]">
                <Gamepad2 className="h-7 w-7 opacity-40" aria-hidden="true" />
              </div>
            )}
          </div>
          <div
            className="absolute -left-2 -top-2 flex h-7 w-7 items-center justify-center rounded-full bg-[var(--hall-ouro)] shadow-md"
            aria-hidden="true"
          >
            <Crown className="h-3.5 w-3.5 fill-current text-[var(--hall-dark-icon)]" />
          </div>
        </div>

        <div className="flex min-w-0 flex-1 flex-col gap-2">
          <h3
            className="line-clamp-3 text-[15px] font-bold leading-[1.3] text-[var(--text-primary)] transition-colors group-hover:text-[var(--hall-ano-ouro)]"
            title={jogo.nome}
          >
            {jogo.nome}
          </h3>
          <p className="truncate text-[12px] text-[var(--hall-plataforma)]">
            {jogo.console}
          </p>
          <div className="self-start">
            <NotaBadge nota={jogo.nota} tamanho="sm" />
          </div>
        </div>
      </Link>

      <div className="flex items-center justify-between border-t border-[var(--hall-divider)] pt-[12px]">
        <span className="text-[12px] text-[var(--hall-muted)]">
          {textoJogos}
        </span>
        <button
          type="button"
          onClick={() => onTrocar(ano, jogo)}
          className="flex h-[32px] items-center rounded-[8px] border border-[var(--hall-btn-trocar-border)] bg-transparent px-[12px] text-[13px] font-medium text-[var(--hall-btn-trocar-text)] transition-colors hover:bg-[var(--bg-surface-alt)]"
        >
          Trocar
        </button>
      </div>
    </article>
  )
}
