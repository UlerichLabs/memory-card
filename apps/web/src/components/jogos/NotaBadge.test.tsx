import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { NotaBadge } from './NotaBadge'

describe('NotaBadge', () => {
  it('renderiza apenas o número da nota sem estrela', () => {
    const { container } = render(<NotaBadge nota={10} tamanho="sm" />)
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(container.querySelector('svg')).toBeNull()
    expect(container.textContent).toBe('10')
  })

  it('aplica cores de fundo corretas para 1, 5, 10 e 11', () => {
    const { rerender } = render(<NotaBadge nota={1} tamanho="sm" />)
    let badge = screen.getByText('1')
    expect(badge.style.backgroundColor).toBe('var(--nota-1-fill)')

    rerender(<NotaBadge nota={5} tamanho="sm" />)
    badge = screen.getByText('5')
    expect(badge.style.backgroundColor).toBe('var(--nota-5-fill)')

    rerender(<NotaBadge nota={10} tamanho="sm" />)
    badge = screen.getByText('10')
    expect(badge.style.backgroundColor).toBe('var(--nota-10-fill)')

    rerender(<NotaBadge nota={11} tamanho="sm" />)
    badge = screen.getByText('11')
    expect(badge.style.backgroundColor).toBe('var(--nota-11-fill)')
  })

  it('aplica animação pulsoOuro somente na nota 11 quando pulsar é true', () => {
    const { rerender } = render(<NotaBadge nota={11} tamanho="sm" pulsar />)
    let badge = screen.getByText('11')
    expect(badge.className).toContain('animate-pulso-ouro')

    rerender(<NotaBadge nota={11} tamanho="sm" pulsar={false} />)
    badge = screen.getByText('11')
    expect(badge.className).not.toContain('animate-pulso-ouro')
    expect(badge.style.boxShadow).toBe('0 0 12px 2px rgba(240, 190, 80, 0.45)')

    rerender(<NotaBadge nota={10} tamanho="sm" pulsar />)
    badge = screen.getByText('10')
    expect(badge.className).not.toContain('animate-pulso-ouro')
  })

  it('aplica as classes de tamanho sm, md e lg', () => {
    const { rerender } = render(<NotaBadge nota={8} tamanho="sm" />)
    let badge = screen.getByText('8')
    expect(badge.className).toContain('min-w-[32px]')

    rerender(<NotaBadge nota={8} tamanho="md" />)
    badge = screen.getByText('8')
    expect(badge.className).toContain('w-[44px]')

    rerender(<NotaBadge nota={8} tamanho="lg" />)
    badge = screen.getByText('8')
    expect(badge.className).toContain('w-[56px]')
    expect(badge.style.borderColor).toBe('var(--nota-8-fill-border)')
  })
})
