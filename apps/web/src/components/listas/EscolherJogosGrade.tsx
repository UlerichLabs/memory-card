import { EscolherJogosCard } from "./EscolherJogosCard";
import { formatarCapaIGDB } from "@/lib/utils";
import type { CatalogoItem, CriarListaItemPayload } from "@/types/listas";

interface Props {
  itens: CatalogoItem[];
  selecionados: Map<number, CriarListaItemPayload>;
  idsExistentes: Set<number>;
  carregando: boolean;
  erro: string | null;
  revisando: boolean;
  total: number;
  porPagina: number;
  onToggle: (item: CatalogoItem) => void;
  onDesmarcar: (id: number) => void;
  onRecarregar: () => void;
  onCarregarMais: () => void;
}

export function EscolherJogosGrade({
  itens,
  selecionados,
  idsExistentes,
  carregando,
  erro,
  revisando,
  total,
  porPagina,
  onToggle,
  onDesmarcar,
  onRecarregar,
  onCarregarMais,
}: Props) {
  if (carregando && itens.length === 0) {
    return (
      <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto p-[0_28px_20px]">
        <div className="grid grid-cols-3 gap-x-[14px] gap-y-[18px] md:grid-cols-4 xl:grid-cols-6">
          {Array.from({ length: 12 }, (_, index) => (
            <div
              key={index}
              className="aspect-[3/4] animate-pulse rounded-[10px] bg-[var(--lista-cover-bg)]"
            />
          ))}
        </div>
      </div>
    );
  }
  if (revisando) {
    return (
      <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto p-[0_28px_20px]">
        <div className="flex flex-col gap-2">
          {[...selecionados.values()].map((item) => (
            <div
              key={item.igdb_id}
              className="flex items-center gap-3 rounded-lg border border-[var(--lista-card-border)] p-2"
            >
              <span className="flex h-11 w-8 shrink-0 items-center justify-center overflow-hidden rounded bg-[var(--lista-cover-bg)]">
                {item.igdb_capa_url ? (
                  <img
                    src={formatarCapaIGDB(item.igdb_capa_url ?? undefined)}
                    alt=""
                    className="h-full w-full object-cover"
                  />
                ) : (
                  item.nome.slice(0, 2).toUpperCase()
                )}
              </span>
              <span className="min-w-0 flex-1 truncate text-[13px]">
                {item.nome} {item.ano_lancamento ? `· ${item.ano_lancamento}` : ""}
              </span>
              <button
                type="button"
                aria-label={`Desmarcar ${item.nome}`}
                onClick={() => onDesmarcar(item.igdb_id)}
                className="h-9 w-9 rounded-full text-[var(--lista-text-muted)]"
              >
                ×
              </button>
            </div>
          ))}
        </div>
      </div>
    );
  }
  return (
    <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto overflow-x-hidden p-[0_28px_20px]">
      {erro && (
        <div className="py-8 text-center text-[13px] text-[var(--danger)]">
          {erro}{" "}
          <button
            type="button"
            onClick={onRecarregar}
            className="ml-2 underline"
          >
            Tentar de novo
          </button>
        </div>
      )}
      {itens.length === 0 && !carregando ? (
        <div className="py-16 text-center text-[14px] text-[var(--lista-text-muted)]">
          Nenhum jogo encontrado com esses filtros.
        </div>
      ) : (
        <div className="grid grid-cols-3 gap-x-[14px] gap-y-[18px] md:grid-cols-4 xl:grid-cols-6">
          {itens.map((item) => (
            <EscolherJogosCard
              key={item.igdb_id}
              item={item}
              marcado={selecionados.has(item.igdb_id) || idsExistentes.has(item.igdb_id)}
              jaNaLista={idsExistentes.has(item.igdb_id)}
              onToggle={() => onToggle(item)}
            />
          ))}
        </div>
      )}
      {itens.length < total && (
        <button
          type="button"
          onClick={onCarregarMais}
          disabled={carregando}
          className="mx-auto mt-6 flex h-10 items-center rounded-lg border border-[var(--lista-btn-icon-border)] px-4 text-[13px] text-[var(--lista-text-light)]"
        >
          {carregando ? "Carregando..." : `Carregar mais ${porPagina}`}
        </button>
      )}
    </div>
  );
}
