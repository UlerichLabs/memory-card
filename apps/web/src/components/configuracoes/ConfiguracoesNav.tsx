import { Database, Info, Palette, Shield, SlidersHorizontal, User } from 'lucide-react'
import { NavLink } from 'react-router-dom'

const itensDesktop = [
  { label: 'Perfil', href: '/configuracoes/perfil', icon: User },
  { label: 'Conta e segurança', href: '/configuracoes/conta', icon: Shield },
  { label: 'Aparência', href: '/configuracoes/aparencia', icon: Palette },
  { label: 'Preferências', icon: SlidersHorizontal, disabled: true, badge: 'Em breve' },
  { label: 'Dados', href: '/configuracoes/dados', icon: Database, badge: 'Em breve' },
  { label: 'Sobre o app', href: '/configuracoes/sobre', icon: Info },
]

const itensMobile = [
  { label: 'Perfil', href: '/configuracoes/perfil' },
  { label: 'Conta', href: '/configuracoes/conta' },
  { label: 'Aparência', href: '/configuracoes/aparencia' },
  { label: 'Dados', href: '/configuracoes/dados' },
  { label: 'Sobre', href: '/configuracoes/sobre' },
]

export function ConfiguracoesNav() {
  return (
    <>
      <nav aria-label="Navegação mobile de configurações" className="flex items-center gap-2 overflow-x-auto pb-2 md:hidden">
        {itensMobile.map(({ label, href }) => (
          <NavLink
            key={href}
            to={href}
            className={({ isActive }) =>
              `inline-flex shrink-0 items-center rounded-full px-3.5 py-1.5 text-xs transition-colors ${
                isActive
                  ? 'border border-[var(--accent)] bg-[var(--accent)]/15 font-semibold text-[var(--accent)]'
                  : 'border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)]'
              }`
            }
          >
            {label}
          </NavLink>
        ))}
      </nav>

      <nav aria-label="Menu lateral de configurações" className="hidden md:flex md:w-56 md:flex-col md:gap-1">
        {itensDesktop.map((item) => {
          const Icon = item.icon
          if (item.disabled) {
            return (
              <button
                key={item.label}
                type="button"
                disabled
                aria-disabled="true"
                className="flex w-full cursor-not-allowed items-center justify-between rounded-[8px] px-3 py-2 text-left text-[13px] font-medium text-[var(--text-faint)] opacity-60"
              >
                <div className="flex items-center gap-2.5">
                  <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
                  <span>{item.label}</span>
                </div>
                {item.badge && (
                  <span className="rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-1.5 py-0.5 text-[10.5px] font-medium text-[var(--text-muted)]">
                    {item.badge}
                  </span>
                )}
              </button>
            )
          }

          return (
            <NavLink
              key={item.label}
              to={item.href!}
              className={({ isActive }) =>
                `flex items-center justify-between rounded-[8px] px-3 py-2 text-[13px] transition-colors ${
                  isActive
                    ? 'bg-[var(--bg-surface-alt)] font-semibold text-[var(--accent)]'
                    : 'font-medium text-[var(--text-secondary)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)]'
                }`
              }
            >
              <div className="flex items-center gap-2.5">
                <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
                <span>{item.label}</span>
              </div>
              {item.badge && (
                <span className="rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-1.5 py-0.5 text-[10.5px] font-medium text-[var(--text-muted)]">
                  {item.badge}
                </span>
              )}
            </NavLink>
          )
        })}
      </nav>
    </>
  )
}
