import { describe, expect, it } from 'vitest'
import { iniciais } from './iniciais'

describe('iniciais', () => {
  it('retorna a primeira letra do primeiro e do último nome em maiúsculas', () => {
    expect(iniciais('Lucas Ulerich')).toBe('LU')
    expect(iniciais('Lucas da Silva Ulerich')).toBe('LU')
  })

  it('retorna a primeira letra para nome único', () => {
    expect(iniciais('Lucas')).toBe('L')
    expect(iniciais('a')).toBe('A')
  })

  it('ignora espaços extras nas extremidades e entre nomes', () => {
    expect(iniciais('  Lucas   Ulerich  ')).toBe('LU')
    expect(iniciais('   Lucas   ')).toBe('L')
  })

  it('retorna ? para nome vazio ou apenas com espaços', () => {
    expect(iniciais('')).toBe('?')
    expect(iniciais('   ')).toBe('?')
  })
})
