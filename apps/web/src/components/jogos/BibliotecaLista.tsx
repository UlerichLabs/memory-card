import type { JogoZeradoDTO } from '@/types/jogos'
import { BibliotecaListItem } from './BibliotecaListItem'

export interface BibliotecaListaProps {
  jogos: JogoZeradoDTO[]
  onEditar: (jogo: JogoZeradoDTO) => void
  onExcluir: (jogo: JogoZeradoDTO) => void
  onDetalhes?: (jogo: JogoZeradoDTO) => void
}

export function BibliotecaLista({
  jogos,
  onEditar,
  onExcluir,
  onDetalhes,
}: BibliotecaListaProps) {
  return (
    <div data-testid="biblioteca-lista" className="flex flex-col gap-2">
      {jogos.map((jogo) => (
        <BibliotecaListItem
          key={jogo.id}
          jogo={jogo}
          onEditar={onEditar}
          onExcluir={onExcluir}
          onDetalhes={onDetalhes}
        />
      ))}
    </div>
  )
}
