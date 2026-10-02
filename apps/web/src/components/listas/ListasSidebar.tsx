import {
  closestCenter,
  DndContext,
  KeyboardSensor,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { arrayMove, SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { ChevronDown, ChevronUp, List, Plus, Trophy } from "lucide-react";
import { useState } from "react";
import type { ListaResumo, ListaTipo } from "@/types/listas";
import { ListaSidebarItem } from "./ListaSidebarItem";

export interface ListasSidebarProps {
  listas: ListaResumo[];
  selectedId: number | null;
  onSelect: (id: number) => void;
  onNovaLista: () => void;
  onReordenar?: (tipo: ListaTipo, listaIds: number[]) => Promise<void>;
}

export function ListasSidebar({ listas, selectedId, onSelect, onNovaLista, onReordenar = async () => undefined }: ListasSidebarProps) {
  const [desafiosExpandidos, setDesafiosExpandidos] = useState(false);
  const [filasExpandidas, setFilasExpandidas] = useState(false);
  const desafios = listas.filter((lista) => lista.tipo === "desafio");
  const filas = listas.filter((lista) => lista.tipo === "fila");
  const mostrarTodosDesafios = desafiosExpandidos || selecionadoForaDaPrevia(desafios, selectedId);
  const mostrarTodasFilas = filasExpandidas || selecionadoForaDaPrevia(filas, selectedId);
  const sensores = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 200, tolerance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const handleDragStart = (event: DragStartEvent) => {
    const tipo = event.active.data.current?.tipo as ListaTipo | undefined;
    if (tipo === "desafio") setDesafiosExpandidos(true);
    if (tipo === "fila") setFilasExpandidas(true);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const tipo = active.data.current?.tipo as ListaTipo | undefined;
    if (!tipo || over.data.current?.tipo !== tipo) return;
    const grupo = tipo === "desafio" ? desafios : filas;
    const origem = grupo.findIndex((lista) => lista.id === active.id);
    const destino = grupo.findIndex((lista) => lista.id === over.id);
    if (origem === -1 || destino === -1) return;
    const reordenadas = arrayMove(grupo, origem, destino);
    onReordenar(tipo, reordenadas.map((lista) => lista.id)).catch(() => undefined);
  };

  return (
    <DndContext sensors={sensores} collisionDetection={closestCenter} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <aside className="flex w-full flex-col gap-5">
        <div className="flex flex-col gap-1">
          <h1 className="text-[28px] font-bold tracking-[-0.01em] text-[var(--text-primary)]">Listas e Desafios</h1>
          <p className="text-[14px] text-[var(--lista-text-secondary)]">O que jogar em seguida e as metas que você quer bater.</p>
        </div>
        <button type="button" onClick={onNovaLista} className="btn-primario flex h-11 w-full items-center justify-center gap-2 px-4 text-[14px]">
          <Plus className="h-4 w-4" />
          <span>Nova lista ou desafio</span>
        </button>
        <nav aria-label="Seus desafios e listas" className="flex flex-col gap-[18px]">
          <GrupoSidebar tipo="desafio" listas={desafios} visiveis={mostrarTodosDesafios ? desafios : desafios.slice(0, 3)} selectedId={selectedId} expandido={mostrarTodosDesafios} expandir={() => setDesafiosExpandidos((valor) => !valor)} onSelect={onSelect} />
          <GrupoSidebar tipo="fila" listas={filas} visiveis={mostrarTodasFilas ? filas : filas.slice(0, 3)} selectedId={selectedId} expandido={mostrarTodasFilas} expandir={() => setFilasExpandidas((valor) => !valor)} onSelect={onSelect} />
        </nav>
      </aside>
    </DndContext>
  );
}

interface GrupoSidebarProps {
  tipo: ListaTipo;
  listas: ListaResumo[];
  visiveis: ListaResumo[];
  selectedId: number | null;
  expandido: boolean;
  expandir: () => void;
  onSelect: (id: number) => void;
}

function GrupoSidebar({ tipo, listas, visiveis, selectedId, expandido, expandir, onSelect }: GrupoSidebarProps) {
  if (listas.length === 0) return null;
  const ehDesafio = tipo === "desafio";
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-1.5 text-[12px] font-semibold uppercase tracking-[0.04em] text-[var(--lista-text-muted)]">
        {ehDesafio ? <Trophy className="h-3.5 w-3.5 text-[var(--hall-ouro)]" /> : <List className="h-3.5 w-3.5 text-[var(--lista-icon-fila)]" />}
        <span>{ehDesafio ? "Desafios" : "Filas"}</span>
      </div>
      <SortableContext items={visiveis.map((lista) => lista.id)} strategy={verticalListSortingStrategy}>
        <div className="flex flex-row gap-2 overflow-x-auto pb-1 md:flex-col md:overflow-visible">
          {visiveis.map((lista) => <ListaSidebarItem key={lista.id} lista={lista} tipo={tipo} selecionada={lista.id === selectedId} onSelect={onSelect} />)}
        </div>
      </SortableContext>
      {listas.length > 3 && (
        <button type="button" aria-expanded={expandido} onClick={expandir} className="flex items-center justify-center gap-1 rounded-lg py-1 text-[12px] font-semibold text-[var(--accent)] transition-colors hover:text-[var(--text-primary)]">
          {expandido ? <ChevronUp className="h-3.5 w-3.5" aria-hidden="true" /> : <ChevronDown className="h-3.5 w-3.5" aria-hidden="true" />}
          <span>{expandido ? "Mostrar menos" : `Mostrar mais (${listas.length - 3})`}</span>
        </button>
      )}
    </div>
  );
}

function selecionadoForaDaPrevia(listas: ListaResumo[], selectedId: number | null): boolean {
  return listas.findIndex((lista) => lista.id === selectedId) >= 3;
}
