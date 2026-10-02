import { getConsoleTema } from '@/lib/consoles'

export interface ConsoleBadgeProps {
  nome: string
  tamanho?: 'sm' | 'md'
  variante?: 'suave' | 'solido'
}

export function ConsoleBadge({ nome, tamanho = 'sm', variante = 'suave' }: ConsoleBadgeProps) {
  const tema = getConsoleTema(nome)
  const Icone = variante === 'solido' ? tema.Icone : tema.IconeSuave
  const classes = tamanho === 'md' ? 'h-7 px-3' : 'h-6 px-2'
  const solido = variante === 'solido'

  return (
    <span
      title={nome}
      className={`inline-flex min-h-6 min-w-0 max-w-full items-center gap-1.5 rounded-full border text-xs font-semibold ${classes}`}
      style={{
        color: solido ? tema.corTextoSolida : `color-mix(in srgb, ${tema.cor} 55%, white)`,
        backgroundColor: solido ? tema.corSolida : tema.corFundo,
        borderColor: solido ? tema.corSolida : tema.corBorda,
      }}
    >
      <Icone className="h-3.5 w-3.5 shrink-0" style={{ color: solido ? tema.corTextoSolida : tema.cor }} aria-hidden="true" />
      <span className="truncate">{nome}</span>
    </span>
  )
}
