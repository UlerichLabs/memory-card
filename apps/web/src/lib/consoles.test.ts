import { describe, expect, it } from 'vitest'
import { getConsoleTema } from './consoles'

describe('getConsoleTema', () => {
  it.each([
    ['Nintendo Switch', 'Nintendo', '#E60012'],
    ['PlayStation 5', 'PlayStation', '#0070D1'],
    ['Xbox Series X', 'Xbox', '#107C10'],
    ['Windows PC', 'PC', '#8FA8C8'],
    ['Mega Drive', 'Sega', '#3B5BDB'],
    ['Atari 2600', 'Atari', '#F28C28'],
    ['Fliperama', 'Arcade', '#D946EF'],
    ['Android', 'Mobile', '#14B8A6'],
  ])('classifica %s como %s', (nome, familia, cor) => {
    const tema = getConsoleTema(nome)
    expect(tema.familia).toBe(familia)
    expect(tema.cor).toBe(cor)
  })

  it('normaliza caixa e acentos e oferece fallback', () => {
    expect(getConsoleTema('SUPER NÍNTENDO').familia).toBe('Nintendo')
    expect(getConsoleTema('Console desconhecido').familia).toBe('Outro')
    expect(getConsoleTema('Console desconhecido').cor).toBe('var(--text-secondary)')
  })
})
