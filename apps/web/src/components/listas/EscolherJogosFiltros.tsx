import { useEffect, useRef, useState } from 'react'
import { Search } from 'lucide-react'
import { listasService } from '@/lib/services/listasService'
import type { OpcaoOrigem, OrdenarCatalogo, OrigemTipo } from '@/types/listas'

interface EscolherJogosFiltrosProps { origem: OrigemTipo; busca: string; generoId?: number; plataformaId?: number; ordenar: OrdenarCatalogo; token?: string; onBusca: (valor: string) => void; onGenero: (valor?: number) => void; onPlataforma: (valor?: number) => void; onOrdenar: (valor: OrdenarCatalogo) => void }

export function EscolherJogosFiltros({ origem, busca, generoId, plataformaId, ordenar, token, onBusca, onGenero, onPlataforma, onOrdenar }: EscolherJogosFiltrosProps) {
  const [texto, setTexto] = useState(busca); const [generos, setGeneros] = useState<OpcaoOrigem[]>([]); const [plataformas, setPlataformas] = useState<OpcaoOrigem[]>([]); const timer = useRef<ReturnType<typeof setTimeout> | null>(null)
  useEffect(() => { setTexto(busca) }, [busca])
  useEffect(() => { if (origem === 'plataforma') void listasService.listarGeneros(token).then(setGeneros); if (origem === 'genero') void listasService.listarPlataformas(token).then(setPlataformas) }, [origem, token])
  const mudarBusca = (valor: string) => { setTexto(valor); if (timer.current) clearTimeout(timer.current); timer.current = setTimeout(() => onBusca(valor), 400) }
  return <div className="flex flex-col gap-3 md:flex-row md:items-center">
    <label className="relative min-w-0 flex-1"><span className="sr-only">Filtrar por nome</span><Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--lista-text-muted)]" /><input value={texto} onChange={(event) => mudarBusca(event.target.value)} placeholder="Filtrar por nome" className="h-[42px] w-full rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] pl-9 pr-3 text-[13px] text-[var(--text-primary)]" /></label>
    {origem === 'plataforma' && <select value={generoId ?? ''} onChange={(event) => onGenero(event.target.value ? Number(event.target.value) : undefined)} className="h-[42px] w-full rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] px-3 text-[13px] text-[var(--text-primary)] md:w-[190px]"><option value="">Todos os gêneros</option>{generos.map((item) => <option key={item.id} value={item.id}>{item.nome}</option>)}</select>}
    {origem === 'genero' && <select value={plataformaId ?? ''} onChange={(event) => onPlataforma(event.target.value ? Number(event.target.value) : undefined)} className="h-[42px] w-full rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] px-3 text-[13px] text-[var(--text-primary)] md:w-[190px]"><option value="">Todas as plataformas</option>{plataformas.map((item) => <option key={item.id} value={item.id}>{item.nome}</option>)}</select>}
    <label className="flex items-center gap-2 text-[13px] text-[var(--lista-text-secondary)]"><span className="sr-only">Ordenar</span><select value={ordenar} onChange={(event) => onOrdenar(event.target.value as OrdenarCatalogo)} className="h-[42px] w-full rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] px-3 text-[13px] text-[var(--text-primary)] md:w-[170px]"><option value="populares">Mais populares</option><option value="lancamento">Lançamento</option><option value="nome">Nome</option></select></label>
  </div>
}
