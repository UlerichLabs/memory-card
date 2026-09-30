import { useCallback } from "react";
import { useJogosStore } from "@/stores/jogosStore";
import { useListasStore } from "@/stores/listasStore";
import type { ListaItem } from "@/types/listas";

const MENSAGEM_ERRO_VINCULO =
  "O jogo foi registrado, mas não foi possível ligar ao desafio. Tente de novo.";

export function useZereiItem(onVinculoErro?: (mensagem: string) => void) {
  const { associarZeramento } = useListasStore();
  const { abrirModalRegistro } = useJogosStore();

  return useCallback(
    (item: ListaItem) => {
      abrirModalRegistro({
        valoresIniciais: {
          nome: item.nome,
          igdb_id: item.igdb_id ?? undefined,
          igdb_capa_url: item.igdb_capa_url ?? undefined,
          console: item.console ?? undefined,
        },
        onSalvo: async (jogo) => {
          try {
            await associarZeramento(item.id, jogo.id);
          } catch {
            onVinculoErro?.(MENSAGEM_ERRO_VINCULO);
          }
        },
      });
    },
    [abrirModalRegistro, associarZeramento, onVinculoErro],
  );
}
