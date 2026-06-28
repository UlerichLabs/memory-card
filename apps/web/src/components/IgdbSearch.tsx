import { useEffect, useRef, useState } from "react";
import { Loader2 } from "lucide-react";
import type { IgdbJogo } from "@memory-card/types";

import { buscarJogosIgdb } from "@/lib/api";
import { Input } from "@/components/ui/input";

type IgdbSearchProps = {
  onSelect: (jogo: IgdbJogo) => void;
  placeholder?: string;
};

export function IgdbSearch({ onSelect, placeholder = "Buscar na IGDB" }: IgdbSearchProps) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<IgdbJogo[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [active, setActive] = useState(0);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (query.trim().length < 2) {
      setResults([]);
      setOpen(false);
      return;
    }

    const controller = new AbortController();
    const timeout = window.setTimeout(async () => {
      setLoading(true);

      try {
        const jogos = await buscarJogosIgdb(query);

        if (!controller.signal.aborted) {
          setResults(jogos);
          setOpen(true);
          setActive(0);
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    }, 400);

    return () => {
      controller.abort();
      window.clearTimeout(timeout);
    };
  }, [query]);

  useEffect(() => {
    function handleClick(event: MouseEvent) {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  function selectGame(jogo: IgdbJogo) {
    onSelect(jogo);
    setQuery(jogo.nome);
    setOpen(false);
  }

  function handleKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (!open) return;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActive((current) => Math.min(current + 1, results.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive((current) => Math.max(current - 1, 0));
    } else if (event.key === "Enter" && results[active]) {
      event.preventDefault();
      selectGame(results[active]);
    } else if (event.key === "Escape") {
      setOpen(false);
    }
  }

  return (
    <div ref={rootRef} className="relative">
      <div className="relative">
        <Input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className="bg-surface-raised"
        />
        {loading && <Loader2 className="absolute right-3 top-2.5 h-4 w-4 animate-spin text-muted" />}
      </div>
      {open && (
        <div className="absolute z-40 mt-2 w-full overflow-hidden rounded-lg border border-border bg-surface shadow-xl">
          {results.length === 0 && !loading ? (
            <p className="p-3 text-sm text-muted">Nenhum resultado encontrado</p>
          ) : (
            results.map((jogo, index) => (
              <button
                type="button"
                key={jogo.igdbId}
                onClick={() => selectGame(jogo)}
                className={`flex w-full gap-3 p-2 text-left hover:bg-surface-raised ${index === active ? "bg-surface-raised" : ""}`}
              >
                {jogo.capaUrl ? (
                  <img src={jogo.capaUrl} alt="" className="h-14 w-10 rounded object-cover" />
                ) : (
                  <div className="h-14 w-10 rounded bg-surface-raised" />
                )}
                <span className="min-w-0">
                  <span className="block truncate text-sm font-medium text-foreground">{jogo.nome}</span>
                  <span className="block truncate text-xs text-muted">
                    {[jogo.ano, jogo.plataformas.slice(0, 3).join(", ")].filter(Boolean).join(" · ")}
                  </span>
                </span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
