import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { JogandoVazio } from './JogandoVazio'

describe('JogandoVazio', () => {
  it('mostra a faixa e abre o início', () => {
    const onIniciar = vi.fn()
    render(<JogandoVazio onIniciar={onIniciar} />)
    expect(screen.getByText('Nada em andamento. Inicie um jogo para guardar quando você começou, sem precisar lembrar depois.')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Iniciar um jogo →' }))
    expect(onIniciar).toHaveBeenCalledOnce()
  })
})
