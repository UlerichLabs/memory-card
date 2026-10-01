import { describe, expect, it } from 'vitest'
import { diasDesdeInicio, formatarDataJogando, hojeIso, rotuloDiasDesdeInicio } from './jogandoUtils'

describe('jogandoUtils', () => {
  it('formata datas sem deslocamento de fuso', () => {
    expect(formatarDataJogando('2026-09-20T00:00:00Z')).toBe('20/09/2026')
    expect(formatarDataJogando(null)).toBe('sem data')
  })

  it('calcula dias por calendário local, incluindo viradas', () => {
    expect(diasDesdeInicio('2026-09-20T00:00:00Z', new Date(2026, 8, 20))).toBe(0)
    expect(diasDesdeInicio('2026-09-19T00:00:00Z', new Date(2026, 8, 20))).toBe(1)
    expect(diasDesdeInicio('2026-08-31T00:00:00Z', new Date(2026, 8, 2))).toBe(2)
    expect(diasDesdeInicio('2025-12-31T00:00:00Z', new Date(2026, 0, 1))).toBe(1)
  })

  it('produz os rótulos relativos', () => {
    const hoje = new Date(2026, 8, 20)
    expect(rotuloDiasDesdeInicio('2026-09-20T00:00:00Z', hoje)).toBe('começou hoje')
    expect(rotuloDiasDesdeInicio('2026-09-19T00:00:00Z', hoje)).toBe('há 1 dia')
    expect(rotuloDiasDesdeInicio('2026-09-17T00:00:00Z', hoje)).toBe('há 3 dias')
    expect(hojeIso(hoje)).toBe('2026-09-20')
  })
})
