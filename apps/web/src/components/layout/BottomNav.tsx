import { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { Ellipsis, House, Library, ListChecks, Trophy } from 'lucide-react'
import { Dialog, DialogContent } from '@/components/ui/dialog'

const itens = [
  { label: 'Dashboard', href: '/', icon: House, end: true },
  { label: 'Biblioteca', href: '/biblioteca', icon: Library },
  { label: 'Hall da Fama', href: '/hall-da-fama', icon: Trophy, end: true },
  { label: 'Listas', href: '/listas', icon: ListChecks },
]

const maisItens = [
  { label: 'Abandonados', descricao: 'Jogos que você largou', href: '/abandonados' },
  { label: 'Explorador', descricao: 'Descobrir novos jogos', href: '/explorador' },
  { label: 'Meu perfil', descricao: 'Conta e senha', href: '/conta/trocar-senha' },
]

function linkClass({ isActive }: { isActive: boolean }) {
  return `flex min-h-14 min-w-0 flex-1 flex-col items-center justify-center gap-1 text-[10.5px] font-medium ${isActive ? 'text-[var(--accent)]' : 'text-[var(--text-secondary)]'}`
}

export function BottomNav() {
  const { pathname } = useLocation()
  const [maisAberto, setMaisAberto] = useState(false)
  const maisAtivo = maisItens.some((item) => pathname === item.href || pathname.startsWith(`${item.href}/`))
  return <><nav aria-label="Navegação inferior" className="fixed inset-x-0 bottom-0 z-40 flex border-t border-[var(--border)] bg-[var(--bg-surface)]/95 px-2 backdrop-blur md:hidden" style={{ paddingBottom: 'env(safe-area-inset-bottom)' }}>{itens.map(({ label, href, icon: Icon, end }) => <NavLink key={href} to={href} end={end} className={linkClass}>{() => <><Icon className="size-[23px]" aria-hidden="true" /><span>{label}</span></>}</NavLink>)}<button type="button" aria-expanded={maisAberto} aria-haspopup="dialog" onClick={() => setMaisAberto(true)} className={`flex min-h-14 min-w-0 flex-1 flex-col items-center justify-center gap-1 text-[10.5px] font-medium ${maisAtivo || maisAberto ? 'text-[var(--accent)]' : 'text-[var(--text-secondary)]'}`}><Ellipsis className="size-[23px]" aria-hidden="true" /><span>Mais</span></button></nav><Dialog open={maisAberto} onOpenChange={setMaisAberto}><DialogContent showCloseButton={false} className="top-auto bottom-0 left-0 w-full max-w-none translate-x-0 translate-y-0 rounded-t-2xl rounded-b-none border border-[var(--border)] bg-[var(--bg-surface)] p-5 pb-8 sm:max-w-none"><div className="mx-auto mb-4 h-1 w-10 rounded-full bg-[var(--border-subtle)]" /><div className="flex flex-col gap-1">{maisItens.map((item) => <NavLink key={item.href} to={item.href} onClick={() => setMaisAberto(false)} className="rounded-lg px-3 py-3 hover:bg-[var(--bg-surface-alt)]"><span className="block text-sm font-semibold text-[var(--text-primary)]">{item.label}</span><span className="block text-xs text-[var(--text-secondary)]">{item.descricao}</span></NavLink>)}</div></DialogContent></Dialog></>
}
