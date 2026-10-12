import { Outlet } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { ConfiguracoesNav } from '@/components/configuracoes/ConfiguracoesNav'
import { useDocumentTitle } from '@/lib/useDocumentTitle'

export function ConfiguracoesLayout() {
  useDocumentTitle('Configurações')

  return (
    <div className="min-h-full bg-[var(--bg-primary)] text-[var(--text-primary)]">
      <Topbar />
      <main className="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
        <header className="space-y-1">
          <h1 className="text-[22px] font-bold text-[var(--text-primary)]">Configurações</h1>
          <p className="text-[13px] text-[var(--text-secondary)]">
            Gerencie seu perfil, sua conta e como o Memory Card aparece para você.
          </p>
        </header>

        <div className="flex flex-col gap-4 md:grid md:grid-cols-[240px_1fr] md:gap-8 md:items-start">
          <ConfiguracoesNav />
          <div className="min-w-0 flex-1">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  )
}
