import { describe, expect, it } from 'vitest'
import { abandonarJogoSchema } from './abandonarJogo.schema'

describe('abandonarJogoSchema', () => {
  const dadosValidos = {
    nome: 'Chrono Trigger',
    console: 'SNES',
    abandonado_em: '2026-01-15',
    tempo_jogado_horas: 10,
    tempo_jogado_minutos: 30,
    tempo_jogado_segundos: 0,
    motivo: 'Fiquei preso num puzzle',
  }

  it('valida dados corretos com sucesso', () => {
    const res = abandonarJogoSchema.safeParse(dadosValidos)
    expect(res.success).toBe(true)
  })

  it('falha quando nome ou console estão vazios', () => {
    expect(abandonarJogoSchema.safeParse({ ...dadosValidos, nome: '   ' }).success).toBe(false)
    expect(abandonarJogoSchema.safeParse({ ...dadosValidos, console: '' }).success).toBe(false)
  })

  it('falha quando data é futura', () => {
    const res = abandonarJogoSchema.safeParse({ ...dadosValidos, abandonado_em: '2099-01-01' })
    expect(res.success).toBe(false)
    if (!res.success) {
      expect(res.error.issues[0].message).toBe('Data de abandono não pode ser futura')
    }
  })

  it('falha quando horas ultrapassam 100.000', () => {
    const res = abandonarJogoSchema.safeParse({ ...dadosValidos, tempo_jogado_horas: 100001 })
    expect(res.success).toBe(false)
  })

  it('aceita exatamente 100.000 horas com 0 minutos e 0 segundos', () => {
    const res = abandonarJogoSchema.safeParse({
      ...dadosValidos,
      tempo_jogado_horas: 100000,
      tempo_jogado_minutos: 0,
      tempo_jogado_segundos: 0,
    })
    expect(res.success).toBe(true)
  })

  it('falha quando minutos ou segundos são maiores que 59', () => {
    expect(abandonarJogoSchema.safeParse({ ...dadosValidos, tempo_jogado_minutos: 60 }).success).toBe(false)
    expect(abandonarJogoSchema.safeParse({ ...dadosValidos, tempo_jogado_segundos: 60 }).success).toBe(false)
  })

  it('falha quando motivo ultrapassa 500 caracteres', () => {
    const res = abandonarJogoSchema.safeParse({ ...dadosValidos, motivo: 'a'.repeat(501) })
    expect(res.success).toBe(false)
  })
})
