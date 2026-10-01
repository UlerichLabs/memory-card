import type { useNavigate } from "react-router-dom";
import type { ListaDetalhada, ListaOrigem } from "@/types/listas";
import type { EscolherJogosConfig } from "./useEscolherJogos";

export interface EscolherJogosDialogProps {
  open: boolean;
  modo: "criar" | "adicionar";
  config?: EscolherJogosConfig;
  lista?: ListaDetalhada;
  onClose: () => void;
  onBack?: () => void;
  onAdded?: (quantidade: number) => void;
}

export interface ConteudoEscolhaProps extends EscolherJogosDialogProps {
  origem: ListaOrigem;
  navigate: ReturnType<typeof useNavigate>;
  store: ReturnType<(typeof import("@/stores/listasStore"))["useListasStore"]>;
}
