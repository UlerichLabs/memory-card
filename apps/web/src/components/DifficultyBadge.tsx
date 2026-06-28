import { DIFICULDADES, type Dificuldade } from "@memory-card/types";

import { cn } from "@/lib/utils";

export const difficultyConfig: Record<Dificuldade, { className: string }> = {
  AAA: { className: "border-red bg-surface text-red" },
  AA: { className: "border-orange bg-surface text-orange" },
  A: { className: "border-gold bg-gold-dim text-gold" },
  B: { className: "border-accent bg-accent-dim text-accent" },
  C: { className: "border-muted bg-surface text-muted" }
};

type DifficultyBadgeProps = {
  dificuldade: Dificuldade;
  small?: boolean;
};

export function DifficultyBadge({ dificuldade, small = false }: DifficultyBadgeProps) {
  const config = difficultyConfig[dificuldade];
  const dificuldadeInfo = DIFICULDADES.find((item) => item.valor === dificuldade);

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border font-mono font-semibold",
        config.className,
        small ? "px-2 py-1 text-[10px]" : "px-2.5 py-1 text-xs"
      )}
    >
      {dificuldade} · {dificuldadeInfo?.label ?? dificuldade}
    </span>
  );
}
