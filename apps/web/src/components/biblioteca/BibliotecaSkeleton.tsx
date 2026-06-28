import { gameRowGridClass } from "@/components/GameRow";

export function BibliotecaSkeleton() {
  return (
    <>
      {Array.from({ length: 8 }, (_, index) => (
        <div key={index} className={`${gameRowGridClass} border-b-[0.5px] border-border px-5 py-2.5 last:border-b-0`}>
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
    </>
  );
}
