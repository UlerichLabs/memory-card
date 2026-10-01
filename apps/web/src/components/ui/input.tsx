import * as React from 'react'
import { Input as InputPrimitive } from '@base-ui/react/input'
import { cn } from 'cn'

function Input({ className, type, ...props }: React.ComponentProps<'input'>) {
  return (
    <InputPrimitive
      type={type}
      data-slot="input"
      className={cn(
        'h-8 w-full min-w-0 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2.5 py-1 text-sm text-[var(--text-primary)] transition-colors outline-none placeholder:text-[var(--text-muted)] focus-visible:border-[var(--accent)] focus-visible:ring-2 focus-visible:ring-[var(--accent)]/30 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-[var(--danger)] aria-invalid:ring-2 aria-invalid:ring-[var(--danger)]/20',
        className
      )}
      {...props}
    />
  )
}

export { Input }
