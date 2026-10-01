import { X } from "lucide-react";
import { DialogTitle } from "@/components/ui/dialog";
import { EscolherJogosFiltros } from "./EscolherJogosFiltros";
import { ROTULOS_ORIGEM } from "./listas.constants";
import type { ListaOrigem, OrigemTipo } from "@/types/listas";

interface Props {
  origem: ListaOrigem;
  nome?: string;
  busca: string;
  generoId?: number;
  plataformaId?: number;
  ordenar: "populares" | "lancamento" | "nome";
  token?: string;
  onBusca: (valor: string) => void;
  onGenero: (valor?: number) => void;
  onPlataforma: (valor?: number) => void;
  onOrdenar: (valor: "populares" | "lancamento" | "nome") => void;
  onClose: () => void;
}

export function EscolherJogosCabecalho({
  origem,
  nome,
  busca,
  generoId,
  plataformaId,
  ordenar,
  token,
  onBusca,
  onGenero,
  onPlataforma,
  onOrdenar,
  onClose,
}: Props) {
  const titulo = `${ROTULOS_ORIGEM[origem.tipo as OrigemTipo]} · ${origem.nome}`;
  return (
    <div className="shrink-0 border-b border-[var(--lista-card-border)] p-[24px_28px_16px]">
      <div className="flex items-start justify-between">
        <div>
          <DialogTitle className="text-[20px] font-bold">Escolher jogos</DialogTitle>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <span className="rounded-full border border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)] px-[9px] py-[2px] text-[12px] font-semibold text-[var(--hall-ouro)]">
              {titulo}
            </span>
            <span className="text-[13px] text-[var(--lista-text-secondary)]">
              {nome} · marque os jogos que entram no desafio
            </span>
          </div>
        </div>
        <button
          type="button"
          aria-label="Fechar"
          onClick={onClose}
          className="flex h-10 w-10 items-center justify-center rounded-lg text-[var(--lista-text-muted)] hover:bg-[var(--bg-surface-alt)]"
        >
          <X className="h-[18px] w-[18px]" />
        </button>
      </div>
      <div className="mt-4">
        <EscolherJogosFiltros
          origem={origem.tipo}
          busca={busca}
          generoId={generoId}
          plataformaId={plataformaId}
          ordenar={ordenar}
          token={token}
          onBusca={onBusca}
          onGenero={onGenero}
          onPlataforma={onPlataforma}
          onOrdenar={onOrdenar}
        />
      </div>
    </div>
  );
}
