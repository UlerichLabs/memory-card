import { Link } from 'react-router-dom'
import { Crown, Gamepad2, Sparkles, Star, Trophy } from 'lucide-react'
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
  return <div className={`relative overflow-hidden rounded-[7px] border border-[var(--highlight-gold)] bg-[var(--bg-surface-alt)] shadow-[0_0_10px_1px_var(--highlight-gold)] ${pequena ? 'aspect-[3/4] w-full' : 'h-[150px] w-28 border-2 sm:h-[187px] sm:w-[140px]'}`}>{url ? <img src={url} alt={jogo.nome} className="h-full w-full object-cover" /> : <div className="flex h-full items-center justify-center text-xs font-bold text-[var(--text-muted)]">{iniciais(jogo.nome)}</div>}{pequena && <span className="absolute left-1.5 top-1.5 flex size-[26px] items-center justify-center rounded-full bg-[var(--highlight-gold)] text-[10px] font-bold text-[var(--bg-primary)]">11</span>}</div>
}

function Estrelas({ nota }: { nota: number }) {
  const preenchidas = Math.min(5, Math.max(1, Math.round(nota / 2)))
  return <div aria-label={`${preenchidas} de 5 estrelas`} className="flex gap-0.5">{Array.from({ length: 5 }, (_, index) => <Star key={index} className={`size-4 ${index < preenchidas ? 'fill-current text-[var(--highlight-gold)]' : 'text-[var(--text-faint)]'}`} aria-hidden="true" />)}</div>
}

function JogoDoAno({ jogo }: { jogo: DashboardGameDoAno }) {
  const horas = jogo.tempo_jogado === undefined ? '—' : formatarHoras(jogo.tempo_jogado)
  return <article aria-label="Jogo do Ano atual" className="relative isolate overflow-hidden rounded-[12px] border border-[var(--highlight-gold)]/40 bg-[radial-gradient(circle_at_top_right,var(--highlight-gold)_0%,var(--bg-surface)_55%)] p-4 sm:p-5"><Trophy className="pointer-events-none absolute -bottom-8 -right-8 z-[-1] size-[190px] text-[var(--highlight-gold)] opacity-[.13]" aria-hidden="true" /><div className="pointer-events-none absolute right-5 top-4 flex gap-1 text-[var(--highlight-gold)] opacity-70" aria-hidden="true"><Sparkles className="size-4" /><Sparkles className="mt-3 size-3" /><Sparkles className="size-2" /></div><div className="relative"><p className="inline-flex items-center gap-1.5 rounded-full border border-[var(--highlight-gold)]/50 bg-[var(--bg-surface-alt)]/70 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-[var(--highlight-gold)]"><Trophy className="size-3.5" aria-hidden="true" />Jogo do Ano {jogo.ano}</p><Link to={`/biblioteca/${jogo.id}`} className="mt-4 flex gap-4"><div className="relative shrink-0"><Capa jogo={jogo} /><span className="absolute -left-3 -top-3 flex size-9 items-center justify-center rounded-full bg-[var(--highlight-gold)] text-[var(--bg-primary)] shadow-[0_0_10px_1px_var(--highlight-gold)]"><Crown className="size-5 fill-current" aria-hidden="true" /></span></div><div className="min-w-0 pt-1"><h2 className="line-clamp-2 text-[22px] font-bold leading-tight text-[var(--text-primary)]">{jogo.nome}</h2><p className="mt-2 truncate text-xs text-[var(--text-secondary)]"><Gamepad2 className="mr-1 inline size-3.5" aria-hidden="true" />{jogo.console} · {horas} jogadas</p><div className="mt-3 flex items-center gap-3"><Estrelas nota={jogo.nota} /><span className="inline-flex items-center gap-1 rounded-full bg-[var(--highlight-gold)] px-2 py-1 text-[10px] font-bold text-[var(--bg-primary)]"><Star className="size-3 fill-current" aria-hidden="true" />Nota {jogo.nota}</span></div></div></Link></div></article>
}

function JogosDaVida({ jogos, total }: { jogos: JogoZeradoDTO[]; total: number }) {
  const exibidos = jogos.slice(0, 5)
  return <article aria-label="Jogos da Vida" className="rounded-[10px] border border-[var(--border)] bg-[var(--bg-surface)] p-4 sm:p-5"><header className="flex items-center gap-2"><span className="flex size-6 items-center justify-center rounded-full bg-[var(--highlight-gold)] text-[10px] font-bold text-[var(--bg-primary)]">11</span><h2 className="text-base font-bold text-[var(--text-primary)]">Jogos da Vida</h2><span className="text-[11px] text-[var(--text-muted)]">{exibidos.length} de {total}</span><Link to="/hall-da-fama" className="ml-auto text-[11px] font-semibold text-[var(--accent)]">Ver Hall da Fama →</Link></header><div className="mt-4 flex snap-x gap-3 overflow-x-auto pb-2 sm:grid sm:grid-cols-5 sm:overflow-visible">{exibidos.map((jogo) => <Link key={jogo.id} to={`/biblioteca/${jogo.id}`} className="group w-[118px] shrink-0 snap-start sm:w-auto"><Capa jogo={jogo} pequena /><p className="mt-2 truncate text-[13px] font-semibold text-[var(--text-primary)]" title={jogo.nome}>{jogo.nome}</p><p className="truncate text-[11px] text-[var(--text-muted)]">{jogo.console}</p></Link>)}</div></article>
}

export function EliteDoJogador({ jogoDoAno, jogosDaVida, jogosDaVidaTotal }: EliteDoJogadorProps) {
  if (!jogoDoAno && !jogosDaVida.length) return null
  const temAmbos = Boolean(jogoDoAno && jogosDaVida.length)
  return <section aria-label="Elite do jogador" className={`grid grid-cols-1 gap-4 ${temAmbos ? 'xl:grid-cols-[380px_1fr]' : ''}`}>{jogoDoAno && <JogoDoAno jogo={jogoDoAno} />}{jogosDaVida.length > 0 && <JogosDaVida jogos={jogosDaVida} total={jogosDaVidaTotal} />}</section>
}
