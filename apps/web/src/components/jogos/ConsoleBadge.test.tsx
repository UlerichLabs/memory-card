import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
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

  it('usa cor sólida e texto adequado na variante da Biblioteca', () => {
    render(<ConsoleBadge nome="Atari 2600" variante="solido" />)
    const badge = screen.getByTitle('Atari 2600')
    expect(badge).toHaveStyle({ backgroundColor: '#F28C28', borderColor: '#F28C28', color: '#1A1B20' })
    expect(badge).toHaveClass('text-xs', 'font-semibold')
  })

})
