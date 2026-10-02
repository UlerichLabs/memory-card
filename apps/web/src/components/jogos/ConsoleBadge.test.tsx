import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ConsoleBadge } from './ConsoleBadge'

describe('ConsoleBadge', () => {
  it('renderiza o nome do console', () => {
    render(<ConsoleBadge nome="PlayStation 5" />)
    expect(screen.getByText('PlayStation 5')).toBeInTheDocument()
  })
})
