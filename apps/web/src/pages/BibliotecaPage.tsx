import { useState, type CSSProperties } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Search, Disc3 } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { useJogosStore } from '@/stores/jogosStore'
import { BibliotecaCard } from '@/components/jogos/BibliotecaCard'
import { ExcluirJogoDialog } from '@/components/jogos/ExcluirJogoDialog'
import type { JogoZeradoDTO } from '@/lib/services/jogosService'

const bibliotecaTheme = {
  '--bg-primary': '#15161A',
  '--bg-surface': '#1A1B20',
  '--bg-surface-alt': '#1D1F25',
  '--border': '#24262C',
  '--border-subtle': '#2A2C33',
  '--text-primary': '#EDEDED',
  '--text-secondary': '#9A9CA5',
  '--text-muted': '#6B6D76',
  '--text-faint': '#52545C',
  '--accent': '#4F7CFF',
  '--highlight-gold': '#E8C15C',
  '--danger': '#E05A4E',
  backgroundColor: 'var(--bg-primary)',
  color: 'var(--text-primary)',
  fontFamily: 'Inter, sans-serif',
} as CSSProperties

export function BibliotecaPage() {
  const navigate = useNavigate()
  const { jogos, excluirJogo } = useJogosStore()
  const [termoBusca, setTermoBusca] = useState('')
  const [jogoParaExcluir, setJogoParaExcluir] = useState<JogoZeradoDTO | null>(null)
  const [isExcluindo, setIsExcluindo] = useState(false)

  const jogosFiltrados = jogos.filter((jogo) => {
    const termo = termoBusca.toLowerCase()
    return (
      jogo.nome.toLowerCase().includes(termo) ||
      jogo.console.toLowerCase().includes(termo) ||
      (jogo.genero && jogo.genero.toLowerCase().includes(termo))
    )
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
        <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-[22px] font-bold text-[var(--text-primary)]">Biblioteca</h1>
              <span className="rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface)] px-2.5 py-0.5 text-xs font-semibold text-[var(--text-secondary)]">
                {jogos.length} {jogos.length === 1 ? 'jogo' : 'jogos'}
              </span>
            </div>
            <p className="text-[13px] text-[var(--text-secondary)]">Seus registros de zeramentos concluídos.</p>
          </div>

          <button
            type="button"
            onClick={() => navigate('/jogos/novo')}
            className="inline-flex items-center justify-center gap-2 rounded-[7px] bg-[var(--accent)] px-4 py-2 text-[13px] font-bold text-[#0E0F12] transition-opacity hover:opacity-90"
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            <span>Registrar jogo</span>
          </button>
        </header>

        {jogos.length > 0 && (
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-faint)]" aria-hidden="true" />
            <input
              type="search"
              value={termoBusca}
              onChange={(e) => setTermoBusca(e.target.value)}
              placeholder="Filtrar por nome, console ou gênero…"
              aria-label="Filtrar jogos na biblioteca"
              className="h-9 w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] pl-9 pr-3 text-[13px] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
            />
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
              onClick={() => navigate('/jogos/novo')}
              className="inline-flex items-center justify-center gap-2 rounded-[7px] bg-[var(--accent)] px-4 py-2 text-[13px] font-bold text-[#0E0F12] transition-opacity hover:opacity-90"
            >
              <Plus className="h-4 w-4" aria-hidden="true" />
              <span>Registrar primeiro jogo</span>
            </button>
          </div>
        ) : jogosFiltrados.length === 0 ? (
          <div className="rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-8 text-center">
            <p className="text-[13px] text-[var(--text-secondary)]">Nenhum jogo encontrado com o termo "{termoBusca}".</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
            {jogosFiltrados.map((jogo) => (
              <BibliotecaCard
                key={jogo.id}
                jogo={jogo}
                onEditar={(j) => navigate(`/jogos/${j.id}/editar`)}
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
