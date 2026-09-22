import type { CSSProperties } from 'react'
import { useNavigate, useParams, Link } from 'react-router-dom'
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

export function EditarJogoPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { jogos, atualizarJogo } = useJogosStore()

  const jogo = jogos.find((item) => String(item.id) === id)

  async function handleAtualizar(payload: SalvarJogoPayload) {
    if (!jogo) return
    await atualizarJogo(jogo.id, payload)
    navigate('/biblioteca')
  }

  return (
    <div style={pageTheme} className="min-h-svh">
      <Topbar />

      <main className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
        {!jogo ? (
          <div className="rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-8 text-center space-y-4">
            <h1 className="text-lg font-bold text-[var(--text-primary)]">Jogo não encontrado</h1>
            <p className="text-[13px] text-[var(--text-secondary)]">
              O registro solicitado não foi localizado na biblioteca atual.
            </p>
            <Link
              to="/biblioteca"
              className="inline-block rounded-[7px] bg-[var(--accent)] px-4 py-2 text-[13px] font-bold text-[#0E0F12]"
            >
              Voltar para a biblioteca
            </Link>
          </div>
        ) : (
          <>
            <header className="space-y-1">
              <p className="text-xs font-semibold uppercase tracking-wider text-[var(--text-secondary)]">Biblioteca</p>
              <h1 className="text-[22px] font-bold text-[var(--text-primary)]">Editar registro</h1>
              <p className="text-[13px] text-[var(--text-secondary)]">
                Atualize as informações do registro de <strong className="text-[var(--text-primary)]">{jogo.nome}</strong>.
              </p>
            </header>

            <GameForm
              initialData={jogo}
              isEditing={true}
              onSubmit={handleAtualizar}
              onCancel={() => navigate('/biblioteca')}
            />
          </>
        )}
      </main>
    </div>
  )
}
