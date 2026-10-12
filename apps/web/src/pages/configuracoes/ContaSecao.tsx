import { TrocarSenhaForm } from '@/components/auth/TrocarSenhaForm'

export function ContaSecao() {
  return (
    <section className="space-y-6 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <header className="space-y-1">
        <h2 className="text-base font-bold text-[var(--text-primary)]">Trocar senha</h2>
        <p className="text-[13px] text-[var(--text-secondary)]">
          Mantenha sua conta protegida com uma senha nova.
        </p>
      </header>
      <TrocarSenhaForm />
    </section>
  )
}
