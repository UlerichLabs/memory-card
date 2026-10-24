import { Link, useLocation } from "react-router-dom";
import { Trophy } from "lucide-react";
import { formatarCapaIGDB } from "@/lib/utils";
import { obterIniciaisJogo } from "./listas.utils";
import { DesafioCardCapa } from "./DesafioCardCapa";
import { ConsoleBadge } from "@/components/jogos/ConsoleBadge";
import type { ListaItem } from "@/types/listas";

export interface DesafioCardProps {
  item: ListaItem;
  onRemover: (id: number) => void;
  onZerei: (item: ListaItem) => void;
}

export function DesafioCard({ item, onRemover, onZerei }: DesafioCardProps) {
  const location = useLocation();
  const voltarPara = `${location.pathname}${location.search}`;
  const { zerado, jogo_zerado, nome, console: consoleName, ano_lancamento } = item;
  const capaUrl = formatarCapaIGDB(item.igdb_capa_url ?? undefined, "t_cover_big");
  const capaUrl2x = formatarCapaIGDB(item.igdb_capa_url ?? undefined, "t_cover_big_2x");
  const iniciais = obterIniciaisJogo(nome);
  const temLinkCatalogo = !zerado && Boolean(item.igdb_id);

  let metaTexto = "";
  if (item.origem === "regra" && jogo_zerado?.finalizado_em) {
    const anoZerado = jogo_zerado.finalizado_em.slice(0, 4);
    metaTexto = `zerado em ${anoZerado}`;
  } else if (ano_lancamento) {
    metaTexto = String(ano_lancamento);
  }

  const capaElement = (
    <DesafioCardCapa
      item={item}
      zerado={zerado}
      capaUrl={capaUrl}
      capaUrl2x={capaUrl2x}
      iniciais={iniciais}
      onRemover={onRemover}
      onZerei={onZerei}
    />
  );

  const cardContent = (
    <div
      className={`group relative flex flex-col gap-2 rounded-[10px] ${
        zerado || temLinkCatalogo ? "cursor-pointer" : "cursor-default"
      }`}
    >
      {temLinkCatalogo ? (
        <Link
          to={`/catalogo/${item.igdb_id}`}
          state={{ voltarPara }}
          aria-label={`Ver detalhes de ${nome}`}
          className="block min-h-[44px] min-w-[44px] rounded-[10px] focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]"
        >
          {capaElement}
        </Link>
      ) : (
        capaElement
      )}

      {!zerado && (
        <button
          type="button"
          aria-label="Zerei!"
          onClick={(event) => {
            event.preventDefault();
            event.stopPropagation();
            onZerei(item);
          }}
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
        className={`flex min-w-0 flex-col gap-1 ${!zerado ? "opacity-55" : ""}`}
      >
        {temLinkCatalogo ? (
          <Link
            to={`/catalogo/${item.igdb_id}`}
            state={{ voltarPara }}
            className="rounded focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]"
          >
            <span
              title={nome}
              className="line-clamp-2 text-[13px] font-semibold leading-tight text-[var(--lista-text-muted)] hover:text-[var(--lista-text-bright)]"
            >
              {nome}
            </span>
          </Link>
        ) : (
          <span
            title={nome}
            className={`line-clamp-2 text-[13px] font-semibold leading-tight ${
              zerado ? "text-[var(--lista-text-bright)]" : "text-[var(--lista-text-muted)]"
            }`}
          >
            {nome}
          </span>
        )}
        {consoleName && (
          <ConsoleBadge nome={consoleName} tamanho="sm" />
        )}
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
      <Link
        to={`/biblioteca/${jogo_zerado.id}`}
        state={{ voltarPara }}
        className="block rounded-[10px] focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]"
      >
        {cardContent}
      </Link>
    );
  }

  return cardContent;
}
