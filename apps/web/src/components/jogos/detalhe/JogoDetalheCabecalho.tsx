import { ArrowLeft, Crown, Pencil, Trash2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { ConsoleBadge } from '@/components/jogos/ConsoleBadge'

export interface JogoDetalheCabecalhoProps {
  jogo: JogoZeradoDTO
  onVoltar: () => void
  onEditar: () => void
  onExcluir: () => void
}

export function JogoDetalheCabecalho({
  jogo,
  onVoltar,
  onEditar,
  onExcluir,
}: JogoDetalheCabecalhoProps) {
  const anoFinalizado = jogo.finalizado_em ? jogo.finalizado_em.slice(0, 4) : ''
  const generos = jogo.genero
    ? jogo.genero.split(',').map((g) => g.trim()).filter(Boolean)
    : []

  return (
    <div className="flex flex-col gap-6">
      <div className="hidden items-center justify-between sm:flex">
        <button
          type="button"
          onClick={onVoltar}
          className="flex h-[40px] items-center gap-1.5 rounded-[10px] px-3 pl-2 text-[14px] font-medium text-[var(--detalhe-link-voltar)] transition-colors hover:text-[var(--text-primary)]"
        >
          <ArrowLeft className="h-[18px] w-[18px]" aria-hidden="true" />
          <span>Voltar para a Biblioteca</span>
        </button>

        <div className="flex items-center gap-2.5">
          <button
            type="button"
            onClick={onEditar}
            className="flex h-[40px] items-center gap-2 rounded-[10px] border border-[var(--detalhe-btn-editar-border)] bg-[var(--detalhe-btn-editar-bg)] px-4 text-[14px] font-medium text-[var(--detalhe-btn-editar-text)] transition-colors hover:border-[var(--biblioteca-control-border-hover)] hover:text-[var(--text-primary)]"
          >
            <Pencil className="h-4 w-4" aria-hidden="true" />
            <span>Editar</span>
          </button>
          <button
            type="button"
            onClick={onExcluir}
            className="flex h-[40px] items-center gap-2 rounded-[10px] border border-[var(--detalhe-btn-excluir-border)] bg-transparent px-4 text-[14px] font-medium text-[var(--detalhe-btn-excluir-text)] transition-colors hover:bg-[var(--danger)]/10"
          >
            <Trash2 className="h-4 w-4" aria-hidden="true" />
            <span>Excluir</span>
          </button>
        </div>
      </div>

      <div className="flex flex-col gap-3.5">
        {jogo.destaque && (
          <div className="flex">
            <span className="inline-flex items-center gap-1.5 rounded-full bg-[var(--ouro-jogo-ano)] px-3 py-1.5 text-[12px] font-bold text-[var(--ouro-jogo-ano-text)] sm:text-[13px]">
              <Crown className="h-3.5 w-3.5" aria-hidden="true" />
              <span>Jogo do ano {anoFinalizado}</span>
            </span>
          </div>
        )}

        <h1 className="text-[26px] font-bold tracking-[-0.015em] text-[var(--text-primary)] leading-[1.15] sm:text-[36px]">
          {jogo.nome}
        </h1>

        <div className="flex flex-wrap items-center gap-2">
          {jogo.console && (
            <ConsoleBadge nome={jogo.console} tamanho="md" />
          )}

          {generos.map((gen) => (
            <span
              key={gen}
              className="inline-flex h-[28px] sm:h-[30px] items-center rounded-full border border-[var(--detalhe-chip-genero-border)] bg-[var(--detalhe-chip-genero-bg)] px-3 text-[12px] sm:text-[13px] text-[var(--detalhe-chip-genero-text)]"
            >
              {gen}
            </span>
          ))}

          {jogo.tipo && (
            <span className="inline-flex h-[28px] sm:h-[30px] items-center rounded-full border border-dashed border-[var(--detalhe-chip-tipo-border)] bg-transparent px-3 text-[12px] sm:text-[13px] text-[var(--detalhe-chip-tipo-text)]">
              {jogo.tipo}
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
