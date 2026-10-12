import { iniciais } from '@/lib/iniciais'

interface UserAvatarProps {
  nome: string
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const sizeClasses: Record<'sm' | 'md' | 'lg', string> = {
  sm: 'size-6 text-[10.5px]',
  md: 'size-9 text-xs',
  lg: 'size-12 text-sm',
}

export function UserAvatar({ nome, size = 'sm', className = '' }: UserAvatarProps) {
  const classes = sizeClasses[size] || sizeClasses.sm
  return (
    <div
      aria-hidden="true"
      className={`flex shrink-0 items-center justify-center rounded-full bg-[var(--accent)] font-bold text-[var(--accent-foreground)] select-none ${classes} ${className}`}
    >
      {iniciais(nome ?? '')}
    </div>
  )
}
