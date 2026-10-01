import { Play } from 'lucide-react'

export function JogandoVazio({ onIniciar }: { onIniciar: () => void }) {
  return (
    <section className="flex items-center gap-3 rounded-[10px] border border-dashed border-[var(--border-subtle)] px-4 py-4">
      <Play className="size-5 shrink-0 text-[var(--text-faint)]" aria-hidden="true" />
      <div className="min-w-0 flex-1">
        <h2 className="text-sm font-bold text-[var(--text-primary)]">Jogando agora</h2>
        <p className="mt-1 text-xs text-[var(--text-muted)]">Nada em andamento. Inicie um jogo para guardar quando você começou, sem precisar lembrar depois.</p>
      </div>
      <button type="button" onClick={onIniciar} className="shrink-0 text-xs font-semibold text-[var(--accent)] hover:underline">Iniciar um jogo →</button>
    </section>
  )
}
