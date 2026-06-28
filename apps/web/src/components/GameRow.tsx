import { Gamepad2 } from "lucide-react";
import type { JogoZerado } from "@memory-card/types";

import { DifficultyBadge } from "@/components/DifficultyBadge";
import { notaClass } from "@/components/NotaBadge";
import { formatDate, formatTime } from "@/lib/formatters";

type GameRowProps = {
  jogo: JogoZerado;
  onClick: () => void;
};

const gridClass = "grid grid-cols-[minmax(220px,2fr)_90px_90px_80px_110px_70px_50px] items-center gap-4";

export function GameRow({ jogo, onClick }: GameRowProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`${gridClass} w-full cursor-pointer border-b-[0.5px] border-[rgba(255,255,255,0.05)] px-5 py-2.5 text-left transition last:border-b-0 hover:bg-[rgba(255,255,255,0.03)]`}
    >
      <span className="flex min-w-0 items-center gap-3">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-sm border-[0.5px] border-border bg-surface-raised">
          {jogo.igdbCapaUrl ? (
            <img src={jogo.igdbCapaUrl} alt="" className="h-full w-full object-cover" />
          ) : (
            <Gamepad2 className="h-4 w-4 text-muted" />
          )}
        </span>
        <span className="min-w-0">
          <span className="block truncate whitespace-nowrap text-[13px] font-medium text-foreground">{jogo.nome}</span>
          <span className="mt-px block font-mono text-[10px] text-muted">#{String(jogo.id).padStart(3, "0")}</span>
        </span>
      </span>
      <span className="font-mono text-xs text-muted">{jogo.iniciadoEm ? formatDate(jogo.iniciadoEm) : "—"}</span>
      <span className="font-mono text-xs text-muted">{formatDate(jogo.finalizadoEm)}</span>
      <span className="font-mono text-xs text-muted">{jogo.tempoJogado > 0 ? formatTime(jogo.tempoJogado) : "—"}</span>
      <span className="truncate text-xs text-[#888888]">{jogo.console}</span>
      <span className="min-w-0">
        <DifficultyBadge dificuldade={jogo.dificuldade} small />
      </span>
      <span className={`flex h-[30px] w-[30px] items-center justify-center rounded-sm border-[0.5px] font-mono text-[13px] font-medium ${notaClass(jogo.nota)}`}>
        {jogo.nota}
      </span>
    </button>
  );
}

export { gridClass as gameRowGridClass };
