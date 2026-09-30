import { useCallback } from "react";
import type { CatalogoItem, CriarListaItemPayload } from "@/types/listas";

interface UseMarcarSugeridosOptions {
  itens: CatalogoItem[];
  pagina: number;
  total: number;
  porPagina: number;
  idsExistentes: Set<number>;
  carregar: (pagina: number, substituir: boolean) => Promise<CatalogoItem[]>;
  onSelecionar: (
    atualizar: (
      selecionados: Map<number, CriarListaItemPayload>,
    ) => Map<number, CriarListaItemPayload>,
  ) => void;
  paraPayload: (item: CatalogoItem) => CriarListaItemPayload;
}

export function useMarcarSugeridos({
  itens,
  pagina,
  total,
  porPagina,
  idsExistentes,
  carregar,
  onSelecionar,
  paraPayload,
}: UseMarcarSugeridosOptions) {
  return useCallback(async () => {
    const todos = [...itens];
    const paginas = Math.ceil(total / porPagina);
    for (let proxima = pagina + 1; proxima <= paginas; proxima += 1) {
      todos.push(...(await carregar(proxima, false)));
    }
    onSelecionar((anteriores) => {
      const novo = new Map(anteriores);
      todos
        .filter((item) => item.sugerido && !idsExistentes.has(item.igdb_id))
        .forEach((item) => novo.set(item.igdb_id, paraPayload(item)));
      return novo;
    });
  }, [carregar, idsExistentes, itens, onSelecionar, pagina, paraPayload, porPagina, total]);
}
