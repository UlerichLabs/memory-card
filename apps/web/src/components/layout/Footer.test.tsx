import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Footer } from './Footer'

describe('Footer', () => {
  it('renderiza a marca, autoria e ano atual', () => {
    render(<Footer />)
    expect(screen.getByText('Memory Card')).toBeInTheDocument()
    expect(screen.getByText(/Desenvolvido por/)).toHaveTextContent('UlerichLabs')
    expect(screen.getByText(/Todos os direitos reservados/)).toHaveTextContent(String(new Date().getFullYear()))
  })
})
