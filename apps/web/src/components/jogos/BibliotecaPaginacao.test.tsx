import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BibliotecaPaginacao } from './BibliotecaPaginacao'

describe('BibliotecaPaginacao', () => {
  it('exibe todos os controles desabilitados quando há apenas uma página', async () => {
    const user = userEvent.setup()
    const onMudarPagina = vi.fn()
    render(
      <BibliotecaPaginacao
        meta={{ pagina: 1, por_pagina: 100, total: 10, total_paginas: 1 }}
        onMudarPagina={onMudarPagina}
      />
    )
    expect(screen.getByRole('navigation')).toHaveTextContent('Mostrando 1–10 de 10 jogos')
    expect(screen.getByRole('button', { name: 'Página 1' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('button', { name: 'Página anterior' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Página anterior' })).toHaveAttribute('aria-disabled', 'true')
    expect(screen.getByRole('button', { name: 'Próxima página' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Próxima página' })).toHaveAttribute('aria-disabled', 'true')
    await user.click(screen.getByRole('button', { name: 'Página 1' }))
    await user.click(screen.getByRole('button', { name: 'Página anterior' }))
    await user.click(screen.getByRole('button', { name: 'Próxima página' }))
    expect(onMudarPagina).not.toHaveBeenCalled()
  })

  it('desabilita botão Anterior na primeira página e Próxima na última página', () => {
    const onMudarPagina = vi.fn()
    const { rerender } = render(
      <BibliotecaPaginacao
        meta={{ pagina: 1, por_pagina: 24, total: 48, total_paginas: 2 }}
        onMudarPagina={onMudarPagina}
      />
    )

    expect(screen.getByRole('button', { name: 'Página anterior' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Próxima página' })).toBeEnabled()

    rerender(
      <BibliotecaPaginacao
        meta={{ pagina: 2, por_pagina: 24, total: 48, total_paginas: 2 }}
        onMudarPagina={onMudarPagina}
      />
    )

    expect(screen.getByRole('button', { name: 'Página anterior' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Próxima página' })).toBeDisabled()
  })

  it('navega ao clicar em página numérica, Anterior e Próxima', async () => {
    const user = userEvent.setup()
    const onMudarPagina = vi.fn()
    render(
      <BibliotecaPaginacao
        meta={{ pagina: 2, por_pagina: 24, total: 72, total_paginas: 3 }}
        onMudarPagina={onMudarPagina}
      />
    )

    await user.click(screen.getByRole('button', { name: 'Página 1' }))
    expect(onMudarPagina).toHaveBeenCalledWith(1)

    await user.click(screen.getByRole('button', { name: 'Página anterior' }))
    expect(onMudarPagina).toHaveBeenCalledWith(1)

    await user.click(screen.getByRole('button', { name: 'Próxima página' }))
    expect(onMudarPagina).toHaveBeenCalledWith(3)
  })

  it('exibe reticências adequadas quando há muitas páginas', () => {
    const onMudarPagina = vi.fn()
    const { rerender } = render(
      <BibliotecaPaginacao
        meta={{ pagina: 1, por_pagina: 10, total: 100, total_paginas: 10 }}
        onMudarPagina={onMudarPagina}
      />
    )
    expect(screen.getByText('...')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Página 5' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Página 10' })).toBeInTheDocument()

    rerender(
      <BibliotecaPaginacao
        meta={{ pagina: 6, por_pagina: 10, total: 100, total_paginas: 10 }}
        onMudarPagina={onMudarPagina}
      />
    )
    expect(screen.getAllByText('...')).toHaveLength(2)
    expect(screen.getByRole('button', { name: 'Página 1' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Página 6' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('button', { name: 'Página 10' })).toBeInTheDocument()

    rerender(
      <BibliotecaPaginacao
        meta={{ pagina: 10, por_pagina: 10, total: 100, total_paginas: 10 }}
        onMudarPagina={onMudarPagina}
      />
    )
    expect(screen.getByText('...')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Página 6' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Página 10' })).toHaveAttribute('aria-current', 'page')
  })
})
