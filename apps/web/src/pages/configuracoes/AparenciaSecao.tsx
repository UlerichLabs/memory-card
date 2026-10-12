import { Hammer } from 'lucide-react'

export function AparenciaSecao() {
  return (
    <section className="space-y-6 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <header>
        <h2 className="text-base font-bold text-[var(--text-primary)]">Aparência</h2>
      </header>
      <div className="space-y-3 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface-alt)] p-8 text-center">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface)] text-[var(--accent)]">
          <Hammer className="h-6 w-6" aria-hidden="true" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-semibold text-[var(--text-primary)]">Em construção</p>
          <p className="text-[13px] text-[var(--text-secondary)]">
            Este módulo faz parte das próximas etapas do roadmap e estará disponível em breve.
          </p>
        </div>
      </div>
    </section>
  )
}
