import { describe, expect, it } from 'vitest'
import { getConsoleTema } from './consoles'

describe('getConsoleTema', () => {
  it.each([
    ['Sega Mega Drive/Genesis', 'Sega', '#3B5BDB'],
    ['Genesis', 'Sega', '#3B5BDB'],
    ['Sega Game Gear', 'Sega', '#3B5BDB'],
    ['Nintendo Entertainment System', 'Nintendo', '#E60012'],
    ['Nintendo Switch', 'Nintendo', '#E60012'],
    ['Super Nintendo', 'Nintendo', '#E60012'],
    ['Nintendo DS', 'Nintendo', '#E60012'],
    ['Nintendo 3DS', 'Nintendo', '#E60012'],
    ['Game Boy Advance', 'Nintendo', '#E60012'],
    ['PlayStation 5', 'PlayStation', '#0070D1'],
    ['PlayStation Vita', 'PlayStation', '#0070D1'],
    ['PSP', 'PlayStation', '#0070D1'],
    ['Xbox Series X|S', 'Xbox', '#107C10'],
    ['Xbox One', 'Xbox', '#107C10'],
    ['Xbox 360', 'Xbox', '#107C10'],
    ['PC (Microsoft Windows)', 'PC', '#8FA8C8'],
    ['Linux', 'PC', '#8FA8C8'],
    ['Mac', 'PC', '#8FA8C8'],
    ['Arcade', 'Arcade', '#D946EF'],
    ['Atari 2600', 'Atari', '#F28C28'],
    ['Android', 'Mobile', '#14B8A6'],
    ['iOS', 'Mobile', '#14B8A6'],
  ])('classifica %s como %s', (nome, familia, cor) => {
    const tema = getConsoleTema(nome)
    expect(tema.familia).toBe(familia)
    expect(tema.cor).toBe(cor)
    expect(['function', 'object']).toContain(typeof tema.Icone)
  })

  it.each([
    ['Nintendo Switch', '#E60012', '#fff'],
    ['PlayStation 5', '#0070D1', '#fff'],
    ['Xbox Series X|S', '#107C10', '#fff'],
    ['PC', '#4B5563', '#fff'],
    ['Sega Genesis', '#1D4ED8', '#fff'],
    ['Atari 2600', '#F28C28', '#1A1B20'],
    ['Arcade', '#A21CAF', '#fff'],
    ['Android', '#0F9D8A', '#fff'],
    ['Neo Geo', '#4B5563', '#fff'],
  ])('expõe cores sólidas acessíveis para %s', (nome, cor, corTexto) => {
    const tema = getConsoleTema(nome)
    expect(tema.corSolida).toBe(cor)
    expect(tema.corTextoSolida).toBe(corTexto)
  })

  it('normaliza caixa e acentos e oferece fallback', () => {
    expect(getConsoleTema('SUPER NÍNTENDO').familia).toBe('Nintendo')
    for (const nome of ['Neo Geo', 'Amiga', '3DO', 'PC Engine']) {
      expect(getConsoleTema(nome).familia).toBe('Outro')
    }
    expect(getConsoleTema('Console desconhecido').cor).toBe('var(--text-secondary)')
  })

  it('usa o mesmo ícone universal para toda a família Nintendo', () => {
    expect(getConsoleTema('Nintendo Switch').Icone).toBe(getConsoleTema('Game Boy Color').Icone)
  })

  it('usa o logo de Xbox para todas as plataformas Xbox', () => {
    expect(getConsoleTema('Xbox One').Icone).toBe(getConsoleTema('Xbox Series X').Icone)
  })
})
