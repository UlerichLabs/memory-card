import { Trophy, List, Plus } from "lucide-react";
import { formatarSubFila } from "./listas.utils";
import type { ListaResumo } from "@/types/listas";

export interface ListasSidebarProps {
  listas: ListaResumo[];
  selectedId: number | null;
  onSelect: (id: number) => void;
  onNovaLista: () => void;
}

export function ListasSidebar({ listas, selectedId, onSelect, onNovaLista }: ListasSidebarProps) {
  const desafios = listas.filter((l) => l.tipo === "desafio");
  const filas = listas.filter((l) => l.tipo === "fila");

  return (
    <aside className="flex flex-col gap-5 w-full">
      <div className="flex flex-col gap-1">
        <h1 className="text-[28px] font-bold tracking-[-0.01em] text-[var(--text-primary)]">
          Listas e Desafios
        </h1>
        <p className="text-[14px] text-[var(--lista-text-secondary)]">
          O que jogar em seguida e as metas que você quer bater.
        </p>
      </div>

      <button
        type="button"
        onClick={onNovaLista}
        className="btn-primario flex h-11 w-full items-center justify-center gap-2 px-4 text-[14px]"
      >
        <Plus className="h-4 w-4" />
        <span>Nova lista ou desafio</span>
      </button>

      <nav
        aria-label="Seus desafios e listas"
        className="flex flex-col gap-[18px]"
      >
        {desafios.length > 0 && (
          <div className="flex flex-col gap-2">
            <div className={["flex items-center gap-1.5 text-[12px] font-semibold uppercase",
  "tracking-[0.04em] text-[var(--lista-text-muted)]"].join(" ")}>
              <Trophy className="h-3.5 w-3.5 text-[var(--hall-ouro)]" />
              <span>Desafios</span>
            </div>

            <div className="flex flex-row overflow-x-auto pb-1 gap-2 md:flex-col md:overflow-visible">
              {desafios.map((desafio) => {
                const ativo = desafio.id === selectedId;
                const progresso = desafio.progresso;
                const concluido = progresso?.concluido ?? false;
                const quantidade = desafio.total_itens === 1 ? "jogo" : "jogos";
                const nomeOrigem = desafio.origem
                  ? ` · ${desafio.origem.tipo[0].toUpperCase()}${desafio.origem.tipo.slice(1)}`
                  : "";
                const subTexto = `${desafio.total_itens} ${quantidade}${nomeOrigem}`;
                const pct = progresso?.percentual ?? 0;

                return (
                  <button
                    key={desafio.id}
                    type="button"
                    aria-current={ativo ? "page" : undefined}
                    onClick={() => onSelect(desafio.id)}
                    className={`flex flex-col gap-1 rounded-xl p-[12px_14px] text-left transition-colors min-w-[240px]
md:min-w-0 md:w-full shrink-0 ${
                      ativo
                        ? "border border-[var(--lista-item-selected-border)] bg-[var(--lista-item-selected-bg)]"
                        :
  "border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] hover:border-[var(--lista-text-dim)]"
                    }`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="truncate text-[14px] font-semibold text-[var(--text-primary)]">
                        {desafio.nome}
                      </span>
                      {concluido && (
                        <Trophy
                          aria-label="Concluído"
                          className="h-4 w-4 fill-current text-[var(--hall-ouro)] shrink-0"
                        />
                      )}
                    </div>

                    <span className="text-[12px] text-[var(--lista-text-muted)]">{subTexto}</span>

                    {progresso && (
                      <div className="mt-1 flex flex-col gap-1.5">
                        <div className="h-1 w-full overflow-hidden rounded-[2px] bg-[var(--lista-progress-track)]">
                          <div
                            style={{
                              width: `${Math.min(100, Math.max(0, pct))}%`,
                            }}
                            className={`h-full ${
                              concluido
                                ? "bg-[var(--hall-ouro)]"
                                : "bg-[var(--lista-progress-fill)]"
                            }`}
                          />
                        </div>
                        <span className="text-[12px] tabular-nums text-[var(--lista-text-secondary)]">
                          {progresso.feitos} / {progresso.meta}
                        </span>
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        )}

        {filas.length > 0 && (
          <div className="flex flex-col gap-2">
            <div className={["flex items-center gap-1.5 text-[12px] font-semibold uppercase",
  "tracking-[0.04em] text-[var(--lista-text-muted)]"].join(" ")}>
              <List className="h-3.5 w-3.5 text-[var(--lista-icon-fila)]" />
              <span>Filas</span>
            </div>

            <div className="flex flex-row overflow-x-auto pb-1 gap-2 md:flex-col md:overflow-visible">
              {filas.map((fila) => {
                const ativo = fila.id === selectedId;
                const subTexto = formatarSubFila(fila.itens_pendentes);

                return (
                  <button
                    key={fila.id}
                    type="button"
                    aria-current={ativo ? "page" : undefined}
                    onClick={() => onSelect(fila.id)}
                    className={`flex flex-col gap-1 rounded-xl p-[12px_14px] text-left transition-colors min-w-[240px]
md:min-w-0 md:w-full shrink-0 ${
                      ativo
                        ? "border border-[var(--lista-item-selected-border)] bg-[var(--lista-item-selected-bg)]"
                        :
  "border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] hover:border-[var(--lista-text-dim)]"
                    }`}
                  >
                    <span className="truncate text-[14px] font-semibold text-[var(--text-primary)]">
                      {fila.nome}
                    </span>
                    <span className="text-[12px] text-[var(--lista-text-muted)]">{subTexto}</span>
                  </button>
                );
              })}
            </div>
          </div>
        )}
      </nav>
    </aside>
  );
}
