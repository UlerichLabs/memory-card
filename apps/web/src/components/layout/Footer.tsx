import { Disc3 } from 'lucide-react'

export function Footer() {
  return <footer className="border-t border-[var(--border)] bg-[var(--bg-surface)]/60 text-xs text-[var(--text-muted)]"><div className="mx-auto flex max-w-7xl flex-col items-center gap-2 px-4 py-6 pb-24 text-center md:flex-row md:items-center md:justify-between md:px-6 md:pb-6 lg:px-8"><div className="flex flex-col items-center gap-1 md:items-start"><div className="flex items-center gap-2"><Disc3 className="size-4 text-[var(--accent)]" aria-hidden="true" /><span className="font-semibold text-[var(--text-secondary)]">Memory Card</span><span>Seu histórico de jogos zerados.</span></div></div><div className="text-center md:text-right"><p>Desenvolvido por <span className="font-semibold text-[var(--text-secondary)]">UlerichLabs</span></p><p>© {new Date().getFullYear()} UlerichLabs. Todos os direitos reservados.</p></div></div></footer>
}
