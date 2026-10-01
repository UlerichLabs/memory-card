export function JogoDetalheSkeleton() {
  return (
    <div className="mx-auto max-w-7xl animate-pulse px-4 py-7 sm:px-10 sm:py-8 lg:py-12 space-y-7">
      <div className="hidden sm:flex items-center justify-between">
        <div className="h-10 w-44 rounded-[10px] bg-[var(--detalhe-bloco-bg)]" />
        <div className="flex gap-2.5">
          <div className="h-10 w-24 rounded-[10px] bg-[var(--detalhe-bloco-bg)]" />
          <div className="h-10 w-24 rounded-[10px] bg-[var(--detalhe-bloco-bg)]" />
        </div>
      </div>

      <div className="grid grid-cols-1 items-start gap-8 sm:gap-10 lg:grid-cols-[320px_minmax(0,1fr)] lg:gap-12">
        <div className="flex flex-col items-center sm:items-start gap-3.5">
          <div className="aspect-[3/4] w-[200px] sm:w-[240px] lg:w-[320px] rounded-[14px] lg:rounded-[16px] bg-[var(--detalhe-bloco-bg)]" />
          <div className="h-4 w-32 rounded bg-[var(--detalhe-bloco-bg)]" />
        </div>

        <div className="flex flex-col gap-6">
          <div className="space-y-3">
            <div className="h-8 w-2/3 rounded-[8px] bg-[var(--detalhe-bloco-bg)]" />
            <div className="flex gap-2">
              <div className="h-7 w-24 rounded-full bg-[var(--detalhe-bloco-bg)]" />
              <div className="h-7 w-20 rounded-full bg-[var(--detalhe-bloco-bg)]" />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 lg:grid-cols-4 lg:gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-28 rounded-[14px] bg-[var(--detalhe-bloco-bg)]" />
            ))}
          </div>

          <div className="h-36 rounded-[14px] bg-[var(--detalhe-bloco-bg)]" />
        </div>
      </div>
    </div>
  )
}
