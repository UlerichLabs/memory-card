import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { GameDoAnoVazio } from './GameDoAnoVazio'

describe('GameDoAnoVazio', () => {
  it('renderiza o ano, título de sem destaque e total de jogos plural', () => {
    render(<GameDoAnoVazio ano={2025} totalJogos={4} onEscolher={vi.fn()} />)

    expect(screen.getByText('2025')).toBeInTheDocument()
    expect(screen.getByText('Sem Game do Ano')).toBeInTheDocument()
    expect(screen.getByText('4 jogos zerados em 2025')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Escolher Game do Ano' })).toBeInTheDocument()
  })

  it('exibe total de jogos no singular quando houver 1', () => {
    render(<GameDoAnoVazio ano={2022} totalJogos={1} onEscolher={vi.fn()} />)
    expect(screen.getByText('1 jogo zerado em 2022')).toBeInTheDocument()
  })

  it('chama onEscolher ao clicar no botão', async () => {
    const user = userEvent.setup()
    const onEscolher = vi.fn()
    render(<GameDoAnoVazio ano={2025} totalJogos={2} onEscolher={onEscolher} />)

    await user.click(screen.getByRole('button', { name: 'Escolher Game do Ano' }))
    expect(onEscolher).toHaveBeenCalledWith(2025)
  })
})
