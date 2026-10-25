import { Gamepad2 } from 'lucide-react'

export interface JogarAgoraBotaoProps {
  onClick: () => void
  isIniciando: boolean
  emAndamento: boolean
  className?: string
}

export function JogarAgoraBotao({
  onClick,
  isIniciando,
  emAndamento,
  className = '',
}: JogarAgoraBotaoProps) {
  const rotulo = isIniciando ? 'Iniciando...' : emAndamento ? 'Em andamento' : 'Jogar agora'

  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isIniciando || emAndamento}
      className={`btn-primario flex h-[40px] items-center gap-2 px-4 text-[14px] font-bold ${
        emAndamento ? 'btn-primario-disabled' : ''
      } ${className}`.trim()}
    >
      <Gamepad2 className="h-4 w-4" aria-hidden="true" />
      <span>{rotulo}</span>
    </button>
  )
}
