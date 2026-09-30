export interface JogoDetalheSobreProps {
  descricao?: string | null;
}

export function JogoDetalheSobre({ descricao }: JogoDetalheSobreProps) {
  if (!descricao || !descricao.trim()) return null;

  return (
    <div className="flex flex-col gap-2.5">
      <div className="flex items-center justify-between">
        <h2 className="text-[15px] sm:text-[16px] font-semibold text-[var(--text-primary)]">
          Sobre o jogo
        </h2>
        <span className="text-[12px] text-[var(--detalhe-text-meta)]">
          Fonte: catálogo de jogos
        </span>
      </div>
      <p className="text-[14px] leading-[1.6] text-[var(--detalhe-text-sobre)] whitespace-pre-line">
        {descricao}
      </p>
    </div>
  );
}
