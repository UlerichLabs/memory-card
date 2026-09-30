import { useMemo, useState } from "react";
import type { CatalogoItem, CriarListaItemPayload, ListaItem } from "@/types/listas";
import { useMarcarSugeridos } from "./useMarcarSugeridos";

export function paraPayload(item: CatalogoItem): CriarListaItemPayload {
  return {
    igdb_id: item.igdb_id,
    nome: item.nome,
    igdb_capa_url: item.igdb_capa_url,
    ano_lancamento: item.ano_lancamento,
  };
}

interface UseSelecionarJogosOptions {
  itens: CatalogoItem[];
  existentes: ListaItem[];
  modo: "criar" | "adicionar";
  pagina: number;
  total: number;
  porPagina: number;
  carregar: (pagina: number, substituir: boolean) => Promise<CatalogoItem[]>;
}

export function useSelecionarJogos({
  itens,
  existentes,
  modo,
  pagina,
  total,
  porPagina,
  carregar,
}: UseSelecionarJogosOptions) {
  const [selecionados, setSelecionados] = useState<Map<number, CriarListaItemPayload>>(new Map());
  const idsExistentes = useMemo(
    () => new Set(existentes.map((item) => item.igdb_id).filter((id): id is number => id !== null)),
    [existentes],
  );
  const alternar = (item: CatalogoItem) => {
    if (modo === "adicionar" && idsExistentes.has(item.igdb_id)) return;
    setSelecionados((anteriores) => {
      const novo = new Map(anteriores);
      if (novo.has(item.igdb_id)) novo.delete(item.igdb_id);
      else novo.set(item.igdb_id, paraPayload(item));
      return novo;
    });
  };
  const desmarcar = (igdbId: number) => {
    setSelecionados((anteriores) => {
      const novo = new Map(anteriores);
      novo.delete(igdbId);
      return novo;
    });
  };
  const marcarSugeridos = useMarcarSugeridos({
    itens,
    pagina,
    total,
    porPagina,
    idsExistentes,
    carregar,
    onSelecionar: setSelecionados,
    paraPayload,
  });
  return {
    selecionados,
    idsExistentes,
    payload: [...selecionados.values()].filter((item) => !idsExistentes.has(item.igdb_id)),
    alternar,
    desmarcar,
    marcarSugeridos,
  };
}
