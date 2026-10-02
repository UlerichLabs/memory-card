import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { BottomNav } from './BottomNav'

describe('BottomNav', () => {
  it('renderiza cinco itens e ativa rotas filhas', () => {
    render(<MemoryRouter initialEntries={['/biblioteca/42']}><BottomNav /></MemoryRouter>)
    expect(screen.getByRole('navigation', { name: 'Navegação inferior' })).toHaveClass('md:hidden')
    expect(screen.getByRole('link', { name: /Biblioteca/ })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('button', { name: 'Mais' })).toBeInTheDocument()
  })

  it('abre a folha Mais com três links e fecha com Escape', async () => {
    const user = userEvent.setup()
    render(<MemoryRouter><BottomNav /></MemoryRouter>)
    await user.click(screen.getByRole('button', { name: 'Mais' }))
    expect(screen.getByRole('link', { name: /Abandonados/ })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Explorador/ })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Meu perfil/ })).toBeInTheDocument()
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(screen.queryByRole('link', { name: /Abandonados/ })).not.toBeInTheDocument()
  })
})
