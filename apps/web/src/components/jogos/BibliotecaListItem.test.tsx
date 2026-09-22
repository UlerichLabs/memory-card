import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BibliotecaListItem } from './BibliotecaListItem'
import type { JogoZeradoDTO } from '@/lib/services/jogosService'

const mockJogo: JogoZeradoDTO = {
  id: 1,
  usuario_id: 1,
  nome: 'Chrono Trigger',
  console: 'SNES',
  genero: 'JRPG',
  tipo: 'Campanha',
  iniciado_em: '2026-01-01',
  finalizado_em: '2026-01-15',
  tempo_jogado: 7200,
  nota: 10,
  dificuldade: 'A',
  condicao_zeramento: '100%',
  destaque: true,
  igdb_capa_url: 'https://images.igdb.com/cover.jpg',
}

describe('BibliotecaListItem', () => {
  it('renderiza os dados do jogo em linha horizontal', () => {
    render(<BibliotecaListItem jogo={mockJogo} onEditar={vi.fn()} onExcluir={vi.fn()} />)

    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
    expect(screen.getByText('Destaque')).toBeInTheDocument()
    expect(screen.getByText('JRPG')).toBeInTheDocument()
    expect(screen.getByText('SNES')).toBeInTheDocument()
    expect(screen.getByText('Nota 10')).toBeInTheDocument()
    expect(screen.getByText('2h 0m')).toBeInTheDocument()
    expect(screen.getByText('15/01/2026')).toBeInTheDocument()
  })

  it('abre menu de opções e chama onEditar e onExcluir', async () => {
    const user = userEvent.setup()
    const handleEditar = vi.fn()
    const handleExcluir = vi.fn()

    render(<BibliotecaListItem jogo={mockJogo} onEditar={handleEditar} onExcluir={handleExcluir} />)

    const btnMenu = screen.getByRole('button', { name: 'Opções de Chrono Trigger' })
    await user.click(btnMenu)

    const btnEditar = screen.getByRole('button', { name: 'Editar' })
    await user.click(btnEditar)
    expect(handleEditar).toHaveBeenCalledWith(mockJogo)

    await user.click(btnMenu)
    const btnExcluir = screen.getByRole('button', { name: 'Excluir' })
    await user.click(btnExcluir)
    expect(handleExcluir).toHaveBeenCalledWith(mockJogo)
  })
})
