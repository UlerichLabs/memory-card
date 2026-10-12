import { describe, expect, it } from 'vitest'
import { perfilSchema, PERFIL_BACKEND_ERROS } from './perfilSchema'

describe('perfilSchema', () => {
  const anoAtual = new Date().getFullYear()

  it('valida dados corretos com sucesso', () => {
    const input = {
      nome: 'Lucas Ulerich',
      username: 'lucas_123',
      bio: 'Gamer veterano',
      jogo_favorito_id: 10,
      console_favorito: 'Super Nintendo',
      jogando_desde: 1995,
    }

    const res = perfilSchema.safeParse(input)
    expect(res.success).toBe(true)
    if (res.success) {
      expect(res.data.nome).toBe('Lucas Ulerich')
      expect(res.data.jogando_desde).toBe(1995)
    }
  })

  it('rejeita nome vazio ou acima de 100 caracteres', () => {
    expect(perfilSchema.safeParse({ nome: '' }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: '   ' }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'a'.repeat(101) }).success).toBe(false)
  })

  it('valida formato de username quando preenchido e aceita vazio/nulo', () => {
    expect(perfilSchema.safeParse({ nome: 'Lucas', username: '' }).success).toBe(true)
    expect(perfilSchema.safeParse({ nome: 'Lucas', username: null }).success).toBe(true)
    expect(perfilSchema.safeParse({ nome: 'Lucas', username: 'ab' }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'Lucas', username: 'a'.repeat(31) }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'Lucas', username: 'user-name' }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'Lucas', username: 'user@name' }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'Lucas', username: 'user_123' }).success).toBe(true)
  })

  it('rejeita bio acima de 280 caracteres e aceita vazio/nulo', () => {
    expect(perfilSchema.safeParse({ nome: 'Lucas', bio: 'a'.repeat(280) }).success).toBe(true)
    expect(perfilSchema.safeParse({ nome: 'Lucas', bio: 'a'.repeat(281) }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'Lucas', bio: null }).success).toBe(true)
  })

  it('valida ano jogando_desde entre 1970 e ano atual', () => {
    expect(perfilSchema.safeParse({ nome: 'Lucas', jogando_desde: 1969 }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'Lucas', jogando_desde: anoAtual + 1 }).success).toBe(false)
    expect(perfilSchema.safeParse({ nome: 'Lucas', jogando_desde: 1970 }).success).toBe(true)
    expect(perfilSchema.safeParse({ nome: 'Lucas', jogando_desde: anoAtual }).success).toBe(true)
    expect(perfilSchema.safeParse({ nome: 'Lucas', jogando_desde: '' }).success).toBe(true)
    expect(perfilSchema.safeParse({ nome: 'Lucas', jogando_desde: null }).success).toBe(true)
  })

  it('contém mapeamento para todos os códigos de erro conhecidos do backend', () => {
    expect(PERFIL_BACKEND_ERROS['perfil.nome_obrigatorio'].campo).toBe('nome')
    expect(PERFIL_BACKEND_ERROS['perfil.username_em_uso'].campo).toBe('username')
    expect(PERFIL_BACKEND_ERROS['perfil.bio_muito_longa'].campo).toBe('bio')
    expect(PERFIL_BACKEND_ERROS['perfil.jogo_favorito_nao_encontrado'].campo).toBe('jogo_favorito_id')
    expect(PERFIL_BACKEND_ERROS['perfil.console_invalido'].campo).toBe('console_favorito')
    expect(PERFIL_BACKEND_ERROS['perfil.jogando_desde_invalido'].campo).toBe('jogando_desde')
    expect(PERFIL_BACKEND_ERROS['perfil.entrada_invalida'].campo).toBe('geral')
  })
})
