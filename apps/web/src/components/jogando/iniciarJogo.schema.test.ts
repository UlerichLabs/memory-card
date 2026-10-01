import { describe, expect, it } from 'vitest'
import { iniciarJogoSchema } from './iniciarJogo.schema'

describe('iniciarJogoSchema', () => {
  it('aceita nome e data válidos', () => {
    expect(iniciarJogoSchema.safeParse({ nome: 'Jogo', iniciado_em: '2026-09-30' }).success).toBe(true)
  })

  it('rejeita nome vazio, data inválida e data futura', () => {
    expect(iniciarJogoSchema.safeParse({ nome: ' ', iniciado_em: '2026-09-30' }).success).toBe(false)
    expect(iniciarJogoSchema.safeParse({ nome: 'Jogo', iniciado_em: '2026-02-31' }).success).toBe(false)
    expect(iniciarJogoSchema.safeParse({ nome: 'Jogo', iniciado_em: '2099-01-01' }).success).toBe(false)
  })
})
