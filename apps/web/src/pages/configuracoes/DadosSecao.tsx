export function DadosSecao() {
  return (
    <section className="space-y-6 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <header>
        <h2 className="text-base font-bold text-[var(--text-primary)]">Dados</h2>
      </header>
      <div className="space-y-2 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface-alt)] p-8 text-center">
        <p className="text-sm font-semibold text-[var(--text-primary)]">Em breve</p>
        <p className="text-[13px] text-[var(--text-secondary)]">
          O gerenciamento e exportação de dados estarão disponíveis em breve.
        </p>
      </div>
    </section>
  )
}
