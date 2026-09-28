import type { JogoZeradoDTO } from '@/types/jogos'
import { BibliotecaCard } from './BibliotecaCard'

export interface BibliotecaGradeProps {
  jogos: JogoZeradoDTO[]
  onEditar: (jogo: JogoZeradoDTO) => void
  onExcluir: (jogo: JogoZeradoDTO) => void
  onDetalhes?: (jogo: JogoZeradoDTO) => void
}

export function BibliotecaGrade({
  jogos,
  onEditar,
  onExcluir,
  onDetalhes,
}: BibliotecaGradeProps) {
  return (
    <div
      data-testid="biblioteca-grade"
      className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
    >
      {jogos.map((jogo) => (
        <BibliotecaCard
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
