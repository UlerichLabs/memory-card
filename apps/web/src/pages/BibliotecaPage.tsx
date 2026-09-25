import { useMemo, useState } from 'react'
import { Search, Disc3, LayoutGrid, List, Plus } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { useJogosStore } from '@/stores/jogosStore'
import { BibliotecaCard } from '@/components/jogos/BibliotecaCard'
import { BibliotecaListItem } from '@/components/jogos/BibliotecaListItem'
import { ExcluirJogoDialog } from '@/components/jogos/ExcluirJogoDialog'
import { CustomSelect } from '@/components/ui/CustomSelect'
import type { JogoZeradoDTO } from '@/lib/services/jogosService'
import { bibliotecaTheme, NOTA_OPCOES } from './Biblioteca.constants'

export function BibliotecaPage() {
  const { jogos, excluirJogo, abrirModalRegistro, abrirModalEdicao } = useJogosStore()
  const [termoBusca, setTermoBusca] = useState('')
  const [filtroConsole, setFiltroConsole] = useState('')
  const [filtroGenero, setFiltroGenero] = useState('')
  const [filtroNota, setFiltroNota] = useState('')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>(() => {
    return (localStorage.getItem('biblioteca_view_mode') as 'grid' | 'list') || 'grid'
  })
  const [jogoParaExcluir, setJogoParaExcluir] = useState<JogoZeradoDTO | null>(null)
  const [isExcluindo, setIsExcluindo] = useState(false)

  const consoleOptions = useMemo(() => [
    { value: '', label: 'Todos os consoles' },
    ...Array.from(new Set(jogos.map((j) => j.console).filter(Boolean))).sort().map((c) => ({ value: c, label: c })),
  ], [jogos])

  const generoOptions = useMemo(() => [
    { value: '', label: 'Todos os gêneros' },
    ...Array.from(new Set(jogos.flatMap((j) => (j.genero ? j.genero.split(',').map((g) => g.trim()) : [])).filter(Boolean))).sort().map((g) => ({ value: g, label: g })),
  ], [jogos])

  const handleViewMode = (mode: 'grid' | 'list') => {
    setViewMode(mode)
    localStorage.setItem('biblioteca_view_mode', mode)
  }

  const jogosFiltrados = jogos.filter((jogo) => {
    const termo = termoBusca.toLowerCase()
    const matchBusca = !termoBusca || jogo.nome.toLowerCase().includes(termo) ||
      jogo.console.toLowerCase().includes(termo) || Boolean(jogo.genero?.toLowerCase().includes(termo))
    const matchConsole = !filtroConsole || jogo.console === filtroConsole
    const matchGenero = !filtroGenero || Boolean(jogo.genero?.toLowerCase().includes(filtroGenero.toLowerCase()))
    const matchNota = !filtroNota || jogo.nota >= Number(filtroNota)
    return matchBusca && matchConsole && matchGenero && matchNota
  })

  async function handleConfirmarExclusao() {
    if (!jogoParaExcluir) return
    setIsExcluindo(true)
    try {
      await excluirJogo(jogoParaExcluir.id)
      setJogoParaExcluir(null)
    } finally {
      setIsExcluindo(false)
    }
  }

  return (
    <div style={bibliotecaTheme} className="min-h-svh">
      <Topbar />

      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
        <header className="flex flex-col gap-1">
          <div className="flex items-center gap-3">
            <h1 className="text-[22px] font-bold text-[var(--text-primary)]">Biblioteca</h1>
            <span className="rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface)] px-2.5 py-0.5 text-xs font-semibold text-[var(--text-secondary)]">
              {jogos.length} {jogos.length === 1 ? 'jogo' : 'jogos'}
            </span>
          </div>
          <p className="text-[13px] text-[var(--text-secondary)]">Seus registros de zeramentos concluídos.</p>
        </header>

        {jogos.length > 0 && (
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="flex flex-1 flex-wrap items-center gap-2.5">
              <div className="relative min-w-[200px] flex-1 sm:max-w-xs">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-faint)]" aria-hidden="true" />
                <input
                  type="search"
                  value={termoBusca}
                  onChange={(e) => setTermoBusca(e.target.value)}
                  placeholder="Buscar por nome, console ou gênero…"
                  aria-label="Filtrar jogos na biblioteca"
                  className="h-8 w-full rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] pl-9 pr-3 text-xs text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
                />
              </div>
              <div className="w-40">
                <CustomSelect value={filtroConsole} onChange={setFiltroConsole} options={consoleOptions} placeholder="Console" ariaLabel="Filtrar por console" />
              </div>
              <div className="w-40">
                <CustomSelect value={filtroGenero} onChange={setFiltroGenero} options={generoOptions} placeholder="Gênero" ariaLabel="Filtrar por gênero" />
              </div>
              <div className="w-32">
                <CustomSelect value={filtroNota} onChange={setFiltroNota} options={NOTA_OPCOES} placeholder="Nota" ariaLabel="Filtrar por nota mínima" />
              </div>
            </div>

            <div className="flex items-center gap-1 self-end md:self-center rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-1">
              <button
                type="button"
                onClick={() => handleViewMode('grid')}
                aria-label="Visualização em grade"
                className={`p-1.5 rounded-md transition-colors ${
                  viewMode === 'grid'
                    ? 'bg-[var(--bg-surface-alt)] text-[var(--text-primary)] shadow-sm'
                    : 'text-[var(--text-muted)] hover:text-[var(--text-secondary)]'
                }`}
              >
                <LayoutGrid className="h-4 w-4" aria-hidden="true" />
              </button>
              <button
                type="button"
                onClick={() => handleViewMode('list')}
                aria-label="Visualização em lista"
                className={`p-1.5 rounded-md transition-colors ${
                  viewMode === 'list'
                    ? 'bg-[var(--bg-surface-alt)] text-[var(--text-primary)] shadow-sm'
                    : 'text-[var(--text-muted)] hover:text-[var(--text-secondary)]'
                }`}
              >
                <List className="h-4 w-4" aria-hidden="true" />
              </button>
            </div>
          </div>
        )}

        {jogos.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-12 text-center space-y-4">
            <div className="flex h-14 w-14 items-center justify-center rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] text-[var(--text-secondary)]">
              <Disc3 className="h-7 w-7 text-[var(--accent)]" aria-hidden="true" />
            </div>
            <div className="space-y-1">
              <h2 className="text-base font-bold text-[var(--text-primary)]">Nenhum jogo registrado ainda</h2>
              <p className="text-[13px] text-[var(--text-secondary)] max-w-sm">
                Sua biblioteca está vazia. Comece a registrar seus jogos zerados com nota, tempo jogado e dificuldade.
              </p>
            </div>
            <button
              type="button"
              onClick={abrirModalRegistro}
              className="inline-flex items-center justify-center gap-2 rounded-[7px] bg-[var(--accent)] px-4 py-2 text-[13px] font-bold text-[#0E0F12] transition-opacity hover:opacity-90"
            >
              <Plus className="h-4 w-4" aria-hidden="true" />
              <span>Registrar primeiro jogo</span>
            </button>
          </div>
        ) : jogosFiltrados.length === 0 ? (
          <div className="rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-8 text-center">
            <p className="text-[13px] text-[var(--text-secondary)]">Nenhum jogo encontrado com os filtros selecionados.</p>
          </div>
        ) : viewMode === 'grid' ? (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
            {jogosFiltrados.map((jogo) => (
              <BibliotecaCard
                key={jogo.id}
                jogo={jogo}
                onEditar={(j) => abrirModalEdicao(j)}
                onExcluir={(j) => setJogoParaExcluir(j)}
              />
            ))}
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {jogosFiltrados.map((jogo) => (
              <BibliotecaListItem
                key={jogo.id}
                jogo={jogo}
                onEditar={(j) => abrirModalEdicao(j)}
                onExcluir={(j) => setJogoParaExcluir(j)}
              />
            ))}
          </div>
        )}
      </main>

      <ExcluirJogoDialog
        open={!!jogoParaExcluir}
        onOpenChange={(open) => !open && setJogoParaExcluir(null)}
        jogoNome={jogoParaExcluir?.nome}
        isLoading={isExcluindo}
        onConfirm={handleConfirmarExclusao}
      />
    </div>
  )
}
