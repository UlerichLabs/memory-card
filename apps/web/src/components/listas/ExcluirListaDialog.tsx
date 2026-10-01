import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useListasStore } from '@/stores/listasStore'

export function ExcluirListaDialog() {
  const navigate = useNavigate()
  const { isExcluirListaOpen, listaParaExcluir, fecharModalExcluir, excluirLista } = useListasStore()
  const [isLoading, setIsLoading] = useState(false)

  async function handleConfirm() {
    if (!listaParaExcluir) return
    setIsLoading(true)
    try {
      await excluirLista(listaParaExcluir.id)
      fecharModalExcluir()
      navigate('/listas')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={isExcluirListaOpen} onOpenChange={(open) => !open && fecharModalExcluir()}>
      <DialogContent className="border-border bg-surface sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-primary text-base font-bold">Excluir lista</DialogTitle>
          <DialogDescription className="text-secondary text-sm">
            {listaParaExcluir
              ? `Excluir "${listaParaExcluir.nome}"? Os jogos da sua Biblioteca não são afetados.`
              : 'Tem certeza que deseja excluir esta lista? Os jogos da sua Biblioteca não são afetados.'}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2 sm:justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={fecharModalExcluir}
            disabled={isLoading}
            className="border-border text-primary hover:bg-surface-alt"
          >
            Cancelar
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={handleConfirm}
            disabled={isLoading}
            className="bg-[var(--danger)] text-white hover:bg-[var(--danger)]/80"
          >
            {isLoading ? 'Excluindo...' : 'Excluir'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
