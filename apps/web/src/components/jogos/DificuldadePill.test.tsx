import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { DificuldadePill } from './DificuldadePill'
import type { Dificuldade } from '@/types/jogos'

describe('DificuldadePill', () => {
  const casos: Array<{ nivel: Dificuldade; label: string; code: string }> = [
    { nivel: 'C', label: 'Muito fácil', code: 'c' },
    { nivel: 'B', label: 'Fácil', code: 'b' },
    { nivel: 'A', label: 'Normal', code: 'a' },
    { nivel: 'AA', label: 'Difícil', code: 'aa' },
    { nivel: 'AAA', label: 'Muito difícil', code: 'aaa' },
  ]

  it('exibe o nome por extenso e nunca a letra isolada para todos os níveis', () => {
    casos.forEach(({ nivel, label }) => {
      const { unmount } = render(<DificuldadePill nivel={nivel} variante="neutra" />)
      expect(screen.getByText(label)).toBeInTheDocument()
      unmount()
    })
  })

  it('aplica cores corretas para a variante sobreCapa', () => {
    casos.forEach(({ nivel, label, code }) => {
      const { unmount } = render(<DificuldadePill nivel={nivel} variante="sobreCapa" />)
      const pill = screen.getByText(label)
      expect(pill.style.backgroundColor).toBe('var(--dif-pill-sobre-capa-bg)')
      expect(pill.style.color).toBe(`var(--dif-${code}-text)`)
      expect(pill.style.borderColor).toBe(`var(--dif-${code}-border)`)
      unmount()
    })
  })

  it('aplica cores corretas para a variante neutra', () => {
    casos.forEach(({ nivel, label, code }) => {
      const { unmount } = render(<DificuldadePill nivel={nivel} variante="neutra" />)
      const pill = screen.getByText(label)
      expect(pill.style.backgroundColor).toBe('var(--dif-pill-neutra-bg)')
      expect(pill.style.color).toBe(`var(--dif-${code}-text)`)
      expect(pill.style.borderColor).toBe(`var(--dif-${code}-border)`)
      unmount()
    })
  })

  it('aplica cores corretas para a variante solida', () => {
    casos.forEach(({ nivel, label, code }) => {
      const { unmount } = render(<DificuldadePill nivel={nivel} variante="solida" />)
      const pill = screen.getByText(label)
      expect(pill.style.backgroundColor).toBe(`var(--dif-${code}-fill)`)
      expect(pill.style.color).toBe('var(--nota-badge-text)')
      unmount()
    })
  })
})
