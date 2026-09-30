import { useState } from "react";
import type { OrdenarCatalogo } from "@/types/listas";

export function useCatalogoFiltros() {
  const [busca, setBusca] = useState("");
  const [generoId, setGeneroId] = useState<number | undefined>();
  const [plataformaId, setPlataformaId] = useState<number | undefined>();
  const [ordenar, setOrdenar] = useState<OrdenarCatalogo>("populares");
  return {
    busca,
    generoId,
    plataformaId,
    ordenar,
    setBusca,
    setGeneroId,
    setPlataformaId,
    setOrdenar,
  };
}
