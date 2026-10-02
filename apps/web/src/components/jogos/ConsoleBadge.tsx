import { getConsoleTema } from '@/lib/consoles'

export interface ConsoleBadgeProps {
  nome: string
  tamanho?: 'sm' | 'md'
}

export function ConsoleBadge({ nome, tamanho = 'sm' }: ConsoleBadgeProps) {
  const tema = getConsoleTema(nome)
  const Icone = tema.Icone
  const classes = tamanho === 'md' ? 'h-7 px-3' : 'h-6 px-2'

  return (
    <span
      title={nome}
      className={`inline-flex min-h-6 min-w-0 max-w-full items-center gap-1.5 rounded-full border text-xs font-semibold ${classes}`}
      style={{ color: `color-mix(in srgb, ${tema.cor} 55%, white)`, backgroundColor: tema.corFundo, borderColor: tema.corBorda }}
    >
      <Icone className="h-3.5 w-3.5 shrink-0" style={{ color: tema.cor }} aria-hidden="true" />
      <span className="truncate">{nome}</span>
    </span>
  )
}
