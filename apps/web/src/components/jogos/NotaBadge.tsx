import type { CSSProperties } from 'react'

export interface NotaBadgeProps {
  nota: number
  tamanho: 'sm' | 'md' | 'lg'
  pulsar?: boolean
  className?: string
}

export function NotaBadge({ nota, tamanho, pulsar = false, className = '' }: NotaBadgeProps) {
  const isNota11 = nota === 11

  const sizeClasses = {
    sm: 'min-w-[32px] min-h-[32px] h-8 px-1.5 rounded-[8px] text-[14px] font-bold',
    md: 'w-[44px] h-[44px] rounded-[10px] text-[20px] font-extrabold',
    lg: 'w-[56px] h-[56px] rounded-[12px] text-[26px] font-extrabold border',
  }[tamanho]

  const style: CSSProperties = {
    backgroundColor: `var(--nota-${nota}-fill)`,
    color: 'var(--nota-badge-text)',
  }

  if (tamanho === 'lg') {
    style.borderColor = `var(--nota-${nota}-fill-border)`
  }

  if (isNota11 && !pulsar) {
    style.boxShadow = '0 0 12px 2px rgba(240, 190, 80, 0.45)'
  }

  const animationClass = isNota11 && pulsar ? 'animate-pulso-ouro' : ''

  return (
    <span
      aria-label={`Nota ${nota}`}
      style={style}
      className={`inline-flex items-center justify-center select-none shrink-0 ${sizeClasses} ${animationClass} ${className}`.trim()}
    >
      {nota}
    </span>
  )
}
