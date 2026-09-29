import { describe, expect, it } from 'vitest'
import { obterRotuloNota } from './rotulosNota'

describe('obterRotuloNota', () => {
  it('retorna os rótulos corretos de 1 a 11', () => {
    expect(obterRotuloNota(11)).toBe('Jogo da Vida')
    expect(obterRotuloNota(10)).toBe('Incrível')
    expect(obterRotuloNota(9)).toBe('Ótimo')
    expect(obterRotuloNota(8)).toBe('Muito bom')
    expect(obterRotuloNota(7)).toBe('Bom')
    expect(obterRotuloNota(6)).toBe('Decente')
    expect(obterRotuloNota(5)).toBe('Tanto faz')
    expect(obterRotuloNota(4)).toBe('Medíocre')
    expect(obterRotuloNota(3)).toBe('Ruim')
    expect(obterRotuloNota(2)).toBe('Terrível')
    expect(obterRotuloNota(1)).toBe('Tragédia')
  })

  it('retorna string vazia para nota inválida', () => {
    expect(obterRotuloNota(0)).toBe('')
    expect(obterRotuloNota(12)).toBe('')
  })
})
