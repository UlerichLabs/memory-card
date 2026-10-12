import { type CSSProperties } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Disc3, Search, Bell, Settings, LogOut } from 'lucide-react'
import { Menu } from '@base-ui/react/menu'
import { useAuthStore } from '@/store/authStore'
import { usePerfilStore } from '@/stores/perfilStore'
import { NovoMenu } from './NovoMenu'
import { UserAvatar } from './UserAvatar'

const navLinks = [
  { label: 'Dashboard', href: '/' },
  { label: 'Biblioteca', href: '/biblioteca' },
  { label: 'Hall da Fama', href: '/hall-da-fama' },
  { label: 'Listas e Desafios', href: '/listas' },
  { label: 'Abandonados', href: '/abandonados' },
  { label: 'Explorador', href: '/explorador' },
]

const dropdownTheme = {
  backgroundColor: '#1A1B20', color: '#EDEDED', borderColor: '#24262C',
  '--bg-surface': '#1A1B20', '--bg-surface-alt': '#1D1F25', '--border': '#24262C',
  '--border-subtle': '#2A2C33', '--text-primary': '#EDEDED', '--text-secondary': '#9A9CA5',
  '--danger': '#E05A4E', fontFamily: 'Inter, sans-serif',
} as CSSProperties

export function Topbar() {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const { sessao, logout } = useAuthStore()
  const { perfil } = usePerfilStore()
  const nomeUsuario = perfil?.nome ?? sessao?.usuario?.nome ?? 'Jogador'
  const fotoUsuario = perfil?.avatar_url ?? perfil?.foto_url ?? null

  async function handleLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <header className="sticky top-0 z-50 h-[60px] border-b border-[var(--border)] bg-[var(--bg-surface)]/95 backdrop-blur">
      <div className="mx-auto flex h-full max-w-7xl items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-6 xl:gap-8">
          <Link to="/" className="flex shrink-0 items-center gap-2 font-bold tracking-tight text-[var(--text-primary)]">
            <Disc3 className="h-5 w-5 text-[var(--accent)]" aria-hidden="true" />
            <span>Memory Card</span>
          </Link>

          <nav aria-label="Navegação Principal" className="hidden items-center gap-4 text-[13px] md:flex lg:gap-5">
            {navLinks.map((item) => {
              const ativo = item.href === '/' ? pathname === '/' : pathname.startsWith(item.href)
              return (
                <Link
                  key={item.label}
                  to={item.href}
                  className={`flex h-[60px] items-center border-b-2 transition-colors ${
                    ativo
                      ? 'border-[var(--nav-link-active-border)] font-semibold text-[var(--nav-link-active-text)]'
                      : 'border-transparent font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)]'
                  }`}
                >
                  {item.label}
                </Link>
              )
            })}
          </nav>
        </div>

        <div className="relative mx-2 hidden max-w-xs flex-1 sm:block md:max-w-sm lg:max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-faint)]" aria-hidden="true" />
          <input
            type="search"
            placeholder="Buscar jogo ou usuário…"
            aria-label="Buscar jogo ou usuário"
            className="h-9 w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] pl-9 pr-3 text-[13px] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
          />
        </div>

        <div className="flex items-center gap-3 sm:gap-4">
          <button
            type="button"
            aria-label="Notificações"
            className="flex h-9 w-9 items-center justify-center rounded-[7px] text-[var(--text-secondary)] transition-colors hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)]"
          >
            <Bell className="h-4 w-4" aria-hidden="true" />
          </button>

          <Menu.Root>
            <Menu.Trigger
              type="button"
              className="flex h-9 cursor-pointer items-center gap-2 rounded-full border border-[var(--border-subtle)] pl-1.5 pr-3 text-[13px] text-[var(--text-secondary)] transition-colors hover:text-[var(--text-primary)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]"
              aria-label="Perfil do usuário"
            >
              <UserAvatar nome={nomeUsuario} src={fotoUsuario} size="sm" />
              <span className="hidden font-medium text-[var(--text-primary)] sm:inline">
                {nomeUsuario}
              </span>
            </Menu.Trigger>

            <Menu.Portal>
              <Menu.Positioner
                side="bottom"
                align="end"
                sideOffset={8}
                className="z-50 outline-none"
                style={{ zIndex: 50 }}
              >
                <Menu.Popup
                  style={dropdownTheme}
                  className="z-50 w-60 rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-1.5 shadow-2xl shadow-black/60 outline-none"
                >
                  <div className="flex items-center gap-2.5 border-b border-[var(--border)] px-3 py-2.5">
                    <UserAvatar nome={nomeUsuario} src={fotoUsuario} size="md" />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-[13px] font-semibold text-[var(--text-primary)]">{nomeUsuario}</p>
                    </div>
                  </div>
                  <div className="pt-1">
                    <Menu.Item
                      className="flex w-full cursor-pointer select-none items-center gap-2.5 rounded-[6px] px-3 py-2 text-left outline-none transition-colors hover:bg-[var(--bg-surface-alt)] focus:bg-[var(--bg-surface-alt)] data-highlighted:bg-[var(--bg-surface-alt)]"
                      onClick={() => navigate('/configuracoes')}
                    >
                      <Settings className="h-4 w-4 shrink-0 text-[var(--text-secondary)]" aria-hidden="true" />
                      <div className="flex flex-col">
                        <span className="text-[13px] font-medium text-[var(--text-primary)]">Configurações</span>
                        <span className="text-[11px] text-[var(--text-secondary)]">Perfil, conta e aparência</span>
                      </div>
                    </Menu.Item>
                    <Menu.Item
                      className="flex w-full cursor-pointer select-none items-center gap-2.5 rounded-[6px] px-3 py-2 text-left text-[13px] font-medium text-[var(--danger)] outline-none transition-colors hover:bg-[var(--bg-surface-alt)] focus:bg-[var(--bg-surface-alt)] data-highlighted:bg-[var(--bg-surface-alt)]"
                      onClick={handleLogout}
                    >
                      <LogOut className="h-4 w-4 shrink-0 text-[var(--danger)]" aria-hidden="true" />
                      <span>Sair</span>
                    </Menu.Item>
                  </div>
                </Menu.Popup>
              </Menu.Positioner>
            </Menu.Portal>
          </Menu.Root>

          <NovoMenu />
        </div>
      </div>
    </header>
  )
}
