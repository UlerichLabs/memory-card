import type { CSSProperties } from 'react'

export const bibliotecaTheme = {
  '--bg-primary': '#15161A',
  '--bg-surface': '#1A1B20',
  '--bg-surface-alt': '#1D1F25',
  '--border': '#24262C',
  '--border-subtle': '#2A2C33',
  '--text-primary': '#EDEDED',
  '--text-secondary': '#9A9CA5',
  '--text-muted': '#6B6D76',
  '--text-faint': '#52545C',
  '--accent': '#4F7CFF',
  '--highlight-gold': '#E8C15C',
  '--danger': '#E05A4E',
  backgroundColor: 'var(--bg-primary)',
  color: 'var(--text-primary)',
  fontFamily: 'Inter, sans-serif',
} as CSSProperties

export const NOTA_OPCOES = [
  { value: '', label: 'Todas as notas' },
  { value: '10', label: 'Nota 10+' },
  { value: '9', label: 'Nota 9+' },
  { value: '8', label: 'Nota 8+' },
  { value: '7', label: 'Nota 7+' },
  { value: '6', label: 'Nota 6+' },
]
