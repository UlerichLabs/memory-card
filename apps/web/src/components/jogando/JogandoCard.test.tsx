import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { JogandoCard } from './JogandoCard'

const jogo = { id: 1, nome: 'Hades', igdb_id: null, igdb_capa_url: null, iniciado_em: '2026-09-20T00:00:00Z' }

describe('JogandoCard', () => {
  it('mostra jogo, data e ações', () => {
    render(<JogandoCard jogo={jogo} onZerei={vi.fn()} onAbandonei={vi.fn()} onRemover={vi.fn()} />)
    expect(screen.getByText('Hades')).toBeInTheDocument()
    expect(screen.getByText(/Começou em 20\/09\/2026/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Zerei!' })).toBeInTheDocument()
  })

  it('dispara as ações do card e do menu', () => {
    const onZerei = vi.fn()
    const onAbandonei = vi.fn()
    const onRemover = vi.fn()
    render(<JogandoCard jogo={jogo} onZerei={onZerei} onAbandonei={onAbandonei} onRemover={onRemover} />)
    fireEvent.click(screen.getByRole('button', { name: 'Zerei!' }))
    fireEvent.click(screen.getByRole('button', { name: 'Abandonei' }))
    fireEvent.click(screen.getByRole('button', { name: 'Mais opções' }))
    fireEvent.click(screen.getByRole('menuitem', { name: 'Remover' }))
    expect(onZerei).toHaveBeenCalledOnce()
    expect(onAbandonei).toHaveBeenCalledOnce()
    expect(onRemover).toHaveBeenCalledOnce()
  })
})
