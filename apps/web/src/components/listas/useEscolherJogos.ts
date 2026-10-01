import { useContext } from "react";
import { AuthContext } from "@/store/authStore";
import type { ListaItem, ListaOrigem } from "@/types/listas";
import { useCatalogoJogos } from "./useCatalogoJogos";
import { useSelecionarJogos } from "./useSelecionarJogos";

export interface EscolherJogosConfig {
  nome: string;
  descricao: string | null;
  origem: ListaOrigem;
}

export interface UseEscolherJogosOptions {
  origem: ListaOrigem;
  existentes?: ListaItem[];
  config?: EscolherJogosConfig;
  modo: "criar" | "adicionar";
}

export function useEscolherJogos({
  origem,
  existentes = [],
  modo,
}: UseEscolherJogosOptions) {
  const token = useContext(AuthContext)?.sessao?.access_token;
  const catalogo = useCatalogoJogos({ origem, token });
  const selecao = useSelecionarJogos({
    itens: catalogo.itens,
    existentes,
    modo,
    pagina: catalogo.pagina,
    total: catalogo.meta.total,
    porPagina: catalogo.meta.por_pagina,
    carregar: catalogo.carregar,
  });
  return {
    ...catalogo,
    ...selecao,
    token,
    carregarMais: () => catalogo.carregar(catalogo.pagina + 1, false),
    recarregar: () => catalogo.carregar(1, true),
  };
}
