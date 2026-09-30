import type { JogoAbandonado } from '@/types/abandonados'
import { AbandonadoCard } from './AbandonadoCard'

export interface AbandonadosGradeProps {
  jogos: JogoAbandonado[]
  onRetomar: (jogo: JogoAbandonado) => void
  onEditar: (jogo: JogoAbandonado) => void
  onExcluir: (jogo: JogoAbandonado) => void
}

export function AbandonadosGrade({ jogos, onRetomar, onEditar, onExcluir }: AbandonadosGradeProps) {
  return (
    <section
      aria-label="Jogos abandonados"
      className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
    >
      {jogos.map((jogo) => (
        <AbandonadoCard
          key={jogo.id}
          jogo={jogo}
          onRetomar={onRetomar}
          onEditar={onEditar}
          onExcluir={onExcluir}
        />
      ))}
    </section>
  )
}
