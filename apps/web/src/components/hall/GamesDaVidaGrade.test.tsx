import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { GamesDaVidaGrade } from './GamesDaVidaGrade'
import type { JogoZeradoDTO } from '@/types/jogos'

const jogosNota11Mock: JogoZeradoDTO[] = [
  {
    id: 1,
    usuario_id: 1,
    nome: 'The Legend of Zelda: Ocarina of Time',
    console: 'Nintendo 64',
    finalizado_em: '2024-05-15T00:00:00Z',
    tempo_jogado: 120000,
    nota: 11,
    dificuldade: 'AA',
    destaque: true,
  },
  {
    id: 2,
    usuario_id: 1,
    nome: 'Super Mario World',
    console: 'SNES',
    finalizado_em: '2023-11-20T00:00:00Z',
    tempo_jogado: 30000,
    nota: 11,
    dificuldade: 'A',
    destaque: false,
  },
]

describe('GamesDaVidaGrade', () => {
  it('renderiza o cabeçalho, contagem e os cards de jogos', () => {
    render(
      <MemoryRouter>
        <GamesDaVidaGrade jogos={jogosNota11Mock} />
      </MemoryRouter>
    )

    expect(screen.getByText('Games da Vida')).toBeInTheDocument()
    expect(screen.getByText('2 jogos')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ver na Biblioteca →' })).toHaveAttribute(
      'href',
      '/biblioteca?nota_min=11'
    )
    expect(
      screen.getByRole('heading', { name: 'The Legend of Zelda: Ocarina of Time' })
    ).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: 'Super Mario World' })
    ).toBeInTheDocument()
  })

  it('exibe o selo "Jogo do ano AAAA" apenas nos jogos com destaque', () => {
    render(
      <MemoryRouter>
        <GamesDaVidaGrade jogos={jogosNota11Mock} />
      </MemoryRouter>
    )

    expect(screen.getByText('Jogo do ano 2024')).toBeInTheDocument()
    expect(screen.queryByText('Jogo do ano 2023')).not.toBeInTheDocument()
  })

  it('exibe mensagem quando não houver jogos com nota 11', () => {
    render(
      <MemoryRouter>
        <GamesDaVidaGrade jogos={[]} />
      </MemoryRouter>
    )

    expect(screen.getByText('Nenhum Game da Vida ainda')).toBeInTheDocument()
    expect(
      screen.getByText('Quando um jogo marcar a sua vida, dê nota 11 a ele.')
    ).toBeInTheDocument()
  })
})
