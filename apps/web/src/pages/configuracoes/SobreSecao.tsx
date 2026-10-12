export function SobreSecao() {
  return (
    <section className="space-y-6 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <header>
        <h2 className="text-base font-bold text-[var(--text-primary)]">Sobre o app</h2>
      </header>
      <div className="space-y-4 text-[13px] text-[var(--text-secondary)]">
        <div>
          <h3 className="text-base font-semibold text-[var(--text-primary)]">Memory Card</h3>
          <p className="mt-1">
            Sua caderneta digital para registrar, organizar e celebrar sua jornada nos videogames.
          </p>
        </div>
        <p className="border-t border-[var(--border)] pt-4 text-xs text-[var(--text-muted)]">
          Nomes e logotipos de consoles e plataformas pertencem aos seus respectivos donos. O Memory Card não é afiliado a nenhum deles.
        </p>
      </div>
    </section>
  )
}
