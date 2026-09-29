import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { useListasStore } from '@/stores/listasStore'
import { useJogosStore } from '@/stores/jogosStore'
import { AdicionarJogoInput } from './AdicionarJogoInput'
import { FilaItem } from './FilaItem'
import type { ListaDetalhada, ListaItem } from '@/types/listas'

export interface FilaListaProps {
  lista: ListaDetalhada
}

export function FilaLista({ lista }: FilaListaProps) {
  const { reordenarItens, adicionarItem, removerItem, associarZeramento, error } = useListasStore()
  const { abrirModalRegistro } = useJogosStore()

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 200, tolerance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  const todosItens = lista.itens ?? []
  const pendentes = todosItens
    .filter((it) => !it.zerado)
    .sort((a, b) => a.posicao - b.posicao)
  const zerados = todosItens
    .filter((it) => it.zerado)
    .sort((a, b) => a.posicao - b.posicao)

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = pendentes.findIndex((it) => it.id === active.id)
    const newIndex = pendentes.findIndex((it) => it.id === over.id)
    if (oldIndex === -1 || newIndex === -1) return
    const reordenados = arrayMove(pendentes, oldIndex, newIndex)
    const todosIds = [...reordenados.map((it) => it.id), ...zerados.map((it) => it.id)]
    reordenarItens(todosIds).catch(() => undefined)
  }

  const handleZerei = (item: ListaItem) => {
    abrirModalRegistro({
      valoresIniciais: {
        nome: item.nome,
        igdb_id: item.igdb_id ?? undefined,
        igdb_capa_url: item.igdb_capa_url ?? undefined,
        console: item.console ?? undefined,
      },
      onSalvo: async (jogo) => {
        await associarZeramento(item.id, jogo.id)
      },
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <AdicionarJogoInput
        onAdicionar={adicionarItem}
        placeholder="Adicionar jogo: busque no IGDB ou digite o nome…"
        ariaLabel="Adicionar jogo à fila"
      />

      {error && <span className="text-xs text-[var(--danger)]">{error}</span>}

      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <div className="flex items-baseline gap-2">
            <span className="text-[15px] font-semibold text-[var(--text-primary)]">A jogar</span>
            <span className="text-[13px] text-[var(--lista-text-muted)]">
              {pendentes.length} {pendentes.length === 1 ? 'jogo' : 'jogos'}
            </span>
          </div>
          <span className="text-[12px] text-[var(--lista-text-dim)]">
            Arraste para mudar a ordem
          </span>
        </div>

        {pendentes.length === 0 ? (
          <div className="py-8 text-center text-[14px] text-[var(--lista-text-muted)]">
            Nenhum jogo na fila. Adicione o próximo que você quer zerar.
          </div>
        ) : (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={pendentes.map((it) => it.id)} strategy={verticalListSortingStrategy}>
              <ol className="flex flex-col gap-2">
                {pendentes.map((it, idx) => (
                  <FilaItem
                    key={it.id}
                    item={it}
                    posicaoExibicao={idx + 1}
                    isFirst={idx === 0}
                    onZerei={handleZerei}
                    onRemover={removerItem}
                  />
                ))}
              </ol>
            </SortableContext>
          </DndContext>
        )}
      </div>
    </div>
  )
}
