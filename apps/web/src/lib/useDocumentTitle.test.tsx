import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { useDocumentTitle } from './useDocumentTitle'

function Titulo({ valor }: { valor?: string }) {
  useDocumentTitle(valor)
  return null
}

describe('useDocumentTitle', () => {
  it('define o título e restaura o padrão ao desmontar', () => {
    const { unmount, rerender } = render(<Titulo valor="Biblioteca" />)
    expect(document.title).toBe('Biblioteca · Memory Card')
    rerender(<Titulo />)
    expect(document.title).toBe('Memory Card')
    unmount()
    expect(document.title).toBe('Memory Card')
  })
})
