export function ListasSkeleton() {
  return (
    <div className="mx-auto grid max-w-7xl grid-cols-1 items-start gap-8 px-4 py-8 sm:px-6 md:grid-cols-[300px_minmax(0,1fr)] lg:px-8">
      <div className="flex flex-col gap-5">
        <div className="h-8 w-48 animate-pulse rounded bg-[var(--lista-cover-bg)]" />
        <div className="h-4 w-64 animate-pulse rounded bg-[var(--lista-cover-bg)]" />
        <div className="h-11 w-full animate-pulse rounded-[10px] bg-[var(--lista-cover-bg)]" />
        <div className="mt-4 flex flex-col gap-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 w-full animate-pulse rounded-xl bg-[var(--lista-card-bg)]" />
          ))}
        </div>
      </div>

      <div className="flex flex-col gap-6 rounded-2xl border border-[var(--lista-card-border)] bg-[var(--lista-card-bg)] p-6 sm:p-7">
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-2">
            <div className="h-5 w-20 animate-pulse rounded-full bg-[var(--lista-cover-bg)]" />
            <div className="h-7 w-56 animate-pulse rounded bg-[var(--lista-cover-bg)]" />
          </div>
          <div className="flex gap-2">
            <div className="h-10 w-10 animate-pulse rounded-[10px] bg-[var(--lista-cover-bg)]" />
            <div className="h-10 w-10 animate-pulse rounded-[10px] bg-[var(--lista-cover-bg)]" />
          </div>
        </div>
        <div className="h-32 w-full animate-pulse rounded-xl bg-[var(--lista-cover-bg)]" />
        <div className="grid grid-cols-3 gap-4 md:grid-cols-4 xl:grid-cols-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="aspect-[3/4] w-full animate-pulse rounded-[10px] bg-[var(--lista-cover-bg)]" />
          ))}
        </div>
      </div>
    </div>
  )
}
