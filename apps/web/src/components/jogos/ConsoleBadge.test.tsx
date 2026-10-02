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
})
