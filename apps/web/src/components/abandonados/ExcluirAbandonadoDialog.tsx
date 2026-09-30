import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

export interface ExcluirAbandonadoDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => Promise<void> | void
  jogoNome?: string
  isLoading?: boolean
}

export function ExcluirAbandonadoDialog({
  open,
  onOpenChange,
  onConfirm,
  jogoNome,
  isLoading = false,
}: ExcluirAbandonadoDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="border-[var(--border)] bg-[var(--bg-surface)] sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-base font-bold text-[var(--text-primary)]">
            Excluir jogo abandonado
          </DialogTitle>
          <DialogDescription className="text-sm text-[var(--text-secondary)] leading-relaxed">
            {jogoNome
              ? `Tem certeza que deseja excluir "${jogoNome}"? Esta ação não pode ser desfeita. ` +
                'O jogo será removido da sua lista de abandonados.'
              : 'Esta ação não pode ser desfeita. O jogo será removido da sua lista de abandonados.'}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2 sm:justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
            className="border-[var(--border)] text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]"
          >
            Cancelar
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={() => onConfirm()}
            disabled={isLoading}
            className="bg-[var(--danger)] text-white hover:bg-[var(--danger)]/80"
          >
            {isLoading ? 'Excluindo...' : 'Excluir jogo'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
