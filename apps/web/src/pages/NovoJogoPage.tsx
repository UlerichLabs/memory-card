import type { CSSProperties } from 'react'
import { useNavigate } from 'react-router-dom'
import { Topbar } from '@/components/layout/Topbar'
import { GameForm } from '@/components/jogos/GameForm/GameForm'
import { useJogosStore } from '@/stores/jogosStore'
import type { SalvarJogoPayload } from '@/lib/services/jogosService'

const pageTheme = {
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

export function NovoJogoPage() {
  const navigate = useNavigate()
  const { criarJogo } = useJogosStore()

  async function handleSalvar(payload: SalvarJogoPayload) {
    await criarJogo(payload)
    navigate('/biblioteca')
  }

  return (
    <div style={pageTheme} className="min-h-svh">
      <Topbar />

      <main className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
        <header className="space-y-1">
          <p className="text-xs font-semibold uppercase tracking-wider text-[var(--text-secondary)]">Biblioteca</p>
          <h1 className="text-[22px] font-bold text-[var(--text-primary)]">Registrar jogo</h1>
          <p className="text-[13px] text-[var(--text-secondary)]">
            Preencha os dados do jogo que você zerou para registrar em seu histórico.
          </p>
        </header>

        <GameForm onSubmit={handleSalvar} onCancel={() => navigate('/biblioteca')} />
      </main>
    </div>
  )
}
