import { describe, expect, it } from 'vitest'
import { formatarCapaIGDB, isoParaDataPt, dataPtParaIso, aplicarMascaraData } from './utils'

describe('formatarCapaIGDB', () => {
  it('troca t_thumb por t_cover_big', () => {
    const input = 'https://images.igdb.com/igdb/image/upload/t_thumb/co39u6.jpg'
    expect(formatarCapaIGDB(input)).toBe('https://images.igdb.com/igdb/image/upload/t_cover_big/co39u6.jpg')
  })

  it('prefixa https: quando a url começar com //', () => {
    const input = '//images.igdb.com/igdb/image/upload/t_thumb/co39u6.jpg'
    expect(formatarCapaIGDB(input)).toBe('https://images.igdb.com/igdb/image/upload/t_cover_big/co39u6.jpg')
  })

  it('retorna string vazia sem quebrar para url vazia ou undefined', () => {
    expect(formatarCapaIGDB('')).toBe('')
    expect(formatarCapaIGDB(undefined)).toBe('')
  })
})

describe('funções de data', () => {
  it('converte iso para data em pt', () => {
    expect(isoParaDataPt('2026-03-28T00:00:00Z')).toBe('28/03/2026')
    expect(isoParaDataPt('')).toBe('')
  })

  it('converte data em pt para iso', () => {
    expect(dataPtParaIso('28/03/2026')).toBe('2026-03-28')
  })

  it('aplica mascara de data', () => {
    expect(aplicarMascaraData('28032026')).toBe('28/03/2026')
  })
})
