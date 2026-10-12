import { useState } from 'react'
import { iniciais } from '@/lib/iniciais'

interface UserAvatarProps {
  nome: string
  src?: string | null
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
}

const sizeClasses: Record<'sm' | 'md' | 'lg' | 'xl', string> = {
  sm: 'size-6 text-[10.5px]',
  md: 'size-9 text-xs',
  lg: 'size-12 text-sm',
  xl: 'size-20 text-2xl',
}

export function UserAvatar({ nome, src, size = 'sm', className = '' }: UserAvatarProps) {
  const [falhaSrc, setFalhaSrc] = useState<string | null>(null)
  const classes = sizeClasses[size] || sizeClasses.sm

  if (src && falhaSrc !== src) {
    return (
      <img
        src={src}
        alt={nome ? `Foto de perfil de ${nome}` : 'Foto de perfil'}
        onError={() => setFalhaSrc(src)}
        className={`shrink-0 rounded-full object-cover select-none ${classes} ${className}`}
      />
    )
  }

  return (
    <div
      aria-hidden="true"
      className={`flex shrink-0 items-center justify-center rounded-full bg-[var(--accent)] font-bold text-[var(--accent-foreground)] select-none ${classes} ${className}`}
    >
      {iniciais(nome ?? '')}
    </div>
  )
}
