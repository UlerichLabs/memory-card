import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { GameDoAnoCard } from './GameDoAnoCard'
import type { JogoZeradoDTO } from '@/types/jogos'

const jogoMock: JogoZeradoDTO = {
  id: 42,
  usuario_id: 1,
  nome: 'Chrono Trigger',
  console: 'Super Nintendo',
  finalizado_em: '2024-05-15T00:00:00Z',
  tempo_jogado: 36000,
  nota: 10,
  dificuldade: 'A',
  destaque: true,
  igdb_capa_url: 'https://images.igdb.com/igdb/image/upload/t_thumb/co123.jpg',
}

describe('GameDoAnoCard', () => {
  it('renderiza ano, capa, nome, plataforma, nota e contagem plural', () => {
    const onTrocar = vi.fn()
    render(
      <MemoryRouter>
        <GameDoAnoCard ano={2024} totalJogos={5} jogo={jogoMock} onTrocar={onTrocar} />
      </MemoryRouter>
    )

    expect(screen.getByText('2024')).toBeInTheDocument()
    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
    expect(screen.getByText('Super Nintendo')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(screen.getByText('5 jogos')).toBeInTheDocument()
    expect(screen.queryByText('Em andamento')).not.toBeInTheDocument()
  })

  it('exibe contagem no singular quando houver apenas 1 jogo', () => {
    render(
      <MemoryRouter>
        <GameDoAnoCard ano={2023} totalJogos={1} jogo={jogoMock} onTrocar={vi.fn()} />
      </MemoryRouter>
    )

    expect(screen.getByText('1 jogo')).toBeInTheDocument()
  })

  it('exibe pill "Em andamento" apenas no ano atual', () => {
    const anoAtual = new Date().getFullYear()
    render(
      <MemoryRouter>
        <GameDoAnoCard ano={anoAtual} totalJogos={2} jogo={jogoMock} onTrocar={vi.fn()} />
      </MemoryRouter>
    )

    expect(screen.getByText('Em andamento')).toBeInTheDocument()
  })

  it('chama onTrocar ao clicar no botão Trocar', async () => {
    const user = userEvent.setup()
    const onTrocar = vi.fn()
    render(
      <MemoryRouter>
        <GameDoAnoCard ano={2024} totalJogos={3} jogo={jogoMock} onTrocar={onTrocar} />
      </MemoryRouter>
    )

    const btnTrocar = screen.getByRole('button', { name: 'Trocar' })
    await user.click(btnTrocar)

    expect(onTrocar).toHaveBeenCalledWith(2024, jogoMock)
  })

  it('link direciona para /biblioteca/:id', () => {
    render(
      <MemoryRouter>
        <GameDoAnoCard ano={2024} totalJogos={3} jogo={jogoMock} onTrocar={vi.fn()} />
      </MemoryRouter>
    )

    const link = screen.getByRole('link', { name: /ver detalhes de chrono trigger/i })
    expect(link).toHaveAttribute('href', '/biblioteca/42')
  })

  it('exibe placeholder com ícone e nome quando o jogo não tem capa', () => {
    render(
      <MemoryRouter>
        <GameDoAnoCard
          ano={2024}
          totalJogos={1}
          jogo={{ ...jogoMock, igdb_capa_url: undefined }}
          onTrocar={vi.fn()}
        />
      </MemoryRouter>
    )

    expect(screen.getByLabelText('Ver detalhes de Chrono Trigger')).toBeInTheDocument()
    expect(screen.getAllByText('Chrono Trigger')).toHaveLength(2)
    expect(screen.queryByRole('img')).not.toBeInTheDocument()
  })
})
