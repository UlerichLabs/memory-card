import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { AbandonadoCard } from './AbandonadoCard'
import type { JogoAbandonado } from '@/types/abandonados'

const mockJogo: JogoAbandonado = {
  id: 1,
  nome: 'Final Fantasy VII',
  console: 'PlayStation',
  igdb_id: 42,
  igdb_capa_url: 'https://images.igdb.com/cover.jpg',
  tempo_jogado: 7200,
  motivo: 'Dificuldade excessiva no chefe final',
  iniciado_em: '2026-01-01',
  abandonado_em: '2026-02-15',
  created_at: '2026-02-15T12:00:00Z',
  updated_at: '2026-02-15T12:00:00Z',
}

describe('AbandonadoCard', () => {
  it('exibe o rótulo "Retomar" completo e sem classe truncate', () => {
    render(
      <AbandonadoCard
        jogo={mockJogo}
        onRetomar={vi.fn()}
        onEditar={vi.fn()}
        onExcluir={vi.fn()}
      />
    )

    const btnRetomar = screen.getByRole('button', { name: `Retomar ${mockJogo.nome}` })
    expect(btnRetomar).toBeInTheDocument()
    expect(btnRetomar).toHaveClass('w-full', 'min-h-[44px]')
    const spanRetomar = screen.getByText('Retomar')
    expect(spanRetomar).toBeInTheDocument()
    expect(spanRetomar).not.toHaveClass('truncate')
  })

  it('exibe botões Editar e Excluir com alvo de toque adequado e rótulos visíveis', () => {
    render(
      <AbandonadoCard
        jogo={mockJogo}
        onRetomar={vi.fn()}
        onEditar={vi.fn()}
        onExcluir={vi.fn()}
      />
    )

    const btnEditar = screen.getByRole('button', { name: `Editar ${mockJogo.nome}` })
    const btnExcluir = screen.getByRole('button', { name: `Excluir ${mockJogo.nome}` })

    expect(btnEditar).toHaveClass('min-h-[44px]')
    expect(btnExcluir).toHaveClass('min-h-[44px]')
    expect(screen.getByText('Editar')).toBeInTheDocument()
    expect(screen.getByText('Excluir')).toBeInTheDocument()
  })

  it('chama os callbacks corretos ao clicar nos botões de ação', async () => {
    const user = userEvent.setup()
    const handleRetomar = vi.fn()
    const handleEditar = vi.fn()
    const handleExcluir = vi.fn()

    render(
      <AbandonadoCard
        jogo={mockJogo}
        onRetomar={handleRetomar}
        onEditar={handleEditar}
        onExcluir={handleExcluir}
      />
    )

    await user.click(screen.getByRole('button', { name: `Retomar ${mockJogo.nome}` }))
    expect(handleRetomar).toHaveBeenCalledWith(mockJogo)

    await user.click(screen.getByRole('button', { name: `Editar ${mockJogo.nome}` }))
    expect(handleEditar).toHaveBeenCalledWith(mockJogo)

    await user.click(screen.getByRole('button', { name: `Excluir ${mockJogo.nome}` }))
    expect(handleExcluir).toHaveBeenCalledWith(mockJogo)
  })

  it('renderiza data de abandono sem truncate com line-clamp-2 e altura mínima', () => {
    render(
      <AbandonadoCard
        jogo={mockJogo}
        onRetomar={vi.fn()}
        onEditar={vi.fn()}
        onExcluir={vi.fn()}
      />
    )

    const textoData = screen.getByText('Abandonado em 15/02/2026')
    expect(textoData).toBeInTheDocument()
    expect(textoData).not.toHaveClass('truncate')
    expect(textoData).toHaveClass('line-clamp-2', 'min-h-[32px]')
  })

  it('renderiza motivo do abandono com line-clamp-2 e title com o texto completo', () => {
    render(
      <AbandonadoCard
        jogo={mockJogo}
        onRetomar={vi.fn()}
        onEditar={vi.fn()}
        onExcluir={vi.fn()}
      />
    )

    const textoMotivo = screen.getByText(`"${mockJogo.motivo}"`)
    expect(textoMotivo).toHaveClass('line-clamp-2', 'min-h-[32px]')
    expect(textoMotivo).toHaveAttribute('title', mockJogo.motivo!)
  })

  it('renderiza ConsoleBadge sólido com nome do console e selo de tempo sobre a capa', () => {
    render(
      <AbandonadoCard
        jogo={mockJogo}
        onRetomar={vi.fn()}
        onEditar={vi.fn()}
        onExcluir={vi.fn()}
      />
    )

    expect(screen.getByTitle('PlayStation')).toBeInTheDocument()
    expect(screen.getByText('2h')).toBeInTheDocument()
  })
})
