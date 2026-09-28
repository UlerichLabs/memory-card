import { Skeleton } from '@/components/ui/skeleton'

export interface BibliotecaSkeletonsProps {
  modo?: 'grid' | 'list'
}

export function BibliotecaSkeletons({ modo = 'grid' }: BibliotecaSkeletonsProps) {
  const itens = Array.from({ length: 12 }, (_, i) => i)

  if (modo === 'list') {
    return (
      <div className="flex flex-col gap-2" data-testid="biblioteca-skeletons">
        {itens.map((i) => (
          <div
            key={i}
            className="flex h-14 items-center gap-3 rounded-[8px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] px-3"
          >
            <Skeleton className="h-[42px] w-[32px] rounded-[4px] bg-[var(--biblioteca-control-bg)]" />
            <div className="flex flex-1 flex-col gap-1.5">
              <Skeleton className="h-4 w-1/3 bg-[var(--biblioteca-control-bg)]" />
              <Skeleton className="h-3 w-1/4 bg-[var(--biblioteca-control-bg)]" />
            </div>
            <Skeleton className="h-6 w-12 rounded-[4px] bg-[var(--biblioteca-control-bg)]" />
          </div>
        ))}
      </div>
    )
  }

  return (
    <div
      className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
      data-testid="biblioteca-skeletons"
    >
      {itens.map((i) => (
        <div
          key={i}
          className="flex flex-col rounded-[10px] border border-[var(--biblioteca-card-border)] bg-[var(--biblioteca-card-bg)] p-3"
        >
          <Skeleton className="aspect-[3/4] w-full rounded-[6px] bg-[var(--biblioteca-control-bg)]" />
          <div className="mt-3 space-y-2">
            <Skeleton className="h-4 w-3/4 bg-[var(--biblioteca-control-bg)]" />
            <Skeleton className="h-3 w-1/2 bg-[var(--biblioteca-control-bg)]" />
            <div className="flex justify-between pt-1">
              <Skeleton className="h-5 w-10 rounded-[4px] bg-[var(--biblioteca-control-bg)]" />
              <Skeleton className="h-5 w-8 rounded-[4px] bg-[var(--biblioteca-control-bg)]" />
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
