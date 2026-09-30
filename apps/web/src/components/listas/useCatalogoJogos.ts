import { useCallback, useEffect, useRef, useState } from "react";
import { listasService } from "@/lib/services/listasService";
import { resolverMensagemErro } from "./listas.constants";
import { concatenarSemDuplicados, criarFiltroCatalogo } from "./catalogo.utils";
import { useCatalogoFiltros } from "./useCatalogoFiltros";
import type { CatalogoItem, ListaOrigem } from "@/types/listas";

interface UseCatalogoJogosOptions {
  origem: ListaOrigem;
  token?: string;
}

export function useCatalogoJogos({ origem, token }: UseCatalogoJogosOptions) {
  const [itens, setItens] = useState<CatalogoItem[]>([]);
  const [pagina, setPagina] = useState(1);
  const [meta, setMeta] = useState({
    pagina: 1,
    por_pagina: 60,
    total: 0,
    total_sugeridos: null as number | null,
  });
  const [carregando, setCarregando] = useState(true);
  const [erro, setErro] = useState<string | null>(null);
  const filtros = useCatalogoFiltros();
  const controller = useRef<AbortController | null>(null);

  const carregar = useCallback(
    async (page: number, substituir: boolean) => {
      controller.current?.abort();
      const atual = new AbortController();
      controller.current = atual;
      setCarregando(true);
      setErro(null);
      try {
        const filtro = criarFiltroCatalogo(origem, {
          busca: filtros.busca,
          ordenar: filtros.ordenar,
          pagina: page,
          generoId: filtros.generoId,
          plataformaId: filtros.plataformaId,
        });
        const resposta = await listasService.buscarCatalogo(filtro, token, atual.signal);
        if (!atual.signal.aborted) {
          setItens((anteriores) =>
            substituir ? resposta.itens : concatenarSemDuplicados(anteriores, resposta.itens),
          );
          setMeta(resposta.meta);
          setPagina(page);
        }
        return resposta.itens;
      } catch (caught: unknown) {
        if (!atual.signal.aborted) {
          const codigo =
            caught instanceof Error && "codigo" in caught ? String(caught.codigo) : undefined;
          setErro(resolverMensagemErro(codigo));
        }
        return [];
      } finally {
        if (!atual.signal.aborted) setCarregando(false);
      }
    },
    [
      filtros.busca,
      filtros.generoId,
      filtros.ordenar,
      filtros.plataformaId,
      origem.igdb_id,
      origem.tipo,
      token,
    ],
  );

  useEffect(() => {
    void carregar(1, true);
    return () => controller.current?.abort();
  }, [carregar]);

  return {
    itens,
    pagina,
    meta,
    carregando,
    erro,
    ...filtros,
    carregar,
  };
}
