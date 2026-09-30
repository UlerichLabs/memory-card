import { useState } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAbandonadosStore } from '@/stores/abandonadosStore'
import { AbandonarJogoForm } from './AbandonarJogoForm'
import { ERRO_FALHA_REMOCAO_FILA } from './abandonados.constants'
import type { AbandonarJogoFormData } from './abandonarJogo.schema'

export function AbandonarJogoDialog() {
  const { isModalOpen, fecharModal, jogoEmEdicao, modalOpcoes, criarJogo, atualizarJogo } = useAbandonadosStore()
  const [jogoCriadoId, setJogoCriadoId] = useState<number | null>(null)
  const [prevModalOpen, setPrevModalOpen] = useState(isModalOpen)
  const initialData = jogoEmEdicao ?? modalOpcoes?.valoresIniciais
  const origemFila = modalOpcoes?.origemFila

  const formKey = jogoEmEdicao ? `edit-${jogoEmEdicao.id}` : modalOpcoes?.valoresIniciais?.nome ?? 'novo-abandono'

  if (prevModalOpen !== isModalOpen) {
    setPrevModalOpen(isModalOpen)
    if (!isModalOpen) {
      setJogoCriadoId(null)
    }
  }

  function handleFechar() {
    setJogoCriadoId(null)
    fecharModal()
  }

  async function handleSalvar(data: AbandonarJogoFormData) {
    if (jogoEmEdicao) {
      await atualizarJogo(jogoEmEdicao.id, data)
    } else {
      let id = jogoCriadoId
      if (!id) {
        const criado = await criarJogo(data)
        id = criado.id
        setJogoCriadoId(id)
      }
      if (origemFila?.onSalvo) {
        try {
          await origemFila.onSalvo()
        } catch {
          throw new Error(ERRO_FALHA_REMOCAO_FILA)
        }
      }
    }
    handleFechar()
  }

  return (
    <Dialog open={isModalOpen} onOpenChange={(open) => !open && handleFechar()}>
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
          onCancel={handleFechar}
        />
      </DialogContent>
    </Dialog>
  )
}
