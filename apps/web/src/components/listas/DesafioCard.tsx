import { Link } from "react-router-dom";
import { Trophy } from "lucide-react";
import { formatarCapaIGDB } from "@/lib/utils";
import { obterIniciaisJogo } from "./listas.utils";
import { DesafioCardCapa } from "./DesafioCardCapa";
import type { ListaItem } from "@/types/listas";

export interface DesafioCardProps {
  item: ListaItem;
  onRemover: (id: number) => void;
  onZerei: (item: ListaItem) => void;
}

export function DesafioCard({ item, onRemover, onZerei }: DesafioCardProps) {
  const { zerado, jogo_zerado, nome, console: consoleName, ano_lancamento } = item;
  const capaUrl = formatarCapaIGDB(item.igdb_capa_url ?? undefined, "t_cover_big");
  const capaUrl2x = formatarCapaIGDB(item.igdb_capa_url ?? undefined, "t_cover_big_2x");
  const iniciais = obterIniciaisJogo(nome);

  let metaTexto = "";
  if (item.origem === "regra" && jogo_zerado?.finalizado_em) {
    const anoZerado = jogo_zerado.finalizado_em.slice(0, 4);
    metaTexto = [consoleName, `zerado em ${anoZerado}`].filter(Boolean).join(" · ");
  } else {
    metaTexto = [ano_lancamento, consoleName].filter(Boolean).join(" · ");
  }

  const cardContent = (
    <div
      className={`group relative flex flex-col gap-2 rounded-[10px] ${
        zerado ? "cursor-pointer" : "cursor-default"
      }`}
    >
      <DesafioCardCapa
        item={item}
        zerado={zerado}
        capaUrl={capaUrl}
        capaUrl2x={capaUrl2x}
        iniciais={iniciais}
        onRemover={onRemover}
        onZerei={onZerei}
      />

      {!zerado && (
        <button
          type="button"
          aria-label="Zerei!"
          onClick={() => onZerei(item)}
          className="hidden h-8 w-full items-center justify-center gap-1.5 rounded-lg
            border border-[var(--lista-btn-zerei-border)] bg-[var(--lista-btn-zerei-bg)]
            text-[13px] font-bold text-[var(--lista-btn-zerei-text)] opacity-0
            transition-opacity hover:opacity-90 focus:opacity-100 group-hover:opacity-100 sm:flex"
        >
          <Trophy className="h-3.5 w-3.5" />
          Zerei!
        </button>
      )}

      <div
        className={`flex flex-col gap-0.5 ${!zerado ? "opacity-55" : ""}`}
      >
        <span
          title={nome}
          className={`line-clamp-2 text-[13px] font-semibold leading-tight ${
            zerado ? "text-[var(--lista-text-bright)]" : "text-[var(--lista-text-muted)]"
          }`}
        >
          {nome}
        </span>
        {metaTexto && (
          <span className="truncate text-[11px] text-[var(--lista-text-dim)]">
            {metaTexto}
          </span>
        )}
      </div>
    </div>
  );

  if (zerado && jogo_zerado) {
    return (
      <Link to={`/biblioteca/${jogo_zerado.id}`} className="block focus:outline-none">
        {cardContent}
      </Link>
    );
  }

  return cardContent;
}
