import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Trophy } from "lucide-react";
import { formatarSubFila } from "./listas.utils";
import type { ListaResumo, ListaTipo } from "@/types/listas";

interface ListaSidebarItemProps {
  lista: ListaResumo;
  tipo: ListaTipo;
  selecionada: boolean;
  onSelect: (id: number) => void;
}

export function ListaSidebarItem({ lista, tipo, selecionada, onSelect }: ListaSidebarItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: lista.id,
    data: { tipo },
  });
  const estilo = { transform: CSS.Transform.toString(transform), transition };
  const progresso = lista.progresso;
  const concluido = progresso?.concluido ?? false;
  const quantidade = lista.total_itens === 1 ? "jogo" : "jogos";
  const nomeOrigem = lista.origem
    ? ` · ${lista.origem.tipo[0].toUpperCase()}${lista.origem.tipo.slice(1)}`
    : "";
  const subTexto = tipo === "desafio"
    ? `${lista.total_itens} ${quantidade}${nomeOrigem}`
    : formatarSubFila(lista.itens_pendentes);
  const pct = progresso?.percentual ?? 0;

  return (
    <button
      ref={setNodeRef}
      style={estilo}
      {...attributes}
      {...listeners}
      type="button"
      aria-current={selecionada ? "page" : undefined}
      aria-label={`${lista.nome}. Arraste para mudar a ordem`}
      onClick={() => onSelect(lista.id)}
      className={`flex shrink-0 min-w-[240px] flex-col gap-1 rounded-xl p-[12px_14px] text-left transition-colors md:min-w-0 md:w-full ${isDragging ? "z-10 opacity-70" : ""} ${selecionada ? "border border-[var(--lista-item-selected-border)] bg-[var(--lista-item-selected-bg)]" : "border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] hover:border-[var(--lista-text-dim)]"}`}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="truncate text-[14px] font-semibold text-[var(--text-primary)]">{lista.nome}</span>
        {tipo === "desafio" && concluido && <Trophy aria-label="Concluído" className="h-4 w-4 shrink-0 fill-current text-[var(--hall-ouro)]" />}
      </div>
      <span className="text-[12px] text-[var(--lista-text-muted)]">{subTexto}</span>
      {tipo === "desafio" && progresso && (
        <div className="mt-1 flex flex-col gap-1.5">
          <div className="h-1 w-full overflow-hidden rounded-[2px] bg-[var(--lista-progress-track)]">
            <div style={{ width: `${Math.min(100, Math.max(0, pct))}%` }} className={`h-full ${concluido ? "bg-[var(--hall-ouro)]" : "bg-[var(--lista-progress-fill)]"}`} />
          </div>
          <span className="text-[12px] tabular-nums text-[var(--lista-text-secondary)]">{progresso.feitos} / {progresso.meta}</span>
        </div>
      )}
    </button>
  );
}
