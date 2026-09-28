import { describe, expect, it } from 'vitest'
import { gameFormSchema } from './GameForm.schema'

const valido = {
  nome: 'Jogo', console: 'PC', finalizado_em: '2026-03-13',
  tempo_jogado_horas: 0, tempo_jogado_minutos: 0, tempo_jogado_segundos: 0,
  nota: 10, dificuldade: 'A' as const, destaque: false,
}

describe('gameFormSchema', () => {
  it.each([
    ['nome', 201], ['console', 101], ['genero', 151], ['tipo', 51], ['review', 5001],
  ])('rejeita %s acima do limite', (campo, tamanho) => {
    const resultado = gameFormSchema.safeParse({ ...valido, [campo]: 'a'.repeat(tamanho) })
    expect(resultado.success).toBe(false)
    if (!resultado.success) expect(resultado.error.issues[0].path).toEqual([campo])
  })

  it('aceita review no limite de 5.000 caracteres', () => {
    expect(gameFormSchema.safeParse({ ...valido, review: 'a'.repeat(5000) }).success).toBe(true)
  })
})
