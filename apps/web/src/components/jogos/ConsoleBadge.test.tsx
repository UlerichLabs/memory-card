import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { faXbox } from '@fortawesome/free-brands-svg-icons'
import { ConsoleBadge } from './ConsoleBadge'

describe('ConsoleBadge', () => {
  it('renderiza o nome do console', () => {
    render(<ConsoleBadge nome="PlayStation 5" />)
    expect(screen.getByText('PlayStation 5')).toBeInTheDocument()
    expect(screen.getByText('PlayStation 5').previousElementSibling?.tagName).toBe('svg')
  })

  it('renderiza ícones distintos para computador e console', () => {
    const { rerender, container } = render(<ConsoleBadge nome="PC" />)
    expect(container.querySelector('svg')).toBeInTheDocument()
    rerender(<ConsoleBadge nome="Sega Genesis" />)
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('renderiza o path real de Xbox e mantém Nintendo universal', () => {
    const { container, rerender } = render(<ConsoleBadge nome="Xbox One" />)
    const path = container.querySelector('path')
    const xboxPath = Array.isArray(faXbox.icon[4]) ? faXbox.icon[4].join(' ') : faXbox.icon[4]
    expect(path).toHaveAttribute('d', xboxPath)

    rerender(<ConsoleBadge nome="Nintendo Switch" />)
    const nintendoPath = container.querySelector('svg')?.innerHTML
    rerender(<ConsoleBadge nome="Game Boy Color" />)
    expect(container.querySelector('svg')?.innerHTML).toBe(nintendoPath)
  })

  it('renderiza sempre no visual sólido com cor sólida e texto adequado', () => {
    render(<ConsoleBadge nome="Atari 2600" />)
    const badge = screen.getByTitle('Atari 2600')
    expect(badge).toHaveStyle({ backgroundColor: '#F28C28', borderColor: '#F28C28', color: '#1A1B20' })
    expect(badge).toHaveClass('text-xs', 'font-semibold')
  })

  it('aplica classes de tamanho sm e md', () => {
    const { rerender } = render(<ConsoleBadge nome="PC" tamanho="sm" />)
    expect(screen.getByTitle('PC')).toHaveClass('h-6', 'px-2')

    rerender(<ConsoleBadge nome="PC" tamanho="md" />)
    expect(screen.getByTitle('PC')).toHaveClass('h-7', 'px-3')
  })

  it('renderiza nome longo truncado com title e max-w-full', () => {
    const nomeLongo = 'Super Nintendo Entertainment System Classic'
    render(<ConsoleBadge nome={nomeLongo} />)
    const badge = screen.getByTitle(nomeLongo)
    expect(badge).toHaveClass('max-w-full')
    expect(screen.getByText(nomeLongo)).toHaveClass('truncate')
  })
})
