import { describe, expect, it } from 'vitest'
import {
  formatarData,
  formatarDuracao,
  formatarHoras,
  formatarPercentual,
  iniciais,
} from './dashboardUtils'

describe('dashboardUtils', () => {
  it('formata horas com arredondamento e milhar pt-BR', () => {
    expect(formatarHoras(0)).toBe('0m')
    expect(formatarHoras(59)).toBe('1m')
    expect(formatarHoras(3599)).toBe('1h')
    expect(formatarHoras(3600)).toBe('1h')
    expect(formatarHoras(5400)).toBe('2h')
    expect(formatarHoras(10584000)).toBe('2.940h')
  })

  it('formata durações completas', () => {
    expect(formatarDuracao(0)).toBe('0m')
    expect(formatarDuracao(59)).toBe('1m')
    expect(formatarDuracao(3600)).toBe('1h 00m')
    expect(formatarDuracao(7800)).toBe('2h 10m')
    expect(formatarDuracao(511200)).toBe('142h 00m')
  })

  it('formata percentuais inteiros e decimais', () => {
    expect(formatarPercentual(25)).toBe('25')
    expect(formatarPercentual(33.333)).toBe('33.3')
  })

  it('formata datas em UTC e trata ausência', () => {
    expect(formatarData('2026-09-20T00:00:00Z')).toBe('20/09/2026')
    expect(formatarData('2026-12-31T23:59:59Z')).toBe('31/12/2026')
    expect(formatarData(null)).toBe('sem data')
  })

  it('extrai iniciais ignorando espaços extras', () => {
    expect(iniciais('Zelda')).toBe('Z')
    expect(iniciais('Hollow Knight')).toBe('HK')
    expect(iniciais('The Legend of Zelda')).toBe('TL')
    expect(iniciais('')).toBe('')
    expect(iniciais('  hollow   knight  ')).toBe('HK')
  })
})
