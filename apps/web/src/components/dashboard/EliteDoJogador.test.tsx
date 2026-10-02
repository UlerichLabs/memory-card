import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { EliteDoJogador } from './EliteDoJogador'

const jogo = { id: 1, usuario_id: 1, nome: 'Hades', console: 'PC', finalizado_em: '2025-01-01', tempo_jogado: 3600, nota: 11, dificuldade: 'A' as const, destaque: true, igdb_capa_url: '' }

describe('EliteDoJogador', () => {
  it('renderiza GOTY, Jogos da Vida e link do Hall', () => {
    render(<MemoryRouter><EliteDoJogador jogoDoAno={{ ...jogo, ano: 2025 }} jogosDaVida={[jogo]} jogosDaVidaTotal={1} /></MemoryRouter>)
    expect(screen.getByText('Jogo do Ano 2025')).toBeInTheDocument()
    expect(screen.getByText('Jogos da Vida')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Ver Hall da Fama/ })).toHaveAttribute('href', '/hall-da-fama')
  })

  it('não renderiza painel quando não há elite', () => {
    render(<MemoryRouter><EliteDoJogador jogoDoAno={null} jogosDaVida={[]} jogosDaVidaTotal={0} /></MemoryRouter>)
    expect(screen.queryByLabelText('Elite do jogador')).not.toBeInTheDocument()
  })

  it('renderiza somente Jogos da Vida em uma coluna larga com o total real', () => {
    render(<MemoryRouter><EliteDoJogador jogoDoAno={null} jogosDaVida={[jogo]} jogosDaVidaTotal={11} /></MemoryRouter>)
    expect(screen.getByText('1 de 11')).toBeInTheDocument()
    expect(screen.getByText('Jogos da Vida')).toBeInTheDocument()
  })

  it('renderiza somente o Jogo do Ano em uma coluna larga', () => {
    render(<MemoryRouter><EliteDoJogador jogoDoAno={{ ...jogo, ano: 2025 }} jogosDaVida={[]} jogosDaVidaTotal={0} /></MemoryRouter>)
    expect(screen.getByText('Jogo do Ano 2025')).toBeInTheDocument()
  })
})
