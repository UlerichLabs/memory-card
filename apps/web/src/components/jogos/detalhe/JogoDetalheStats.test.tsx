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
  it('renderiza exatamente dois badges de nota envolvidos pelos wrappers de visibilidade mobile e desktop', () => {
    render(<JogoDetalheStats jogo={jogoMock} />)

    const badges = screen.getAllByLabelText('Nota 10')
    expect(badges).toHaveLength(2)

    const wrapperMobile = badges[0].parentElement
    const wrapperDesktop = badges[1].parentElement

    expect(wrapperMobile).toHaveClass('sm:hidden')
    expect(badges[0]).not.toHaveClass('sm:hidden')

    expect(wrapperDesktop).toHaveClass('hidden', 'sm:inline-flex')
    expect(badges[1]).not.toHaveClass('hidden')
  })

  it('renderiza corretamente nota 11 com os mesmos wrappers responsivos e efeito de pulso', () => {
    const jogoNota11: JogoZeradoDTO = { ...jogoMock, nota: 11 }
    render(<JogoDetalheStats jogo={jogoNota11} />)

    const badges = screen.getAllByLabelText('Nota 11')
    expect(badges).toHaveLength(2)

    expect(badges[0].parentElement).toHaveClass('sm:hidden')
    expect(badges[1].parentElement).toHaveClass('hidden', 'sm:inline-flex')
    expect(badges[0]).toHaveClass('animate-pulso-ouro')
    expect(badges[1]).toHaveClass('animate-pulso-ouro')
    expect(screen.getByText('Jogo da Vida')).toBeInTheDocument()
  })
})
