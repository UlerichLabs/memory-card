import { Crown } from 'lucide-react'

export interface GameDoAnoVazioProps {
  ano: number
  totalJogos: number
  onEscolher: (ano: number) => void
}

export function GameDoAnoVazio({ ano, totalJogos, onEscolher }: GameDoAnoVazioProps) {
  const textoJogos = totalJogos === 1 ? `1 jogo zerado em ${ano}` : `${totalJogos} jogos zerados em ${ano}`

  return (
    <article className="flex flex-col justify-between gap-[14px] rounded-[16px] border border-dashed border-[var(--hall-card-vazio-border)] bg-transparent p-[18px]">
      <div>
        <span className="text-[32px] font-extrabold tracking-[-0.02em] tabular-nums text-[var(--hall-ano-vazio)]">
          {ano}
        </span>
      </div>

      <div className="flex min-h-[144px] flex-col items-center justify-center gap-[10px] py-2 text-center">
        <Crown className="h-8 w-8 text-[var(--hall-coroa-vazio)]" aria-hidden="true" />
        <span className="text-[14px] font-semibold text-[var(--hall-vazio-titulo)]">
          Sem Game do Ano
        </span>
        <span className="text-[12px] text-[var(--hall-muted)]">
          {textoJogos}
        </span>
      </div>

      <button
        type="button"
        onClick={() => onEscolher(ano)}
        className="flex h-[40px] w-full items-center justify-center rounded-[10px] border border-[var(--hall-btn-escolher-border)] bg-[var(--hall-btn-escolher-bg)] px-4 text-[14px] font-semibold text-[var(--hall-ouro)] transition-opacity hover:opacity-90"
      >
        Escolher Game do Ano
      </button>
    </article>
  )
}
