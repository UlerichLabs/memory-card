import type { CatalogoItem, FiltroCatalogo, ListaOrigem, OrdenarCatalogo } from "@/types/listas";

export function criarFiltroCatalogo(
  origem: ListaOrigem,
  filtros: {
    busca: string;
    ordenar: OrdenarCatalogo;
    pagina: number;
    generoId?: number;
    plataformaId?: number;
  },
): FiltroCatalogo {
  return {
    origem: origem.tipo,
    id: origem.igdb_id ?? 0,
    busca: filtros.busca,
    ordenar: filtros.ordenar,
    pagina: filtros.pagina,
    por_pagina: 60,
    ...(filtros.generoId ? { genero_id: filtros.generoId } : {}),
    ...(filtros.plataformaId ? { plataforma_id: filtros.plataformaId } : {}),
  };
}

export function concatenarSemDuplicados(
  anteriores: CatalogoItem[],
  novos: CatalogoItem[],
) {
  return [
    ...anteriores,
    ...novos.filter(
      (item) => !anteriores.some((anterior) => anterior.igdb_id === item.igdb_id),
    ),
  ];
}
