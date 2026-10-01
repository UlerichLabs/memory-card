import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { DificuldadeBarra } from './DificuldadeBarra'
import type { Dificuldade } from '@/types/jogos'

describe('DificuldadeBarra', () => {
  const casos: Array<{ nivel: Dificuldade; preenchidos: number }> = [
    { nivel: 'C', preenchidos: 1 },
    { nivel: 'B', preenchidos: 2 },
    { nivel: 'A', preenchidos: 3 },
    { nivel: 'AA', preenchidos: 4 },
    { nivel: 'AAA', preenchidos: 5 },
  ]

  it('preenche a quantidade correta de segmentos por nível', () => {
    casos.forEach(({ nivel, preenchidos }) => {
      const { container, unmount } = render(<DificuldadeBarra nivel={nivel} />)
      const barra = container.firstElementChild
      expect(barra).toHaveAttribute('aria-hidden', 'true')
      const segmentos = barra ? Array.from(barra.children) as HTMLElement[] : []
      expect(segmentos).toHaveLength(5)

      segmentos.forEach((seg, i) => {
        if (i < preenchidos) {
          expect(seg.style.backgroundColor).not.toBe('var(--dif-barra-trilho)')
        } else {
          expect(seg.style.backgroundColor).toBe('var(--dif-barra-trilho)')
        }
      })
      unmount()
    })
  })
})
