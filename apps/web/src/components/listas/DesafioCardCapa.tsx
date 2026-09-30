import { Check, Trophy, X } from "lucide-react";
import { NotaBadge } from "@/components/jogos/NotaBadge";
import type { ListaItem } from "@/types/listas";

export interface DesafioCardCapaProps {
  item: ListaItem;
  zerado: boolean;
  capaUrl: string | null;
  capaUrl2x: string | null;
  iniciais: string;
  onRemover: (id: number) => void;
  onZerei: (item: ListaItem) => void;
}

export function DesafioCardCapa({
  item,
  zerado,
  capaUrl,
  capaUrl2x,
  iniciais,
  onRemover,
  onZerei,
}: DesafioCardCapaProps) {
  return (
    <div
      className={`relative aspect-[3/4] w-full overflow-hidden rounded-[10px]
        bg-[var(--lista-cover-bg)] transition-transform duration-200 ${
          zerado
            ? "border border-[var(--lista-card-zerado-border)] group-hover:-translate-y-1"
            : "border border-dashed border-[var(--lista-card-pendente-border)]"
        }`}
    >
      <div className="h-full w-full" style={{ opacity: zerado ? 1 : 0.55 }}>
        {capaUrl ? (
          <img
            src={capaUrl}
            srcSet={capaUrl2x ? `${capaUrl} 1x, ${capaUrl2x} 2x` : undefined}
            alt={item.nome}
            className="h-full w-full object-cover"
            loading="lazy"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <span
              className={`select-none text-[26px] font-bold ${
                zerado
                  ? "text-[var(--lista-initials-zerado)]"
                  : "text-[var(--lista-initials-pendente)]"
              }`}
            >
              {iniciais}
            </span>
          </div>
        )}
      </div>

      {zerado && (
        <>
          <div
            className={[
              "absolute left-1.5 top-1.5 flex",
              "h-6 w-6 items-center justify-center",
              "rounded-full bg-[var(--lista-progress-fill)]",
              "text-[var(--ouro-jogo-ano-text)] shadow",
            ].join(" ")}
          >
            <Check className="h-3.5 w-3.5 stroke-[3]" />
          </div>
          {item.jogo_zerado && (
            <div className="absolute bottom-1.5 right-1.5">
              <NotaBadge nota={item.jogo_zerado.nota} tamanho="sm" />
            </div>
          )}
        </>
      )}

      {!zerado && (
        <span
          className={[
            "absolute bottom-1.5 left-1.5 rounded-full",
            "border border-[var(--lista-pill-pendente-border)]",
            "bg-[var(--lista-pill-pendente-bg)] px-[7px] py-[3px]",
            "text-[10px] font-semibold text-[var(--lista-pill-pendente-text)]",
          ].join(" ")}
        >
          Pendente
        </span>
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
          className="absolute bottom-1.5 right-1.5 flex h-11 w-11 items-center justify-center
            rounded-lg border border-[var(--lista-btn-zerei-border)]
            bg-[var(--lista-btn-zerei-bg)] text-[var(--lista-btn-zerei-text)]
            hover:opacity-90 sm:hidden"
        >
          <Trophy className="h-5 w-5" />
        </button>
      )}

      <button
        type="button"
        aria-label="Tirar do desafio"
        onClick={(event) => {
          event.preventDefault();
          event.stopPropagation();
          onRemover(item.id);
        }}
        className={[
          "absolute right-1.5 top-1.5 flex h-7 w-7 items-center justify-center",
          "rounded-full border border-[var(--lista-btn-icon-border)]",
          "bg-[var(--lista-card-acao-bg)] text-[var(--lista-text-muted)]",
          "opacity-0 transition-opacity hover:text-[var(--danger)]",
          "focus:opacity-100 group-hover:opacity-100 max-sm:opacity-100",
        ].join(" ")}
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}
