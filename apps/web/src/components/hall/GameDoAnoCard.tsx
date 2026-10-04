import { Link } from 'react-router-dom'
import { Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB } from '@/lib/utils'
import { NotaBadge } from '@/components/jogos/NotaBadge'
import { ConsoleBadge } from '@/components/jogos/ConsoleBadge'

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
  const textoJogos = totalJogos === 1 ? '1 jogo' : `${totalJogos} jogos`

  return (
    <article className="flex h-full flex-col gap-3 rounded-[12px] border border-[var(--hall-card-border)] bg-[var(--hall-card-bg)] p-3">
      <div className="flex items-center justify-between">
        <span className="text-[20px] font-bold tabular-nums text-[var(--hall-ano-ouro)]">
          {ano}
        </span>
        {ehAnoAtual && (
          <span className="rounded-full border border-[var(--hall-andamento-border)] bg-[var(--hall-andamento-bg)] px-2 py-0.5 text-[10px] font-semibold text-[var(--hall-andamento-text)]">
            Em andamento
          </span>
        )}
      </div>

      <Link
        to={`/biblioteca/${jogo.id}`}
        aria-label={`Ver detalhes de ${jogo.nome}`}
        className="group flex min-w-0 flex-col gap-2"
      >
        <div className="relative aspect-[3/4] w-full overflow-hidden rounded-[8px] border border-[var(--hall-ouro)] bg-[var(--bg-surface-alt)] shadow-[0_0_8px_1px_var(--hall-ouro)] transition-transform group-hover:-translate-y-0.5">
          {capaUrl ? (
            <img
              src={capaUrl}
              alt={jogo.nome}
              className="h-full w-full object-cover transition-opacity group-hover:opacity-90"
              loading="lazy"
            />
          ) : (
            <div className="flex h-full w-full flex-col items-center justify-center gap-2 p-3 text-center text-[var(--text-muted)]">
              <Gamepad2 className="h-5 w-5 opacity-40" aria-hidden="true" />
              <span className="line-clamp-3 text-[11px] font-medium">{jogo.nome}</span>
            </div>
          )}
          <div className="absolute right-1.5 top-1.5">
            <NotaBadge nota={jogo.nota} tamanho="sm" />
          </div>
        </div>

        <h3
          className="min-h-[2.6em] line-clamp-2 text-[13px] font-bold leading-[1.3] text-[var(--text-primary)] transition-colors group-hover:text-[var(--hall-ano-ouro)]"
          title={jogo.nome}
        >
          {jogo.nome}
        </h3>
        <ConsoleBadge nome={jogo.console} variante="solido" />
      </Link>

      <div className="mt-auto flex items-center justify-between border-t border-[var(--hall-divider)] pt-2">
        <span className="text-[10px] text-[var(--hall-muted)]">
          {textoJogos}
        </span>
        <button
          type="button"
          onClick={() => onTrocar(ano, jogo)}
          className="flex h-6 items-center rounded-[6px] border border-[var(--hall-btn-trocar-border)] bg-transparent px-2 text-[10px] font-medium text-[var(--hall-btn-trocar-text)] transition-colors hover:bg-[var(--bg-surface-alt)]"
        >
          Trocar
        </button>
      </div>
    </article>
  )
}
