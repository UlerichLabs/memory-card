import { Skeleton } from '@/components/ui/skeleton'

export function AbandonadosSkeleton() {
  const itens = Array.from({ length: 12 }, (_, i) => i)

  return (
    <div
      className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
      data-testid="abandonados-skeletons"
    >
      {itens.map((i) => (
        <div
          key={i}
          className={
            'flex flex-col rounded-xl border ' +
            'border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] p-3'
          }
        >
          <Skeleton className="aspect-[3/4] w-full rounded-[10px] bg-[var(--biblioteca-control-bg)]" />
          <div className="mt-3 space-y-2">
            <Skeleton className="h-4 w-3/4 bg-[var(--biblioteca-control-bg)]" />
            <Skeleton className="h-3 w-1/2 bg-[var(--biblioteca-control-bg)]" />
            <Skeleton className="h-3 w-2/3 bg-[var(--biblioteca-control-bg)]" />
            <div className="flex gap-1.5 pt-2">
              <Skeleton className="h-9 flex-1 rounded-lg bg-[var(--biblioteca-control-bg)]" />
              <Skeleton className="h-9 w-9 rounded-lg bg-[var(--biblioteca-control-bg)]" />
              <Skeleton className="h-9 w-9 rounded-lg bg-[var(--biblioteca-control-bg)]" />
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
