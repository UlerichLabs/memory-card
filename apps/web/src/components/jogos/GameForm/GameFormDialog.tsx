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
  const { isModalOpen, fecharModal, jogoEmEdicao, modalRegistroOpcoes, criarJogo, atualizarJogo } = useJogosStore()

  async function handleSubmit(payload: SalvarJogoPayload) {
    if (jogoEmEdicao) {
      await atualizarJogo(jogoEmEdicao.id, payload)
    } else {
      const criado = await criarJogo(payload)
      if (modalRegistroOpcoes?.onSalvo) {
        await modalRegistroOpcoes.onSalvo(criado)
      }
    }
    fecharModal()
  }

  const initialData = jogoEmEdicao ?? modalRegistroOpcoes?.valoresIniciais

  return (
    <Dialog open={isModalOpen} onOpenChange={(open) => !open && fecharModal()}>
      <DialogContent className="flex !max-h-[calc(100vh-32px)] w-[min(1040px,calc(100vw-32px))] flex-col gap-5 overflow-hidden rounded-2xl border border-[var(--modal-border)] bg-[var(--modal-bg)] px-8 py-7 text-[var(--text-primary)] sm:max-w-none">
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
          key={jogoEmEdicao ? `edit-${jogoEmEdicao.id}` : modalRegistroOpcoes?.valoresIniciais?.nome ?? 'novo-jogo'}
          initialData={initialData}
          isEditing={!!jogoEmEdicao}
          aviso={modalRegistroOpcoes?.aviso}
          textoSubmit={modalRegistroOpcoes?.textoSubmit}
          onSubmit={handleSubmit}
          onCancel={fecharModal}
        />
      </DialogContent>
    </Dialog>
  )
}
