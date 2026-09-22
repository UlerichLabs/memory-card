import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

export interface ExcluirJogoDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => Promise<void> | void
  jogoNome?: string
  isLoading?: boolean
}

export function ExcluirJogoDialog({
  open,
  onOpenChange,
  onConfirm,
  jogoNome,
  isLoading = false,
}: ExcluirJogoDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-surface border-border sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-primary text-base font-bold">Excluir registro</DialogTitle>
          <DialogDescription className="text-secondary text-sm">
            {jogoNome
              ? `Tem certeza que deseja excluir o registro de "${jogoNome}"? Esta ação não pode ser desfeita.`
              : 'Tem certeza que deseja excluir este registro? Esta ação não pode ser desfeita.'}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2 sm:justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
            className="border-border text-primary hover:bg-surface-alt"
          >
            Cancelar
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={() => onConfirm()}
            disabled={isLoading}
            className="bg-[#E05A4E] text-white hover:bg-[#E05A4E]/80"
          >
            {isLoading ? 'Excluindo...' : 'Excluir'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
