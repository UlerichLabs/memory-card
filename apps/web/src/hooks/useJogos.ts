import { useQuery } from "@tanstack/react-query";
import type { FiltrosJogoDto, JogoZerado } from "@memory-card/types";

import { listarJogos } from "@/lib/api";

export type UseJogosParams = Pick<FiltrosJogoDto, "page" | "limit"> &
  Partial<Pick<FiltrosJogoDto, "q" | "console" | "genero" | "tipo" | "notaMin" | "notaMax" | "ano">>;

const emptyJogos: JogoZerado[] = [];

export function useJogos(params: UseJogosParams) {
  const query = useQuery({
    queryKey: ["jogos", params],
    queryFn: () => listarJogos(params)
  });

  return {
    jogos: query.data?.data ?? emptyJogos,
    total: query.data?.total ?? 0,
    totalPaginas: query.data?.totalPages ?? 1,
    isLoading: query.isLoading,
    isError: query.isError
  };
}
