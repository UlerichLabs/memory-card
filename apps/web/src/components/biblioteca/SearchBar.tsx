import { useEffect, useState } from "react";
import { Search, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useBibliotecaStore } from "@/store/bibliotecaStore";

type SearchBarProps = {
  semResultados?: boolean;
};

export function SearchBar({ semResultados = false }: SearchBarProps) {
  const busca = useBibliotecaStore((state) => state.busca);
  const setBusca = useBibliotecaStore((state) => state.setBusca);
  const [texto, setTexto] = useState(busca);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      setBusca(texto.trim());
    }, 300);

    return () => window.clearTimeout(timeoutId);
  }, [setBusca, texto]);

  function limparBusca() {
    setTexto("");
    setBusca("");
  }

  return (
    <div className="mb-4 space-y-3">
      <div className="relative max-w-xl">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
        <Input
          type="search"
          value={texto}
          onChange={(event) => setTexto(event.target.value)}
          placeholder="Buscar na biblioteca"
          className="h-10 border-border bg-surface-raised pl-9 pr-10 text-foreground placeholder:text-muted focus-visible:ring-accent"
        />
        {texto.length > 0 && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Limpar busca"
            onClick={limparBusca}
            className="absolute right-1 top-1/2 h-8 w-8 -translate-y-1/2 text-muted hover:bg-accent-dim hover:text-accent"
          >
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>
      {semResultados && busca.length > 0 && (
        <p className="text-sm text-muted">
          Nenhum jogo encontrado para <span className="font-medium text-foreground">'{busca}'</span>
        </p>
      )}
    </div>
  );
}
