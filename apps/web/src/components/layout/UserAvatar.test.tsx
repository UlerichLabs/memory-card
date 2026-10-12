import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { UserAvatar } from './UserAvatar'

describe('UserAvatar', () => {
  it('renderiza iniciais quando src não é fornecido', () => {
    render(<UserAvatar nome="Lucas Ulerich" />)
    expect(screen.getByText('LU')).toBeInTheDocument()
    expect(screen.queryByRole('img')).not.toBeInTheDocument()
  })

  it('renderiza imagem quando src é fornecido', () => {
    render(<UserAvatar nome="Lucas Ulerich" src="https://example.com/avatar.jpg" />)
    const img = screen.getByRole('img', { name: /foto de perfil de lucas ulerich/i })
    expect(img).toHaveAttribute('src', 'https://example.com/avatar.jpg')
    expect(screen.queryByText('LU')).not.toBeInTheDocument()
  })

  it('faz fallback para iniciais se a imagem falhar ao carregar (onError)', () => {
    render(<UserAvatar nome="Lucas Ulerich" src="https://example.com/broken.jpg" />)
    const img = screen.getByRole('img')
    fireEvent.error(img)
    expect(screen.getByText('LU')).toBeInTheDocument()
    expect(screen.queryByRole('img')).not.toBeInTheDocument()
  })
})
