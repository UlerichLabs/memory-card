import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Gamepad2 } from "lucide-react";
import { useNavigate } from "react-router-dom";

import { GameRow, gameRowGridClass } from "@/components/GameRow";
import { Button } from "@/components/ui/button";
import { listarJogos } from "@/lib/api";

export default function BibliotecaPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const { data, isLoading } = useQuery({
    queryKey: ["jogos", { page }],
    queryFn: () => listarJogos({ page, limit: 20 })
  });
  const jogos = data?.data ?? [];

  return (
    <main className="min-h-screen bg-background p-4 text-foreground md:p-6">
      <header className="mb-6 flex items-center justify-between gap-4">
        <div className="flex items-end gap-3">
          <h1 className="text-2xl font-bold">Biblioteca</h1>
          <span className="pb-1 font-mono text-xs text-muted">{data?.total ?? 0} zeramentos</span>
        </div>
      </header>
      {isLoading ? (
        <div className="overflow-x-auto rounded-lg border border-border bg-background">
          <ListHeader />
          {Array.from({ length: 5 }, (_, index) => (
            <div key={index} className={`${gameRowGridClass} border-b-[0.5px] border-[rgba(255,255,255,0.05)] px-5 py-2.5 last:border-b-0`}>
              <div className="flex items-center gap-3">
                <div className="h-9 w-9 animate-pulse rounded-sm bg-surface-raised" />
                <div className="space-y-1.5">
                  <div className="h-3 w-40 animate-pulse rounded bg-surface-raised" />
                  <div className="h-2 w-10 animate-pulse rounded bg-surface-raised" />
                </div>
              </div>
              {Array.from({ length: 6 }, (_, cellIndex) => (
                <div key={cellIndex} className="h-3 animate-pulse rounded bg-surface-raised" />
              ))}
            </div>
          ))}
        </div>
      ) : jogos.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-4 py-12 text-center">
          <Gamepad2 className="h-12 w-12 text-muted" />
          <p className="text-lg font-medium">Nenhum zeramento ainda</p>
          <Button type="button" onClick={() => navigate("/biblioteca/novo")}>
            Registrar primeiro jogo
          </Button>
        </div>
      ) : (
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
              {data?.page ?? page}/{data?.totalPages ?? 1}
            </span>
            <Button
              type="button"
              variant="secondary"
              disabled={!data || page >= data.totalPages}
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
    <div className={`${gameRowGridClass} border-b-[0.5px] border-[rgba(255,255,255,0.05)] px-5 py-2 font-mono text-[10px] uppercase tracking-[0.6px] text-muted`}>
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
