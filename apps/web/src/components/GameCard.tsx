import { Gamepad2 } from "lucide-react";
import type { JogoZerado } from "@memory-card/types";

import { DifficultyBadge } from "@/components/DifficultyBadge";
import { NotaBadge } from "@/components/NotaBadge";

type GameCardProps = {
  jogo: JogoZerado;
  onClick: () => void;
};

export function GameCard({ jogo, onClick }: GameCardProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="group overflow-hidden rounded-lg border border-border bg-surface text-left transition hover:border-accent"
    >
      <div className="relative aspect-[3/4] w-full bg-surface-raised">
        <span className="absolute left-2 top-2 z-10 rounded border border-border bg-background px-2 py-1 font-mono text-xs text-muted">
          #{String(jogo.id).padStart(3, "0")}
        </span>
        {jogo.igdbCapaUrl ? (
          <img src={jogo.igdbCapaUrl} alt="" className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full items-center justify-center text-muted">
            <Gamepad2 className="h-10 w-10" />
          </div>
        )}
      </div>
      <div className="space-y-2 p-3">
        <h3 className="truncate font-medium text-foreground">{jogo.nome}</h3>
        <p className="text-sm text-muted">{jogo.console}</p>
        <div className="flex items-start gap-2">
          <NotaBadge nota={jogo.nota} small />
          <DifficultyBadge dificuldade={jogo.dificuldade} small />
        </div>
      </div>
    </button>
  );
}
