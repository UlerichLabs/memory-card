import { useState } from "react";
import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { Gamepad2, Star } from "lucide-react";

import { buscarJogo } from "@/lib/api";
import { formatDate, formatTime } from "@/lib/formatters";
import { DifficultyBadge } from "@/components/DifficultyBadge";
import { GameDeleteConfirm } from "@/components/GameDeleteConfirm";
import { NotaBadge } from "@/components/NotaBadge";
import { Button } from "@/components/ui/button";

type GameDetailProps = {
  jogoId: number;
  onEdit?: () => void;
  onDelete?: (jogoId: number) => void;
};

export function GameDetail({ jogoId, onEdit, onDelete }: GameDetailProps) {
  const [expanded, setExpanded] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const { data: jogo, isLoading } = useQuery({
    queryKey: ["jogo", jogoId],
    queryFn: () => buscarJogo(jogoId)
  });

  if (isLoading) {
    return <div className="min-h-[60vh] rounded-lg bg-surface" />;
  }

  if (!jogo) {
    return <p className="text-muted">Jogo não encontrado.</p>;
  }

  const descricaoLonga = (jogo.igdbDescricao?.length ?? 0) > 420;
  const descricao = descricaoLonga && !expanded ? `${jogo.igdbDescricao?.slice(0, 420)}...` : jogo.igdbDescricao;

  return (
    <div className="grid gap-8 lg:grid-cols-[300px_1fr]">
      <div className="aspect-[3/4] w-full max-w-[300px] overflow-hidden rounded-lg bg-surface-raised">
        {jogo.igdbCapaUrl ? (
          <img src={jogo.igdbCapaUrl} alt="" className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full items-center justify-center text-muted">
            <Gamepad2 className="h-14 w-14" />
          </div>
        )}
      </div>
      <section className="space-y-6">
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-3xl font-bold text-foreground">{jogo.nome}</h1>
            {jogo.destaque && (
              <span className="inline-flex items-center gap-1 rounded-md border border-gold bg-gold-dim px-2 py-1 text-xs text-gold">
                <Star className="h-3 w-3" /> Destaque
              </span>
            )}
          </div>
          <p className="text-muted">{jogo.console}</p>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Info label="Nota" value={<NotaBadge nota={jogo.nota} />} />
          <Info label="Dificuldade" value={<DifficultyBadge dificuldade={jogo.dificuldade} />} />
          <Info label="Conclusão" value={<span>{formatDate(jogo.finalizadoEm)}</span>} />
          <Info label="Tempo" value={<span className="font-mono">{formatTime(jogo.tempoJogado)}</span>} />
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <Info label="Gênero" value={<span>{jogo.genero}</span>} />
          <Info label="Tipo" value={<span>{jogo.tipo}</span>} />
        </div>
        {jogo.review && <TextBlock label="Review" text={jogo.review} />}
        {jogo.igdbDescricao && (
          <div className="rounded-lg border border-border bg-surface p-4">
            <p className="mb-2 text-xs uppercase tracking-wide text-muted">Sinopse IGDB</p>
            <p className="text-sm leading-6 text-foreground">{descricao}</p>
            {descricaoLonga && (
              <Button type="button" variant="link" className="mt-2 px-0" onClick={() => setExpanded((value) => !value)}>
                {expanded ? "Mostrar menos" : "Mostrar mais"}
              </Button>
            )}
          </div>
        )}
        <div className="flex gap-3">
          <Button type="button" onClick={onEdit}>Editar</Button>
          <Button type="button" variant="destructive" onClick={() => setConfirmOpen(true)}>Excluir</Button>
        </div>
      </section>
      {confirmOpen && (
        <GameDeleteConfirm
          jogoId={jogo.id}
          nomeJogo={jogo.nome}
          onCancel={() => setConfirmOpen(false)}
          onConfirm={(id) => onDelete?.(id)}
        />
      )}
    </div>
  );
}

function Info({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-surface p-4">
      <p className="mb-2 text-xs uppercase tracking-wide text-muted">{label}</p>
      {value}
    </div>
  );
}

function TextBlock({ label, text }: { label: string; text: string }) {
  return (
    <div className="rounded-lg border border-border bg-surface p-4">
      <p className="mb-2 text-xs uppercase tracking-wide text-muted">{label}</p>
      <p className="text-sm leading-6 text-foreground">{text}</p>
    </div>
  );
}
