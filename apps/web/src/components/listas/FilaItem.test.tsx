import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FilaItem } from './FilaItem'
import type { ListaItem } from '@/types/listas'

vi.mock('@dnd-kit/sortable', () => ({
  useSortable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: vi.fn(),
    transform: null,
    transition: undefined,
    isDragging: false,
  }),
}))

const mockItem: ListaItem = {
  id: 42,
  igdb_id: 100,
  nome: 'Chrono Trigger',
  console: 'SNES',
  igdb_capa_url: null,
  ano_lancamento: 1995,
  posicao: 1,
  origem: 'item',
  zerado: false,
  jogo_zerado: null,
}

describe('FilaItem', () => {
  it('renderiza Abandonei quando onAbandonei existe e chama o callback ao clicar', async () => {
    const user = userEvent.setup()
    const handleAbandonei = vi.fn()

    render(
      <FilaItem
        item={mockItem}
        posicaoExibicao={1}
        isFirst={false}
        onZerei={vi.fn()}
        onAbandonei={handleAbandonei}
        onRemover={vi.fn()}
      />
    )

    const botoes = screen.getAllByRole('button', { name: 'Abandonei' })
    expect(botoes.length).toBeGreaterThan(0)
    await user.click(botoes[0])

    expect(handleAbandonei).toHaveBeenCalledTimes(1)
    expect(handleAbandonei).toHaveBeenCalledWith(mockItem)
  })

  it('não renderiza botão Abandonei quando onAbandonei não for fornecido', () => {
    render(
      <FilaItem
        item={mockItem}
        posicaoExibicao={1}
        isFirst={false}
        onZerei={vi.fn()}
        onRemover={vi.fn()}
      />
    )

    expect(screen.queryByRole('button', { name: 'Abandonei' })).not.toBeInTheDocument()
  })
})
