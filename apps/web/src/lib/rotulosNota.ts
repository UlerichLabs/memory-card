const ROTULOS_NOTA: Record<number, string> = {
  11: 'Jogo da Vida',
  10: 'Incrível',
  9: 'Ótimo',
  8: 'Muito bom',
  7: 'Bom',
  6: 'Decente',
  5: 'Tanto faz',
  4: 'Medíocre',
  3: 'Ruim',
  2: 'Terrível',
  1: 'Tragédia',
}

export function obterRotuloNota(nota: number): string {
  return ROTULOS_NOTA[nota] || ''
}
