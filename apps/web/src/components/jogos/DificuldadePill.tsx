import type { CSSProperties } from 'react'
import type { Dificuldade } from '@/types/jogos'

export interface DificuldadePillProps {
  nivel: Dificuldade
  variante: 'sobreCapa' | 'neutra' | 'solida'
  className?: string
}

const NOMES_DIFICULDADE: Record<Dificuldade, string> = {
  C: 'Muito fácil',
  B: 'Fácil',
  A: 'Normal',
  AA: 'Difícil',
  AAA: 'Muito difícil',
}

export function DificuldadePill({ nivel, variante, className = '' }: DificuldadePillProps) {
  const code = nivel.toLowerCase()
  const nome = NOMES_DIFICULDADE[nivel] || nivel

  const style: CSSProperties = {}

  if (variante === 'sobreCapa') {
    style.backgroundColor = 'var(--dif-pill-sobre-capa-bg)'
    style.color = `var(--dif-${code}-text)`
    style.borderColor = `var(--dif-${code}-border)`
  } else if (variante === 'neutra') {
    style.backgroundColor = 'var(--dif-pill-neutra-bg)'
    style.color = `var(--dif-${code}-text)`
    style.borderColor = `var(--dif-${code}-border)`
  } else {
    style.backgroundColor = `var(--dif-${code}-fill)`
    style.color = 'var(--nota-badge-text)'
  }

  const variantClasses = {
    sobreCapa: 'px-2 py-1 text-[11px] font-semibold border rounded-full backdrop-blur-xs',
    neutra: 'px-2.5 py-1 text-[12px] font-semibold border rounded-full',
    solida: 'px-3 py-[5px] text-[14px] md:px-3.5 md:py-1.5 md:text-[15px] font-bold rounded-full',
  }[variante]

  return (
    <span
      style={style}
      className={`inline-flex items-center justify-center select-none shrink-0 ${variantClasses} ${className}`.trim()}
    >
      {nome}
    </span>
  )
}
