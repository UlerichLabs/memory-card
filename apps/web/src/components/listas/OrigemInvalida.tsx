import { Dialog, DialogContent } from "@/components/ui/dialog";

interface Props {
  open: boolean;
  onClose: () => void;
}

export function OrigemInvalida({ open, onClose }: Props) {
  return (
    <Dialog open={open} onOpenChange={(value) => !value && onClose()}>
      <DialogContent showCloseButton={false} className="p-6 text-[var(--text-primary)]">
        <p>Origem inválida.</p>
      </DialogContent>
    </Dialog>
  );
}
