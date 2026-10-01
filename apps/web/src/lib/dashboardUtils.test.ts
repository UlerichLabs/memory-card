import { describe, expect, it } from 'vitest'
import { formatarData } from './dashboardUtils'

describe('dashboardUtils', () => {
  it('formata datas UTC sem deslocar o dia', () => {
    expect(formatarData('2026-09-20T00:00:00Z')).toBe('20/09/2026')
    expect(formatarData(null)).toBe('sem data')
  })
})
