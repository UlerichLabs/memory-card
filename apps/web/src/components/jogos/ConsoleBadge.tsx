import { getConsoleTema } from '@/lib/consoles'

export interface ConsoleBadgeProps {
  nome: string
  tamanho?: 'sm' | 'md'
}

export function ConsoleBadge({ nome, tamanho = 'sm' }: ConsoleBadgeProps) {
  const tema = getConsoleTema(nome)
  const Icone = tema.Icone
  const classes = tamanho === 'md' ? 'h-7 px-3 text-[12px]' : 'h-6 px-2 text-[10px]'

  return (
    <span
      title={nome}
      className={`inline-flex min-w-0 max-w-full items-center gap-1.5 rounded-full border font-medium ${classes}`}
      style={{ color: tema.cor, backgroundColor: tema.corFundo, borderColor: tema.corBorda }}
    >
      <Icone className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
      <span className="truncate">{nome}</span>
    </span>
  )
}
