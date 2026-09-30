import { Check } from "lucide-react";
import { formatarCapaIGDB } from "@/lib/utils";
import { obterIniciaisJogo } from "./listas.utils";
import type { CatalogoItem } from "@/types/listas";

interface EscolherJogosCardProps {
  item: CatalogoItem;
  marcado: boolean;
  jaNaLista: boolean;
  onToggle: () => void;
}

export function EscolherJogosCard({ item, marcado, jaNaLista, onToggle }: EscolherJogosCardProps) {
  const capa = formatarCapaIGDB(item.igdb_capa_url ?? undefined, "t_cover_big");
  const capa2x = formatarCapaIGDB(item.igdb_capa_url ?? undefined, "t_cover_big_2x");
  return (
    <label
      className={`group flex min-w-0 cursor-pointer flex-col gap-2 ${jaNaLista ? "cursor-not-allowed opacity-50" : ""}`}
    >
      <div
        className={`relative aspect-[3/4] w-full overflow-hidden rounded-[10px] bg-[var(--lista-cover-bg)]
${marcado ? "border-2 border-[var(--hall-ouro)]" : "border border-[var(--lista-card-zerado-border)]"}`}
      >
        <input
          type="checkbox"
          checked={marcado}
          disabled={jaNaLista}
          onChange={onToggle}
          className="absolute left-2 top-2 z-10 h-5 w-5 accent-[var(--hall-ouro)]"
        />
        {capa ? (
          <img
            src={capa}
            srcSet={capa2x ? `${capa} 1x, ${capa2x} 2x` : undefined}
            alt={item.nome}
            className="h-full w-full object-cover"
            loading="lazy"
          />
        ) : (
          <span
            className="flex h-full items-center justify-center text-[22px] font-bold
              text-[var(--lista-text-secondary)]"
          >
            {obterIniciaisJogo(item.nome)}
          </span>
        )}
        {item.ja_zerado && (
          <span
            className={[
              "absolute bottom-1.5 left-1.5 flex",
              "items-center gap-1 rounded-full border",
              "border-[var(--lista-zerado-pill-border)] bg-[var(--lista-zerado-pill-bg)] px-[7px] py-[2px]",
              "text-[10px] font-bold text-[var(--lista-zerado-pill-text)]",
            ].join(" ")}
          >
            <Check className="h-2.5 w-2.5" />
            Zerado
          </span>
        )}
        {jaNaLista && (
          <span
            className={[
              "absolute bottom-1.5 left-1.5 rounded-full",
              "border border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)] px-[7px]",
              "py-[2px] text-[10px] font-bold text-[var(--hall-ouro)]",
            ].join(" ")}
          >
            Já na lista
          </span>
        )}
      </div>
      <span
        className={`line-clamp-2 text-[13px] font-semibold
          ${marcado ? "text-[var(--text-primary)]" : "text-[var(--lista-text-secondary)]"}`}
      >
        {item.nome}
      </span>
      {item.ano_lancamento && (
        <span className="text-[11px] text-[var(--lista-text-dim)]">{item.ano_lancamento}</span>
      )}
    </label>
  );
}
