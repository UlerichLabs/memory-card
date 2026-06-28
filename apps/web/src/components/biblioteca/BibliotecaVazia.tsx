import { Gamepad2 } from "lucide-react";

import { Button } from "@/components/ui/button";

type BibliotecaVaziaProps = {
  onRegistrar: () => void;
};

export function BibliotecaVazia({ onRegistrar }: BibliotecaVaziaProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-12 text-center">
      <Gamepad2 className="h-12 w-12 text-muted" />
      <p className="text-lg font-medium">Nenhum zeramento registrado ainda</p>
      <Button type="button" onClick={onRegistrar}>
        Registrar primeiro jogo
      </Button>
    </div>
  );
}
