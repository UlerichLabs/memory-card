import { ChevronLeft, Loader2 } from "lucide-react";

interface Props {
  modo: "criar" | "adicionar";
  selecionados: number;
  meta: number;
  enviando: boolean;
  erro?: string | null;
  onRevisar: () => void;
  onVoltar: () => void;
  onEnviar: () => void;
}

export function EscolherJogosRodape({
  modo,
  selecionados,
  meta,
  enviando,
  erro,
  onRevisar,
  onVoltar,
  onEnviar,
}: Props) {
  return (
    <div className="flex shrink-0 flex-col gap-3 border-t border-[var(--modal-border)] p-[16px_28px_20px] sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-3">
        <span className="flex h-9 min-w-9 items-center justify-center rounded-lg bg-[var(--lista-pill-desafio-bg)] px-2 text-[15px] font-extrabold text-[var(--hall-ouro)]">
          {selecionados}
        </span>
        <div className="flex flex-col">
          <span className="text-[13px] font-semibold">
            jogos selecionados ·{" "}
            {modo === "criar" ? `meta do desafio: zerar ${meta}` : `meta passa para ${meta}`}
          </span>
          <button
            type="button"
            onClick={onRevisar}
            className="text-left text-[12px] text-[var(--lista-icon-fila)]"
          >
            Ver e revisar seleção
          </button>
        </div>
      </div>
      <div className="flex items-center justify-end gap-3">
        <button
          type="button"
          onClick={onVoltar}
          className="h-11 rounded-[10px] border border-[var(--lista-btn-icon-border)] px-[18px]"
        >
          {modo === "criar" ? (
            <>
              <ChevronLeft className="mr-1 inline h-4 w-4" />
              Voltar
            </>
          ) : (
            "Cancelar"
          )}
        </button>
        <button
          type="button"
          disabled={enviando || selecionados === 0 || selecionados > 1000}
          onClick={onEnviar}
          className="flex h-11 items-center gap-2 rounded-[10px] bg-[var(--hall-ouro)] px-5 text-[14px] font-bold text-[var(--ouro-jogo-ano-text)] disabled:opacity-50"
        >
          {enviando && <Loader2 className="h-4 w-4 animate-spin" />}
          {modo === "criar" ? "Criar desafio" : `Adicionar ${selecionados} jogos`}
        </button>
        {erro && <span className="text-xs text-[var(--danger)]">{erro}</span>}
      </div>
    </div>
  );
}
