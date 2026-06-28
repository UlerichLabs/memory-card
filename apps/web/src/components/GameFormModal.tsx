import { useEffect, useRef, useState } from "react";
import type { IgdbJogo } from "@memory-card/types";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, ImageOff, Search } from "lucide-react";

import { GameForm } from "@/components/GameForm";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { buscarJogo, buscarJogosIgdb } from "@/lib/api";
import { useGameFormStore } from "@/store/gameFormStore";

const FORM_ID = "game-form-modal";

export function GameFormModal() {
  const queryClient = useQueryClient();
  const { isOpen, modo, jogoId, fechar } = useGameFormStore();
  const [etapa, setEtapa] = useState<1 | 2>(1);
  const [busca, setBusca] = useState("");
  const [resultados, setResultados] = useState<IgdbJogo[]>([]);
  const [buscando, setBuscando] = useState(false);
  const [jogoSelecionado, setJogoSelecionado] = useState<IgdbJogo | null>(null);
  const buscaRef = useRef<HTMLInputElement>(null);
  const isEdit = modo === "editar";
  const { data: jogo, isLoading } = useQuery({
    queryKey: ["jogo", jogoId],
    queryFn: () => buscarJogo(jogoId ?? 0),
    enabled: isOpen && isEdit && jogoId !== null
  });
  const titulo = isEdit ? jogo?.nome ?? "Editar zeramento" : jogoSelecionado?.nome ?? "Novo zeramento";

  useEffect(() => {
    if (!isOpen) {
      setEtapa(1);
      setBusca("");
      setResultados([]);
      setJogoSelecionado(null);
      return;
    }

    setEtapa(isEdit ? 2 : 1);
  }, [isEdit, isOpen]);

  useEffect(() => {
    if (!isOpen || isEdit || etapa !== 1 || busca.trim().length < 3) {
      setResultados([]);
      setBuscando(false);
      return;
    }

    let active = true;
    const timeout = window.setTimeout(async () => {
      setBuscando(true);

      try {
        const jogos = await buscarJogosIgdb(busca);

        if (active) {
          setResultados(jogos);
        }
      } finally {
        if (active) {
          setBuscando(false);
        }
      }
    }, 400);

    return () => {
      active = false;
      window.clearTimeout(timeout);
    };
  }, [busca, etapa, isEdit, isOpen]);

  useEffect(() => {
    if (isOpen && etapa === 1 && !isEdit) {
      window.setTimeout(() => buscaRef.current?.focus(), 0);
    }
  }, [etapa, isEdit, isOpen]);

  async function handleSuccess() {
    await queryClient.invalidateQueries({ queryKey: ["jogos"] });
    await queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    if (jogoId) {
      await queryClient.invalidateQueries({ queryKey: ["jogo", jogoId] });
    }
    fechar();
  }

  function selecionarJogo(jogo: IgdbJogo) {
    setJogoSelecionado(jogo);
    setBusca(jogo.nome);
    setEtapa(2);
  }

  function pularBusca() {
    setJogoSelecionado(null);
    setEtapa(2);
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && fechar()}>
      <DialogContent
        overlayClassName="bg-black/70"
        className={etapa === 1 && !isEdit ? "max-w-[480px] gap-0 overflow-hidden rounded-xl border-border bg-surface p-0" : "max-h-[92vh] max-w-[960px] gap-0 overflow-hidden rounded-xl border-border bg-surface p-0"}
      >
        <DialogHeader className="border-b border-border px-6 py-4">
          <DialogTitle className="flex items-center gap-2">
            {etapa === 2 && !isEdit && (
              <button type="button" onClick={() => setEtapa(1)} className="rounded-sm text-muted hover:text-foreground">
                <ArrowLeft className="h-4 w-4" />
              </button>
            )}
            <span>{etapa === 1 && !isEdit ? "Novo zeramento" : titulo}</span>
          </DialogTitle>
        </DialogHeader>
        {etapa === 1 && !isEdit ? (
          <div className="px-10 py-10">
            <div className="mx-auto max-w-sm space-y-4">
              <h2 className="text-center text-lg font-medium text-foreground">Qual jogo você zerou?</h2>
              <div className="relative">
                <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted" />
                <Input
                  ref={buscaRef}
                  value={busca}
                  onChange={(event) => setBusca(event.target.value)}
                  placeholder="Buscar jogo"
                  className="bg-surface-raised pl-9"
                />
              </div>
              {(busca.trim().length >= 3 || resultados.length > 0) && (
                <div className="overflow-hidden rounded-md border border-border bg-surface">
                  {buscando ? (
                    <p className="p-3 text-xs text-muted">Buscando...</p>
                  ) : resultados.length === 0 ? (
                    <p className="p-3 text-xs text-muted">Nenhum resultado</p>
                  ) : (
                    resultados.map((jogo) => (
                      <button
                        type="button"
                        key={jogo.igdbId}
                        onClick={() => selecionarJogo(jogo)}
                        className="flex w-full items-center gap-3 p-2 text-left hover:bg-surface-raised"
                      >
                        {jogo.capaUrl ? (
                          <img src={jogo.capaUrl} alt="" className="h-10 w-10 rounded object-cover" />
                        ) : (
                          <span className="flex h-10 w-10 items-center justify-center rounded bg-surface-raised text-muted">
                            <ImageOff className="h-4 w-4" />
                          </span>
                        )}
                        <span className="min-w-0 flex-1">
                          <span className="block truncate text-sm text-foreground">{jogo.nome}</span>
                          <span className="block text-xs text-muted">{jogo.ano ?? "—"}</span>
                        </span>
                      </button>
                    ))
                  )}
                </div>
              )}
              <button type="button" onClick={pularBusca} className="mx-auto block text-center text-xs text-muted hover:text-foreground">
                Pular busca →
              </button>
            </div>
          </div>
        ) : (
          <>
            <div className="max-h-[calc(92vh-129px)] overflow-hidden">
              {isEdit && (isLoading || !jogo) ? (
                <div className="h-96 bg-surface" />
              ) : (
                <GameForm
                  key={`${modo}-${jogoId ?? jogoSelecionado?.igdbId ?? "novo"}`}
                  formId={FORM_ID}
                  hideFooter
                  compact
                  desativarAutocomplete
                  ocultarNome={!isEdit && Boolean(jogoSelecionado)}
                  jogoSelecionadoInicial={!isEdit ? jogoSelecionado : undefined}
                  jogoInicial={isEdit ? jogo : undefined}
                  onSuccess={() => void handleSuccess()}
                />
              )}
            </div>
            <DialogFooter className="border-t border-border bg-surface px-4 py-3">
              <Button type="button" variant="secondary" onClick={fechar}>
                Cancelar
              </Button>
              <Button type="submit" form={FORM_ID}>
                Salvar zeramento
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
