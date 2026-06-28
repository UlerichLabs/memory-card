import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "react-router-dom";

import { GameDetail } from "@/components/GameDetail";
import { deletarJogo } from "@/lib/api";
import { useGameFormStore } from "@/store/gameFormStore";

export default function DetalheJogoPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const id = Number(useParams().id);
  const abrirEditar = useGameFormStore((state) => state.abrirEditar);
  const deleteMutation = useMutation({
    mutationFn: deletarJogo,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["jogos"] });
      navigate("/biblioteca");
    }
  });

  return (
    <main className="min-h-screen bg-background p-4 text-foreground md:p-6">
      <p className="mb-4 text-sm text-muted">Biblioteca &gt; Detalhe</p>
      <GameDetail
        jogoId={id}
        onEdit={() => abrirEditar(id)}
        onDelete={(jogoId) => deleteMutation.mutate(jogoId)}
      />
    </main>
  );
}
