import { useState } from "react";
import { ApiError } from "@/lib/api";
import { resolverMensagemErro } from "./listas.constants";
import type { CriarListaItemPayload } from "@/types/listas";
import type { EscolherJogosConfig } from "./useEscolherJogos";

interface Options {
  modo: "criar" | "adicionar";
  config?: EscolherJogosConfig;
  itens: CriarListaItemPayload[];
  store: {
    criarLista: (payload: {
      tipo: "desafio";
      nome: string;
      descricao: string | null;
      origem: EscolherJogosConfig["origem"];
      itens: CriarListaItemPayload[];
    }) => Promise<{ id: number }>;
    adicionarItensLote: (itens: CriarListaItemPayload[]) => Promise<{ adicionados: number }>;
  };
  navigate: (path: string) => void;
  onClose: () => void;
  onAdded?: (quantidade: number) => void;
}

export function useEnviarSelecao({
  modo,
  config,
  itens,
  store,
  navigate,
  onClose,
  onAdded,
}: Options) {
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const enviar = async () => {
    setEnviando(true);
    setErro(null);
    try {
      if (modo === "criar" && config) {
        const criada = await store.criarLista({
          tipo: "desafio",
          nome: config.nome,
          descricao: config.descricao,
          origem: config.origem,
          itens,
        });
        onClose();
        navigate(`/listas/${criada.id}`);
      } else if (modo === "adicionar") {
        const resposta = await store.adicionarItensLote(itens);
        onAdded?.(resposta.adicionados);
        onClose();
      }
    } catch (caught: unknown) {
      setErro(
        caught instanceof ApiError
          ? resolverMensagemErro(caught.codigo)
          : "Não foi possível salvar. Tente novamente.",
      );
    } finally {
      setEnviando(false);
    }
  };
  return { enviando, erro, enviar };
}
