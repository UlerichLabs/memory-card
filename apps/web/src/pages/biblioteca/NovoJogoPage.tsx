import { useNavigate } from "react-router-dom";

import { GameForm } from "@/components/GameForm";

export default function NovoJogoPage() {
  const navigate = useNavigate();

  return (
    <main className="min-h-screen bg-background p-4 text-foreground md:p-6">
      <p className="mb-4 text-sm text-muted">Biblioteca &gt; Novo zeramento</p>
      <h1 className="mb-6 text-2xl font-bold">Novo zeramento</h1>
      <GameForm onSuccess={(jogo) => navigate(`/biblioteca/${jogo.id}`, { replace: true })} />
    </main>
  );
}
