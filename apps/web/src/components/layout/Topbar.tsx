import { useContext, type CSSProperties } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Disc3, Search, Bell, User, ChevronDown, LogOut } from 'lucide-react'
import { Menu } from '@base-ui/react/menu'
import { useAuthStore } from '@/store/authStore'
import { JogosContext } from '@/stores/jogosStore'

const navLinks = [
  { label: 'Dashboard', href: '/' },
  { label: 'Biblioteca', href: '/biblioteca' },
  { label: 'Abandonados', href: '/abandonados' },
  { label: 'Desafios', href: '/desafios' },
  { label: 'Listas', href: '/listas' },
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
  const jogos = useContext(JogosContext)
  const nomeUsuario = sessao?.usuario?.nome ?? 'Jogador'

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
                      ? 'border-[var(--accent)] font-semibold text-[var(--accent)]'
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
              className="flex cursor-pointer items-center gap-2 text-[13px] text-[var(--text-secondary)] transition-colors hover:text-[var(--text-primary)] focus-visible:outline-none"
              aria-label="Perfil do usuário"
            >
              <div className="flex h-7 w-7 items-center justify-center rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)]">
                <User className="h-3.5 w-3.5 text-[var(--text-secondary)]" aria-hidden="true" />
              </div>
              <span className="hidden font-medium text-[var(--text-primary)] sm:inline">
                {nomeUsuario}
              </span>
              <ChevronDown className="h-3.5 w-3.5 text-[var(--text-faint)]" aria-hidden="true" />
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
                  className="z-50 w-[190px] rounded-[10px] border border-[#24262C] bg-[#1A1B20] p-1.5 shadow-2xl shadow-black/60 outline-none"
                >
                  <Menu.Item
                    className="flex w-full cursor-pointer select-none items-center gap-2.5 rounded-[6px] px-3.5 py-2.5 text-[13px] font-medium text-[#EDEDED] outline-none transition-colors hover:bg-[#1D1F25] focus:bg-[#1D1F25] data-highlighted:bg-[#1D1F25]"
                    onClick={() => navigate('/conta/trocar-senha')}
                  >
                    <User className="h-4 w-4 text-[#9A9CA5]" aria-hidden="true" />
                    <span>Conta</span>
                  </Menu.Item>
                  <Menu.Item
                    className="flex w-full cursor-pointer select-none items-center gap-2.5 rounded-[6px] px-3.5 py-2.5 text-[13px] font-medium text-[#E05A4E] outline-none transition-colors hover:bg-[#1D1F25] focus:bg-[#1D1F25] data-highlighted:bg-[#1D1F25]"
                    onClick={handleLogout}
                  >
                    <LogOut className="h-4 w-4 text-[#E05A4E]" aria-hidden="true" />
                    <span>Sair</span>
                  </Menu.Item>
                </Menu.Popup>
              </Menu.Positioner>
            </Menu.Portal>
          </Menu.Root>

          <button
            type="button"
            onClick={() => jogos?.abrirModalRegistro()}
            className="inline-flex shrink-0 items-center justify-center rounded-[7px] bg-[var(--accent)] px-3.5 py-1.5 text-[13px] font-bold text-[#0E0F12] transition-opacity hover:opacity-90"
          >
            + Registrar jogo
          </button>
        </div>
      </div>
    </header>
  )
}
