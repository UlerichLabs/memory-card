import type { Dificuldade } from '@/types/jogos'

export interface DificuldadeBarraProps {
  nivel: Dificuldade
  className?: string
}

const NIVEIS: Dificuldade[] = ['C', 'B', 'A', 'AA', 'AAA']

export function DificuldadeBarra({ nivel, className = '' }: DificuldadeBarraProps) {
  const indexAtivo = NIVEIS.indexOf(nivel)

  return (
    <div
      aria-hidden="true"
      className={`flex items-center gap-1 ${className}`.trim()}
    >
      {NIVEIS.map((n, i) => {
        const ativo = i <= indexAtivo
        const code = n.toLowerCase()
        const backgroundColor = ativo ? `var(--dif-${code}-fill)` : 'var(--dif-barra-trilho)'

        return (
          <div
            key={n}
            style={{ backgroundColor }}
            className="h-[6px] w-[22px] rounded-[3px] shrink-0 transition-colors"
          />
        )
      })}
    </div>
  )
}
