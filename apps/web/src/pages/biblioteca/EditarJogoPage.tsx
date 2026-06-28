import { useQuery } from "@tanstack/react-query";
import { useNavigate, useParams } from "react-router-dom";

import { GameForm } from "@/components/GameForm";
import { buscarJogo } from "@/lib/api";

export default function EditarJogoPage() {
  const navigate = useNavigate();
  const id = Number(useParams().id);
  const { data: jogo, isLoading } = useQuery({
    queryKey: ["jogo", id],
    queryFn: () => buscarJogo(id),
    enabled: Number.isFinite(id)
  });

  return (
    <main className="min-h-screen bg-background p-4 text-foreground md:p-6">
      <p className="mb-4 text-sm text-muted">Biblioteca &gt; {jogo?.nome ?? "Jogo"} &gt; Editar</p>
      <h1 className="mb-6 text-2xl font-bold">Editar zeramento</h1>
      {isLoading || !jogo ? <div className="h-96 rounded-lg bg-surface" /> : (
        <GameForm jogoInicial={jogo} onSuccess={(updated) => navigate(`/biblioteca/${updated.id}`, { replace: true })} />
      )}
    </main>
  );
}
