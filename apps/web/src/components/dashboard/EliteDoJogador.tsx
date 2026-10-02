import { Link } from 'react-router-dom'
import { Gamepad2, Star } from 'lucide-react'
import type { DashboardGameDoAno } from '@/types/dashboard'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB } from '@/lib/utils'
import { formatarHoras, iniciais } from '@/lib/dashboardUtils'

interface EliteDoJogadorProps {
  jogoDoAno: DashboardGameDoAno | null
  jogosDaVida: JogoZeradoDTO[]
  jogosDaVidaTotal: number
}

function Capa({ jogo, pequena = false }: { jogo: JogoZeradoDTO | DashboardGameDoAno; pequena?: boolean }) {
  const url = formatarCapaIGDB(jogo.igdb_capa_url, 't_cover_big')
  return <div className={`relative overflow-hidden rounded-[6px] border border-[var(--highlight-gold)] bg-[var(--bg-surface-alt)] shadow-[0_0_8px_1px_var(--highlight-gold)] ${pequena ? 'aspect-[3/4] w-full' : 'h-28 w-[84px] shrink-0'}`}>{url ? <img src={url} alt={jogo.nome} className="h-full w-full object-cover" /> : <div className="flex h-full items-center justify-center text-xs font-bold text-[var(--text-muted)]">{iniciais(jogo.nome)}</div>}{pequena && <span className="absolute left-1 top-1 flex size-5 items-center justify-center rounded-full bg-[var(--highlight-gold)] text-[9px] font-bold text-[var(--bg-primary)]">11</span>}</div>
}

function JogoDoAno({ jogo }: { jogo: DashboardGameDoAno }) {
  return <article aria-label="Jogo do Ano atual" className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><p className="text-[10px] font-bold uppercase tracking-wider text-[var(--highlight-gold)]">Jogo do Ano {jogo.ano}</p><Link to={`/biblioteca/${jogo.id}`} className="mt-3 flex gap-3"><Capa jogo={jogo} /><div className="min-w-0"><h2 className="line-clamp-2 text-sm font-bold text-[var(--text-primary)]">{jogo.nome}</h2><p className="mt-1 truncate text-[11px] text-[var(--text-secondary)]"><Gamepad2 className="mr-1 inline size-3" aria-hidden="true" />{jogo.console} · {jogo.tempo_jogado === undefined ? '—' : formatarHoras(jogo.tempo_jogado)}</p><span className="mt-2 inline-flex items-center gap-1 rounded bg-black/80 px-1.5 py-0.5 text-[10px] font-bold text-[var(--highlight-gold)]"><Star className="size-3 fill-current" aria-hidden="true" />{jogo.nota}</span></div></Link></article>
}

function JogosDaVida({ jogos, total }: { jogos: JogoZeradoDTO[]; total: number }) {
  const exibidos = jogos.slice(0, 7)
  return <article aria-label="Jogos da Vida" className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4"><header className="flex items-center gap-2"><span className="flex size-5 items-center justify-center rounded bg-[var(--highlight-gold)] text-[9px] font-bold text-[var(--bg-primary)]">11</span><h2 className="text-base font-bold text-[var(--text-primary)]">Jogos da Vida</h2><span className="text-[11px] text-[var(--text-muted)]">{exibidos.length} de {total}</span><Link to="/hall-da-fama" className="ml-auto text-[11px] font-semibold text-[var(--accent)]">Ver Hall da Fama →</Link></header><div className="mt-4 grid grid-cols-4 justify-start gap-3 sm:grid-cols-7 lg:grid-cols-[repeat(7,minmax(80px,100px))]">{exibidos.map((jogo) => <Link key={jogo.id} to={`/biblioteca/${jogo.id}`} className="group min-w-0"><Capa jogo={jogo} pequena /><p className="mt-1 line-clamp-2 text-[10px] font-semibold text-[var(--text-primary)]" title={jogo.nome}>{jogo.nome}</p><p className="truncate text-[9px] text-[var(--text-muted)]">{jogo.console}</p></Link>)}</div></article>
}

export function EliteDoJogador({ jogoDoAno, jogosDaVida, jogosDaVidaTotal }: EliteDoJogadorProps) {
  if (!jogoDoAno && !jogosDaVida.length) return null
  const temAmbos = Boolean(jogoDoAno && jogosDaVida.length)
  return <section aria-label="Elite do jogador" className={`grid grid-cols-1 gap-4 ${temAmbos ? 'xl:grid-cols-[340px_1fr]' : ''}`}>{jogoDoAno && <JogoDoAno jogo={jogoDoAno} />}{jogosDaVida.length > 0 && <JogosDaVida jogos={jogosDaVida} total={jogosDaVidaTotal} />}</section>
}
