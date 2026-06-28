import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";

type GameDeleteConfirmProps = {
  jogoId: number;
  nomeJogo: string;
  onConfirm: (jogoId: number) => void;
  onCancel: () => void;
};

export function GameDeleteConfirm({ jogoId, nomeJogo, onConfirm, onCancel }: GameDeleteConfirmProps) {
  return (
    <Dialog open onOpenChange={(open) => !open && onCancel()}>
      <DialogContent className="border-border bg-surface">
        <DialogHeader>
          <DialogTitle>Excluir zeramento</DialogTitle>
          <DialogDescription>
            Tem certeza que deseja excluir "{nomeJogo}"? Esta ação não pode ser desfeita.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="ghost" onClick={onCancel}>
            Cancelar
          </Button>
          <Button type="button" variant="destructive" onClick={() => onConfirm(jogoId)}>
            Excluir
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
