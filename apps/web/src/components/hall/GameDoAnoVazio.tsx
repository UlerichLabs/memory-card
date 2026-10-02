import { Crown } from 'lucide-react'

export interface GameDoAnoVazioProps {
  ano: number
  totalJogos: number
  onEscolher: (ano: number) => void
}

export function GameDoAnoVazio({ ano, totalJogos, onEscolher }: GameDoAnoVazioProps) {
  const textoJogos = totalJogos === 1 ? `1 jogo zerado em ${ano}` : `${totalJogos} jogos zerados em ${ano}`

  return (
    <article className="flex min-h-[210px] flex-col justify-between gap-3 rounded-[10px] border border-dashed border-[var(--hall-card-vazio-border)] bg-transparent p-3">
      <div>
        <span className="text-[20px] font-bold tabular-nums text-[var(--hall-ano-vazio)]">
          {ano}
        </span>
      </div>

      <div className="flex min-h-[102px] flex-col items-center justify-center gap-1 py-2 text-center">
        <Crown className="h-5 w-5 text-[var(--hall-coroa-vazio)]" aria-hidden="true" />
        <span className="text-[12px] font-semibold text-[var(--hall-vazio-titulo)]">
          Sem Game do Ano
        </span>
        <span className="text-[10px] text-[var(--hall-muted)]">
          {textoJogos}
        </span>
      </div>

      <button
        type="button"
        onClick={() => onEscolher(ano)}
        className="flex h-7 w-full items-center justify-center rounded-[6px] border border-[var(--hall-btn-escolher-border)] bg-[var(--hall-btn-escolher-bg)] px-3 text-[10px] font-semibold text-[var(--hall-ouro)] transition-opacity hover:opacity-90"
      >
        Escolher Game do Ano
      </button>
    </article>
  )
}
