import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import { GameRow, gameRowGridClass } from "@/components/GameRow";
import { BibliotecaSkeleton } from "@/components/biblioteca/BibliotecaSkeleton";
import { BibliotecaVazia } from "@/components/biblioteca/BibliotecaVazia";
import { SearchBar } from "@/components/biblioteca/SearchBar";
import { Button } from "@/components/ui/button";
import { useJogos } from "@/hooks/useJogos";
import { useBibliotecaStore } from "@/store/bibliotecaStore";

export default function BibliotecaPage() {
  const navigate = useNavigate();
  const busca = useBibliotecaStore((state) => state.busca);
  const [page, setPage] = useState(1);
  const { jogos, total, totalPaginas, isLoading } = useJogos({ page, limit: 20, q: busca });

  useEffect(() => {
    setPage(1);
  }, [busca]);

  return (
    <main className="min-h-screen bg-background p-4 text-foreground md:p-6">
      <header className="mb-6 flex items-center justify-between gap-4">
        <div className="flex items-end gap-3">
          <h1 className="text-2xl font-bold">Biblioteca</h1>
          <span className="pb-1 font-mono text-xs text-muted">{total} zeramentos</span>
        </div>
      </header>
      <SearchBar semResultados={!isLoading && busca.length > 0 && jogos.length === 0} />
      {isLoading ? (
        <div className="overflow-x-auto rounded-lg border border-border bg-background">
          <ListHeader />
          <BibliotecaSkeleton />
        </div>
      ) : busca.length === 0 && jogos.length === 0 ? (
        <BibliotecaVazia onRegistrar={() => navigate("/biblioteca/novo")} />
      ) : jogos.length === 0 ? null : (
        <>
          <div className="overflow-x-auto rounded-lg border border-border bg-background">
            <ListHeader />
            {jogos.map((jogo) => (
              <GameRow key={jogo.id} jogo={jogo} onClick={() => navigate(`/biblioteca/${jogo.id}`)} />
            ))}
          </div>
          <footer className="mt-6 flex items-center justify-center gap-3">
            <Button type="button" variant="secondary" disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>
              Anterior
            </Button>
            <span className="font-mono text-sm text-muted">
              {page}/{totalPaginas}
            </span>
            <Button
              type="button"
              variant="secondary"
              disabled={page >= totalPaginas}
              onClick={() => setPage((value) => value + 1)}
            >
              Próxima
            </Button>
          </footer>
        </>
      )}
    </main>
  );
}

function ListHeader() {
  return (
    <div className={`${gameRowGridClass} border-b-[0.5px] border-border px-5 py-2 font-mono text-[10px] uppercase tracking-[0.6px] text-muted`}>
      <span>Jogo</span>
      <span>Início</span>
      <span>Conclusão</span>
      <span>Tempo</span>
      <span>Plataforma</span>
      <span>Dificuldade</span>
      <span>Nota</span>
    </div>
  );
}
