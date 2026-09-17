import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { CadastroPage } from './CadastroPage'

describe('CadastroPage', () => {
  it('renderiza o título, instrução, link para login e formulário', () => {
    render(
      <MemoryRouter>
        <CadastroPage />
      </MemoryRouter>
    )

    expect(screen.getByRole('heading', { name: /criar conta/i, level: 1 })).toBeInTheDocument()
    expect(
      screen.getByText(/cadastre-se para começar a registrar sua biblioteca/i)
    ).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /entrar/i })).toHaveAttribute('href', '/login')
    expect(screen.getByRole('button', { name: /cadastrar/i })).toBeInTheDocument()
  })
})
