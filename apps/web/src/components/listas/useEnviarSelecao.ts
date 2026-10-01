import { useRef, useState } from "react";
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
    }, recarregar?: boolean) => Promise<{ id: number }>;
    carregarListas: () => Promise<unknown>;
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
  const enviandoRef = useRef(false);
  const enviar = async () => {
    if (enviandoRef.current) return;
    enviandoRef.current = true;
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
        }, false);
        onClose();
        await store.carregarListas();
        navigate(`/listas/${criada.id}`);
      } else if (modo === "adicionar") {
        const resposta = await store.adicionarItensLote(itens);
        onAdded?.(resposta.adicionados);
        onClose();
      }
    } catch (caught: unknown) {
      enviandoRef.current = false;
      setEnviando(false);
      setErro(
        caught instanceof ApiError
          ? resolverMensagemErro(caught.codigo)
          : "Não foi possível salvar. Tente novamente.",
      );
    }
  };
  return { enviando, erro, enviar };
}
