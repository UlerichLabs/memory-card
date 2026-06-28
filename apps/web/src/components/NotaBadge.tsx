import { cn } from "@/lib/utils";

export const notaLabels: Record<number, string> = {
  11: "⭐ Jogo da Vida",
  10: "Obra-prima",
  9: "Excepcional",
  8: "Ótimo",
  7: "Bom",
  6: "Decente",
  5: "Médio",
  4: "Abaixo da média",
  3: "Ruim",
  2: "Terrível",
  1: "💀 Tragédia"
};

export function notaClass(nota: number) {
  if (nota === 11) return "border-gold bg-gold-dim text-gold";
  if (nota <= 2) return "border-red bg-surface text-red";
  if (nota >= 9) return "border-accent bg-accent-dim text-accent";
  if (nota >= 7) return "border-foreground bg-surface text-foreground";
  return "border-border bg-surface-raised text-muted";
}

type NotaBadgeProps = {
  nota: number;
  small?: boolean;
};

export function NotaBadge({ nota, small = false }: NotaBadgeProps) {
  return (
    <div className="flex flex-col items-center gap-1">
      <span
        className={cn(
          "flex items-center justify-center rounded-full border font-mono font-bold",
          notaClass(nota),
          small ? "h-8 w-8 text-xs" : "h-12 w-12 text-base"
        )}
      >
        {nota}
      </span>
      <span className={cn("text-center text-muted", small ? "text-[10px]" : "text-xs")}>
        {notaLabels[nota]}
      </span>
    </div>
  );
}
