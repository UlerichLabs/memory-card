import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { useListasStore } from "@/stores/listasStore";
import { EscolherJogosBarra } from "./EscolherJogosBarra";
import { EscolherJogosCabecalho } from "./EscolherJogosCabecalho";
import { EscolherJogosGrade } from "./EscolherJogosGrade";
import { EscolherJogosRodape } from "./EscolherJogosRodape";
import { useEscolherJogos } from "./useEscolherJogos";
import { useEnviarSelecao } from "./useEnviarSelecao";
import { OrigemInvalida } from "./OrigemInvalida";
import type { ConteudoEscolhaProps, EscolherJogosDialogProps } from "./EscolherJogosDialog.types";

export function EscolherJogosDialog({
  open,
  modo,
  config,
  lista,
  onClose,
  onBack,
  onAdded,
}: EscolherJogosDialogProps) {
  const navigate = useNavigate();
  const store = useListasStore();
  const origem = config?.origem ?? lista?.origem;
  if (!origem) return null;
  if (origem.igdb_id === null) return <OrigemInvalida open={open} onClose={onClose} />;
  return (
    <ConteudoEscolha
      origem={origem}
      modo={modo}
      config={config}
      lista={lista}
      open={open}
      onClose={onClose}
      onBack={onBack}
      onAdded={onAdded}
      navigate={navigate}
      store={store}
    />
  );
}

function ConteudoEscolha({
  origem,
  modo,
  config,
  lista,
  open,
  onClose,
  onBack,
  onAdded,
  navigate,
  store,
}: ConteudoEscolhaProps) {
  const jogo = useEscolherJogos({ origem, existentes: lista?.itens, modo });
  const [revisando, setRevisando] = useState(false);
  const [marcando, setMarcando] = useState(false);
  const selecionados = jogo.payload;
  const meta =
    modo === "criar" ? selecionados.length : (lista?.itens.length ?? 0) + selecionados.length;
  const marcarVisiveis = () =>
    jogo.itens.forEach(
      (item) =>
        !jogo.idsExistentes.has(item.igdb_id) &&
        !jogo.selecionados.has(item.igdb_id) &&
        jogo.alternar(item),
    );
  const envio = useEnviarSelecao({
    modo,
    config,
    itens: selecionados,
    store,
    navigate,
    onClose,
    onAdded,
  });
  return (
    <Dialog open={open} onOpenChange={(value) => !value && onClose()}>
      <DialogContent
        showCloseButton={false}
        className={[
          "flex h-auto max-h-[calc(100dvh-48px)] w-[1040px]",
          "max-w-[calc(100vw-32px)] flex-col gap-0 overflow-hidden rounded-[16px]",
          "border border-[var(--modal-border)] bg-[var(--modal-bg)] p-0 text-[var(--text-primary)]",
          "max-md:h-dvh max-md:w-screen max-md:max-w-none max-md:rounded-none sm:max-w-none",
        ].join(" ")}
      >
        <EscolherJogosCabecalho
          origem={origem}
          nome={config?.nome ?? lista?.nome}
          busca={jogo.busca}
          generoId={jogo.generoId}
          plataformaId={jogo.plataformaId}
          ordenar={jogo.ordenar}
          token={jogo.token}
          onBusca={jogo.setBusca}
          onGenero={jogo.setGeneroId}
          onPlataforma={jogo.setPlataformaId}
          onOrdenar={jogo.setOrdenar}
          onClose={onClose}
        />
        <EscolherJogosBarra
          carregados={jogo.itens.length}
          total={jogo.meta.total}
          totalSugeridos={jogo.meta.total_sugeridos}
          totalTodos={jogo.meta.total_todos}
          mostrarSugeridos={origem.tipo === "franquia"}
          somenteSugeridos={jogo.somenteSugeridos}
          marcando={marcando}
          onMarcarSugeridos={() => {
            setMarcando(true);
            void jogo.marcarSugeridos().finally(() => setMarcando(false));
          }}
          onMarcarVisiveis={marcarVisiveis}
          onAlternarSugeridos={() => jogo.setSomenteSugeridos(!jogo.somenteSugeridos)}
        />
        <EscolherJogosGrade
          itens={jogo.itens}
          selecionados={jogo.selecionados}
          idsExistentes={jogo.idsExistentes}
          carregando={jogo.carregando}
          erro={jogo.erro}
          revisando={revisando}
          total={jogo.meta.total}
          porPagina={jogo.meta.por_pagina}
          mostrarRelacionados={origem.tipo === "franquia" && jogo.somenteSugeridos === false}
          onToggle={jogo.alternar}
          onDesmarcar={jogo.desmarcar}
          onRecarregar={jogo.recarregar}
          onCarregarMais={jogo.carregarMais}
        />
        <EscolherJogosRodape
          modo={modo}
          selecionados={selecionados.length}
          meta={meta}
          enviando={envio.enviando}
          erro={envio.erro}
          onRevisar={() => setRevisando((value) => !value)}
          onVoltar={modo === "criar" ? (onBack ?? onClose) : onClose}
          onEnviar={envio.enviar}
        />
      </DialogContent>
    </Dialog>
  );
}
