import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { Hammer } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'

const pageTheme = {
  '--bg-primary': '#15161A',
  '--bg-surface': '#1A1B20',
  '--bg-surface-alt': '#1D1F25',
  '--border': '#24262C',
  '--border-subtle': '#2A2C33',
  '--text-primary': '#EDEDED',
  '--text-secondary': '#9A9CA5',
  '--accent': '#4F7CFF',
  backgroundColor: 'var(--bg-primary)',
  color: 'var(--text-primary)',
  fontFamily: 'Inter, sans-serif',
} as CSSProperties

export interface EmConstrucaoPageProps {
  modulo: string
}

export function EmConstrucaoPage({ modulo }: EmConstrucaoPageProps) {
  return (
    <div style={pageTheme} className="min-h-svh">
      <Topbar />

      <main className="mx-auto max-w-lg px-4 py-16 sm:px-6 text-center">
        <div className="rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-8 space-y-4">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] text-[var(--accent)]">
            <Hammer className="h-6 w-6" aria-hidden="true" />
          </div>

          <div className="space-y-1">
            <h1 className="text-[20px] font-bold text-[var(--text-primary)]">{modulo}</h1>
            <p className="text-[13px] text-[var(--text-secondary)]">
              Este módulo faz parte das próximas etapas do roadmap e estará disponível em breve.
            </p>
          </div>

          <Link
            to="/biblioteca"
            className="inline-block rounded-[7px] bg-[var(--accent)] px-4 py-2 text-[13px] font-bold text-[#0E0F12] transition-opacity hover:opacity-90"
          >
            Ir para a Biblioteca
          </Link>
        </div>
      </main>
    </div>
  )
}
