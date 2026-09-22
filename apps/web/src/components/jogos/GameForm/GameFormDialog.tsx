import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { GameForm } from './GameForm'
import { useJogosStore } from '@/stores/jogosStore'
import type { SalvarJogoPayload } from '@/lib/services/jogosService'

export function GameFormDialog() {
  const { isModalOpen, fecharModal, jogoEmEdicao, criarJogo, atualizarJogo } = useJogosStore()

  async function handleSubmit(payload: SalvarJogoPayload) {
    if (jogoEmEdicao) {
      await atualizarJogo(jogoEmEdicao.id, payload)
    } else {
      await criarJogo(payload)
    }
    fecharModal()
  }

  return (
    <Dialog open={isModalOpen} onOpenChange={(open) => !open && fecharModal()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl bg-[var(--bg-surface)] border-[var(--border)] text-[var(--text-primary)] p-6">
        <DialogHeader>
          <DialogTitle className="text-[20px] font-bold text-[var(--text-primary)]">
            {jogoEmEdicao ? 'Editar registro' : 'Registrar jogo'}
          </DialogTitle>
          <DialogDescription className="text-[13px] text-[var(--text-secondary)]">
            {jogoEmEdicao
              ? `Atualize as informações do registro de "${jogoEmEdicao.nome}".`
              : 'Preencha os dados do jogo que você zerou para salvar em seu histórico.'}
          </DialogDescription>
        </DialogHeader>

        <GameForm
          key={jogoEmEdicao ? `edit-${jogoEmEdicao.id}` : 'novo-jogo'}
          initialData={jogoEmEdicao ?? undefined}
          isEditing={!!jogoEmEdicao}
          onSubmit={handleSubmit}
          onCancel={fecharModal}
        />
      </DialogContent>
    </Dialog>
  )
}
