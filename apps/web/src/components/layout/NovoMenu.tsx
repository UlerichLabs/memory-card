import { useContext, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Ban, ChevronDown, Gamepad2, ListChecks, Play, Trophy } from 'lucide-react'
import { Menu } from '@base-ui/react/menu'
import { JogosContext } from '@/stores/jogosStore'
import { JogandoContext } from '@/stores/jogandoStore'
import { AbandonadosContext } from '@/stores/abandonadosStore'

const acoes = [
  { titulo: 'Iniciar jogo', descricao: 'Comece agora, anota a data.', atalho: 'I', icone: Play, cor: 'text-[var(--accent)]' },
  { titulo: 'Registrar jogo zerado', descricao: 'Já terminei: tempo, nota e review.', atalho: 'R', icone: Trophy, cor: 'text-[var(--highlight-gold)]' },
  { titulo: 'Abandonar jogo', descricao: 'Parei de jogar. Registre o motivo.', atalho: 'A', icone: Ban, cor: 'text-[var(--abandonado-text)]' },
  { titulo: 'Nova lista ou desafio', descricao: 'Organize sua coleção ou crie uma meta.', atalho: 'L', icone: ListChecks, cor: 'text-[var(--accent)]' },
]

export function NovoMenu() {
  const navigate = useNavigate()
  const jogos = useContext(JogosContext)
  const jogando = useContext(JogandoContext)
  const abandonados = useContext(AbandonadosContext)
  const [aberto, setAberto] = useState(false)
  const executar = (titulo: string) => {
    if (titulo === 'Iniciar jogo') jogando?.abrirModalIniciar()
    if (titulo === 'Registrar jogo zerado') jogos?.abrirModalRegistro()
    if (titulo === 'Abandonar jogo') abandonados?.abrirModalCriacao()
    if (titulo === 'Nova lista ou desafio') navigate('/listas?novo=1')
    setAberto(false)
  }
  return (
    <Menu.Root open={aberto} onOpenChange={setAberto}>
      <Menu.Trigger type="button" onClick={() => setAberto(true)} aria-haspopup="menu" aria-label="Novo" className="group inline-flex size-[38px] shrink-0 items-center justify-center gap-2 rounded-full bg-[linear-gradient(135deg,var(--novo-gradient-start),var(--novo-gradient-mid)_55%,var(--novo-gradient-end))] px-1.5 text-sm font-bold text-white shadow-[0_6px_18px_rgba(79,124,255,0.33),inset_0_1px_0_rgba(255,255,255,0.25)] transition-[filter] hover:brightness-110 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)] md:h-9 md:w-auto md:px-2 md:pr-3">
        <span className="flex size-6 items-center justify-center rounded-full bg-white/15"><Gamepad2 className="size-[15px]" strokeWidth={2.2} aria-hidden="true" /></span>
        <span className="hidden md:inline">Novo</span>
        <ChevronDown className="hidden size-3.5 transition-transform group-data-popup-open:rotate-180 md:inline" aria-hidden="true" />
      </Menu.Trigger>
      <Menu.Portal>
        <Menu.Positioner side="bottom" align="end" sideOffset={8} className="z-50">
          <Menu.Popup style={{ backgroundColor: 'var(--bg-surface)', color: 'var(--text-primary)' }} className="w-[340px] rounded-[14px] border border-[var(--border)] p-1.5 shadow-2xl shadow-black/60 outline-none">
            <p className="px-3 pb-1 pt-2 text-[11px] font-semibold uppercase tracking-[0.14em] text-[var(--text-faint)]">Criar</p>
            {acoes.map(({ titulo, descricao, atalho, icone: Icon, cor }) => <Menu.Item key={titulo} onClick={() => executar(titulo)} className="flex cursor-pointer items-center gap-3 rounded-[10px] px-2.5 py-2 outline-none hover:bg-[var(--bg-surface-alt)] focus:bg-[var(--bg-surface-alt)] data-highlighted:bg-[var(--bg-surface-alt)]"><span className="flex size-9 shrink-0 items-center justify-center rounded-[10px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)]"><Icon className={`size-[17px] ${cor}`} aria-hidden="true" /></span><span className="min-w-0 flex-1"><strong className="block text-sm font-semibold text-[var(--text-primary)]">{titulo}</strong><small className="block max-w-[220px] text-xs font-normal leading-snug text-[var(--text-secondary)]">{descricao}</small></span><kbd className="hidden rounded border border-[var(--border-subtle)] bg-[var(--bg-surface)] px-1.5 py-0.5 text-[10px] font-semibold text-[var(--text-muted)] md:inline-block">{atalho}</kbd></Menu.Item>)}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  )
}
