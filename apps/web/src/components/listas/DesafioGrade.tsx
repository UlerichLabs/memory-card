import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { useListasStore } from "@/stores/listasStore";
import { EscolherJogosDialog } from "./EscolherJogosDialog";
import { DesafioCard } from "./DesafioCard";
import type { FiltroAba, ListaDetalhada } from "@/types/listas";

export interface DesafioGradeProps {
  lista: ListaDetalhada;
}

export function DesafioGrade({ lista }: DesafioGradeProps) {
  const { filtroAba, setFiltroAba, removerItem } = useListasStore();
  const [aberto, setAberto] = useState(false);
  const [mensagem, setMensagem] = useState<string | null>(null);
  const itens = lista.itens ?? [];
  useEffect(() => {
    if (!mensagem) return;
    const timer = setTimeout(() => setMensagem(null), 4000);
    return () => clearTimeout(timer);
  }, [mensagem]);
  const filtrados = itens.filter(
    (item) => filtroAba === "todos" || (filtroAba === "zerados" ? item.zerado : !item.zerado),
  );
  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div
          role="tablist"
          aria-label="Filtrar jogos do desafio"
          className={[
          "flex items-center gap-1 rounded-[10px]",
          "border border-[var(--lista-tabs-border)] bg-[var(--lista-tabs-bg)] p-1"
        ].join(" ")}
        >
          {(["todos", "zerados", "pendentes"] as FiltroAba[]).map((tab) => {
            const ativo = filtroAba === tab;
            const count =
              tab === "todos"
                ? itens.length
                : tab === "zerados"
                  ? itens.filter((item) => item.zerado).length
                  : itens.filter((item) => !item.zerado).length;
            return (
              <button
                key={tab}
                type="button"
                role="tab"
                aria-selected={ativo}
                onClick={() => setFiltroAba(tab)}
                className={`flex h-8 items-center rounded-[7px] px-3 text-[13px] font-semibold
                  ${ativo ? "bg-[var(--lista-tab-active-bg)]" : "text-[var(--lista-text-muted)]"}
                  ${ativo ? "text-[var(--lista-tab-active-text)]" : ""}`}
              >
                {tab[0].toUpperCase() + tab.slice(1)} · {count}
              </button>
            );
          })}
        </div>
        <div className="flex items-center gap-3">
          {mensagem && (
            <span className="text-[12px] text-[var(--lista-text-light)]">{mensagem}</span>
          )}
          {lista.origem && (
            <button
              type="button"
              onClick={() => setAberto(true)}
              className={[
          "flex h-9 items-center gap-2",
          "rounded-lg border border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)]",
          "px-3 text-[13px] font-semibold text-[var(--hall-ouro)]"
        ].join(" ")}
            >
              <Plus className="h-3.5 w-3.5" />
              Adicionar jogos
            </button>
          )}
        </div>
      </div>
      {filtrados.length === 0 ? (
        <div className="py-12 text-center text-[14px] text-[var(--lista-text-muted)]">
          Nenhum jogo aqui ainda.
        </div>
      ) : (
        <div className="grid grid-cols-3 gap-x-4 gap-y-5 md:grid-cols-4 xl:grid-cols-6">
          {filtrados.map((item) => (
            <DesafioCard
              key={item.id}
              item={item}
              onRemover={removerItem}
            />
          ))}
        </div>
      )}
      {lista.origem && (
        <EscolherJogosDialog
          open={aberto}
          modo="adicionar"
          lista={lista}
          onClose={() => setAberto(false)}
          onAdded={(quantidade) => setMensagem(`${quantidade} jogos adicionados`)}
        />
      )}
    </div>
  );
}
