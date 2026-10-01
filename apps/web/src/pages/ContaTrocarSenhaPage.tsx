import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { TrocarSenhaForm } from '@/components/auth/TrocarSenhaForm'

const contaTheme = {
  '--bg-primary': '#15161A',
  '--bg-surface': '#1A1B20',
  '--border': '#24262C',
  '--text-primary': '#EDEDED',
  '--text-secondary': '#9A9CA5',
  '--accent': '#4F7CFF',
  backgroundColor: 'var(--bg-primary)',
  color: 'var(--text-primary)',
  fontFamily: 'Inter, sans-serif',
} as CSSProperties

export function ContaTrocarSenhaPage() {
  return (
    <div style={contaTheme} className="min-h-svh">
      <Topbar />
      <main className="mx-auto max-w-xl px-4 py-8 sm:px-6 lg:px-8">
        <section className="space-y-6 rounded-xl border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-8">
          <header className="space-y-2">
            <p className="text-xs font-medium text-[var(--text-secondary)]">Conta</p>
            <h1 className="text-[22px] font-bold">Trocar senha</h1>
            <p className="text-[13px] text-[var(--text-secondary)]">Mantenha sua conta protegida com uma senha nova.</p>
          </header>
          <TrocarSenhaForm />
          <Link to="/" className="block text-center text-[13px] font-medium text-[var(--accent)] underline-offset-4 hover:underline">
            Voltar ao dashboard
          </Link>
        </section>
      </main>
    </div>
  )
}
