import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAbandonadosStore } from '@/stores/abandonadosStore'
import { AbandonarJogoForm } from './AbandonarJogoForm'
import type { AbandonarJogoFormData } from './abandonarJogo.schema'

export function AbandonarJogoDialog() {
  const { isModalOpen, fecharModal, jogoEmEdicao, modalOpcoes, criarJogo, atualizarJogo } = useAbandonadosStore()
  const initialData = jogoEmEdicao ?? modalOpcoes?.valoresIniciais
  const origemFila = modalOpcoes?.origemFila

  const formKey = jogoEmEdicao ? `edit-${jogoEmEdicao.id}` : modalOpcoes?.valoresIniciais?.nome ?? 'novo-abandono'

  async function handleSalvar(data: AbandonarJogoFormData) {
    if (jogoEmEdicao) {
      await atualizarJogo(jogoEmEdicao.id, data)
    } else {
      await criarJogo(data)
      if (origemFila?.onSalvo) await origemFila.onSalvo()
    }
    fecharModal()
  }

  return (
    <Dialog open={isModalOpen} onOpenChange={(open) => !open && fecharModal()}>
      <DialogContent
        className={
          'flex !max-h-[calc(100vh-32px)] w-[min(700px,calc(100vw-32px))] flex-col gap-5 ' +
          'overflow-hidden rounded-2xl border border-[var(--modal-border)] bg-[var(--modal-bg)] ' +
          'px-8 py-7 text-[var(--text-primary)] sm:max-w-none'
        }
      >
        <DialogHeader>
          <DialogTitle className="text-[20px] font-bold text-[var(--text-primary)]">
            {jogoEmEdicao ? 'Editar abandono' : 'Abandonar jogo'}
          </DialogTitle>
          <DialogDescription className="text-[13px] text-[var(--text-secondary)]">
            {jogoEmEdicao
              ? `Atualize as informações do jogo "${jogoEmEdicao.nome}".`
              : 'Preencha os dados do jogo que você abandonou para manter o registro.'}
          </DialogDescription>
        </DialogHeader>

        <AbandonarJogoForm
          key={formKey}
          initialData={initialData}
          origemFila={origemFila}
          isEditing={!!jogoEmEdicao}
          onSubmit={handleSalvar}
          onCancel={fecharModal}
        />
      </DialogContent>
    </Dialog>
  )
}
