import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { JogoDetalheStats } from './JogoDetalheStats'
import type { JogoZeradoDTO } from '@/types/jogos'

const jogoMock: JogoZeradoDTO = {
  id: 1,
  usuario_id: 1,
  nome: 'Chrono Trigger',
  console: 'SNES',
  finalizado_em: '2026-03-01T00:00:00Z',
  tempo_jogado: 7200,
  nota: 10,
  dificuldade: 'A',
  destaque: false,
}

describe('JogoDetalheStats', () => {
  it('renderiza exatamente um badge com label da nota', () => {
    render(<JogoDetalheStats jogo={jogoMock} />)

    const badges = screen.getAllByLabelText('Nota 10')
    expect(badges).toHaveLength(1)
    expect(badges[0]).toBeInTheDocument()
    expect(screen.getByText('Incrível')).toBeInTheDocument()
  })

  it('renderiza exatamente um badge para nota 11 com o rotulo Jogo da Vida', () => {
    const jogoNota11: JogoZeradoDTO = { ...jogoMock, nota: 11 }
    render(<JogoDetalheStats jogo={jogoNota11} />)

    const badges = screen.getAllByLabelText('Nota 11')
    expect(badges).toHaveLength(1)
    expect(badges[0]).toBeInTheDocument()
    expect(screen.getByText('Jogo da Vida')).toBeInTheDocument()
  })
})
