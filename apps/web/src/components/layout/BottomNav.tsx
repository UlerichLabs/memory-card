import { useState } from 'react'
import { Ban, ChevronRight, Compass, Ellipsis, House, Library, ListChecks, Trophy, User } from 'lucide-react'
import { NavLink, useLocation } from 'react-router-dom'
import { Dialog, DialogContent } from '@/components/ui/dialog'

const itens = [
  { label: 'Dashboard', href: '/', icon: House, end: true },
  { label: 'Biblioteca', href: '/biblioteca', icon: Library },
  { label: 'Hall da Fama', href: '/hall-da-fama', icon: Trophy, end: true },
  { label: 'Listas', href: '/listas', icon: ListChecks },
]
const maisItens = [
  { label: 'Abandonados', descricao: 'Jogos que você largou', href: '/abandonados', icon: Ban },
  { label: 'Explorador', descricao: 'Descobrir novos jogos', href: '/explorador', icon: Compass },
  { label: 'Meu perfil', descricao: 'Conta e senha', href: '/conta/trocar-senha', icon: User },
]
function linkClass({ isActive }: { isActive: boolean }) { return `flex min-h-14 min-w-0 flex-1 flex-col items-center justify-center gap-1 text-[10.5px] font-medium ${isActive ? 'text-[var(--accent)]' : 'text-[var(--text-secondary)]'}` }
export function BottomNav() {
  const { pathname } = useLocation(); const [maisAberto, setMaisAberto] = useState(false); const maisAtivo = maisItens.some((item) => pathname === item.href || pathname.startsWith(`${item.href}/`))
  return <><nav aria-label="Navegação inferior" className={`fixed inset-x-0 bottom-0 flex border-t border-[var(--border)] bg-[var(--bg-surface)]/95 px-2 backdrop-blur md:hidden ${maisAberto ? 'z-[60]' : 'z-40'}`} style={{ paddingBottom: 'env(safe-area-inset-bottom)' }}>{itens.map(({ label, href, icon: Icon, end }) => <NavLink key={href} to={href} end={end} className={linkClass}>{() => <><Icon className="size-[23px]" aria-hidden="true" /><span>{label}</span></>}</NavLink>)}<button type="button" aria-expanded={maisAberto} aria-haspopup="dialog" onClick={() => setMaisAberto(true)} className={`flex min-h-14 min-w-0 flex-1 flex-col items-center justify-center gap-1 text-[10.5px] font-medium ${maisAtivo || maisAberto ? 'text-[var(--accent)]' : 'text-[var(--text-secondary)]'}`}><Ellipsis className="size-[23px]" aria-hidden="true" /><span>Mais</span></button></nav><Dialog open={maisAberto} onOpenChange={setMaisAberto}><DialogContent showCloseButton={false} className="top-auto bottom-[calc(70px+env(safe-area-inset-bottom))] left-2 mx-2 w-[calc(100%-1rem)] max-w-none translate-x-0 translate-y-0 rounded-[20px] border border-[var(--border)] bg-[var(--bg-surface)] p-3 pb-4 sm:max-w-none"><div className="mx-auto mb-3 h-1 w-10 rounded-full bg-[var(--border-subtle)]" /><div>{maisItens.map(({ label, descricao, href, icon: Icon }) => <NavLink key={href} to={href} onClick={() => setMaisAberto(false)} className="flex items-center gap-3 border-b border-[var(--border)] px-2 py-3 last:border-b-0"><span className="flex size-[38px] shrink-0 items-center justify-center rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface-alt)]"><Icon className="size-[18px] text-[var(--accent)]" aria-hidden="true" /></span><span className="min-w-0 flex-1"><span className="block text-[15px] font-semibold text-[var(--text-primary)]">{label}</span><span className="block text-xs text-[var(--text-secondary)]">{descricao}</span></span><ChevronRight className="size-5 shrink-0 text-[var(--text-muted)]" aria-hidden="true" /></NavLink>)}</div></DialogContent></Dialog></>
}
