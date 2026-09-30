import { useEffect, useState } from "react";
import { List, Trophy, X, Loader2 } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/dialog";
import { useListasStore } from "@/stores/listasStore";
import { EscolherJogosDialog } from "./EscolherJogosDialog";
import { OrigemAutocomplete } from "./OrigemAutocomplete";
import { AJUDA_ORIGEM, ROTULOS_ORIGEM } from "./listas.constants";
import { useNovaListaForm, type ConfiguracaoDesafio } from "./useNovaListaForm";

export function NovaListaDialog() {
  const { isNovaListaOpen, fecharModalNovaLista, listaEmEdicao } = useListasStore();
  const [config, setConfig] = useState<ConfiguracaoDesafio | null>(null);
  const fecharTudo = () => {
    setConfig(null);
    fecharModalNovaLista();
  };
  const form = useNovaListaForm(fecharTudo, setConfig);
  useEffect(() => {
    if (!isNovaListaOpen) setConfig(null);
  }, [isNovaListaOpen]);
  const titulo = form.isEdicao
    ? form.tipo === "desafio"
      ? "Editar desafio"
      : "Editar lista"
    : "Nova lista ou desafio";
  const voltar = () => setConfig(null);

  return (
    <>
      <Dialog
        open={isNovaListaOpen && !config}
        onOpenChange={(open) => !open && fecharTudo()}
      >
        <DialogContent
          showCloseButton={false}
          className={["flex max-h-[calc(100dvh-48px)] w-[640px]", "max-w-[calc(100vw-32px)] flex-col gap-0",
  "overflow-hidden rounded-[16px] border", "border-[var(--modal-border)] bg-[var(--modal-bg)]",
  "p-0 text-[var(--text-primary)] sm:max-w-none"].join(" ")}
        >
          <div className="flex shrink-0 items-start justify-between p-[24px_28px_16px]">
            <div className="flex min-w-0 flex-col gap-1">
              <DialogTitle className="text-[20px] font-bold">{titulo}</DialogTitle>
              <DialogDescription className="text-[13px] text-[var(--lista-text-secondary)]">
                Uma fila do que jogar em seguida, ou uma meta pra bater.
              </DialogDescription>
            </div>
            <button
              type="button"
              aria-label="Fechar"
              onClick={fecharTudo}
              className={["flex h-10 w-10 shrink-0 items-center justify-center",
  "rounded-lg text-[var(--lista-text-muted)]", "hover:bg-[var(--bg-surface-alt)]"].join(" ")}
            >
              <X className="h-[18px] w-[18px]" />
            </button>
          </div>
          <form
            key={listaEmEdicao ? `edit-${listaEmEdicao.id}` : "nova-lista"}
            onSubmit={form.handleSubmit}
            className="flex min-h-0 flex-1 flex-col"
          >
            <div className={["custom-scrollbar flex min-h-0 flex-1 flex-col gap-[18px]",
  "overflow-x-hidden overflow-y-auto", "p-[4px_28px_20px]"].join(" ")}>
              <div
                role="radiogroup"
                aria-label="Tipo"
                className="grid min-w-0 grid-cols-1 gap-[10px] min-[420px]:grid-cols-2"
              >
                <button
                  type="button"
                  role="radio"
                  aria-checked={form.tipo === "fila"}
                  disabled={form.isEdicao}
                  onClick={() => form.setTipo("fila")}
                  className={`flex min-w-0 items-center gap-3
rounded-xl border p-[12px_14px] text-left
${form.tipo === "fila" ? "border-[var(--lista-radio-fila-border)] bg-[var(--lista-item-selected-bg)]" :
  "border-[var(--lista-radio-border)] bg-[var(--lista-radio-bg)]"}
${form.isEdicao ? "opacity-60" : ""}`}
                >
                  <span
  className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[9px] bg-[var(--hall-andamento-bg)]">
                    <List className="h-[18px] w-[18px] text-[var(--lista-icon-fila)]" />
                  </span>
                  <span className="flex min-w-0 flex-col gap-0.5">
                    <span className="truncate text-[14px] font-bold">Fila</span>
                    <span className="truncate whitespace-nowrap text-[12px] text-[var(--lista-text-secondary)]">
                      O que jogar, na sua ordem
                    </span>
                  </span>
                </button>
                <button
                  type="button"
                  role="radio"
                  aria-checked={form.tipo === "desafio"}
                  disabled={form.isEdicao}
                  onClick={() => form.setTipo("desafio")}
                  className={`flex min-w-0 items-center gap-3
rounded-xl border p-[12px_14px] text-left
${form.tipo === "desafio" ? "border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)]" :
  "border-[var(--lista-radio-border)] bg-[var(--lista-radio-bg)]"}
${form.isEdicao ? "opacity-60" : ""}`}
                >
                  <span
  className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[9px] bg-[var(--lista-pill-desafio-bg)]">
                    <Trophy className="h-[18px] w-[18px] text-[var(--hall-ouro)]" />
                  </span>
                  <span className="flex min-w-0 flex-col gap-0.5">
                    <span className="truncate text-[14px] font-bold">Desafio</span>
                    <span className="truncate whitespace-nowrap text-[12px] text-[var(--lista-text-secondary)]">
                      Jogos pra zerar, com progresso
                    </span>
                  </span>
                </button>
              </div>
              <div className="flex min-w-0 flex-col gap-1.5">
                <label
                  htmlFor="lista-nome"
                  className="text-[13px] font-medium text-[var(--lista-text-light)]"
                >
                  Nome *
                </label>
                <input
                  id="lista-nome"
                  value={form.nome}
                  onChange={(event) => form.setNome(event.target.value)}
                  className={["h-11 w-full min-w-0 rounded-[10px] border",
  "border-[var(--lista-input-border)] bg-[var(--lista-input-bg)]",
  "px-3 text-[14px] text-[var(--text-primary)] focus:outline-none"].join(" ")}
                />
                {form.errors.nome && (
                  <span className="text-xs text-[var(--danger)]">{form.errors.nome}</span>
                )}
              </div>
              {form.tipo === "desafio" && !form.isEdicao && (
                <div
                  className={["flex min-w-0 flex-col gap-[14px] rounded-xl border",
  "border-[var(--lista-desafio-box-border)] bg-[var(--lista-desafio-box-bg)] p-4"].join(" ")}
                >
                  <span className="text-[13px] font-medium">De onde vêm os jogos?</span>
                  <div
                    role="radiogroup"
                    aria-label="Origem"
                    className={["grid grid-cols-3 gap-[6px] rounded-[10px] border",
  "border-[var(--lista-desafio-box-border)] bg-[var(--modal-bg)] p-1"].join(" ")}
                  >
                    {(["franquia", "plataforma", "genero"] as const).map((tipo) => (
                      <button
                        key={tipo}
                        type="button"
                        role="radio"
                        aria-checked={form.origemTipo === tipo}
                        onClick={() => form.trocarOrigemTipo(tipo)}

                        className={`h-9 min-w-0 truncate rounded-[7px] border px-2 text-[13px] font-semibold
                          ${form.origemTipo === tipo ?
  "border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)] text-[var(--hall-ouro)]" :
  "border-transparent text-[var(--lista-text-secondary)]"}`}
                      >
                        {ROTULOS_ORIGEM[tipo]}
                      </button>
                    ))}
                  </div>
                  <OrigemAutocomplete
                    tipo={form.origemTipo}
                    valor={form.origem}
                    token={form.token}
                    erro={form.errors.origemNome}
                    onChange={form.escolherOrigem}
                  />
                  <span className="text-[12px] text-[var(--lista-text-muted)]">
                    {AJUDA_ORIGEM[form.origemTipo]}
                  </span>
                </div>
              )}
              <div className="flex min-w-0 flex-col gap-1.5">
                <div className="flex items-center justify-between">
                  <label
                    htmlFor="lista-descricao"
                    className="text-[13px] font-medium text-[var(--lista-text-light)]"
                  >
                    Descrição (opcional)
                  </label>
                  <span className="text-[11px] text-[var(--lista-text-dim)]">
                    {form.descricao.length}/200
                  </span>
                </div>
                <textarea
                  id="lista-descricao"
                  maxLength={200}
                  value={form.descricao}
                  onChange={(event) => form.setDescricao(event.target.value)}
                  className={[
                    "h-[68px] w-full min-w-0 resize-none rounded-[10px] border",
                    "border-[var(--lista-input-border)] bg-[var(--lista-input-bg)]",
                    "p-[10px_12px] text-[14px] text-[var(--text-primary)] focus:outline-none",
                  ].join(" ")}
                />
                {form.errors.descricao && (
                  <span className="text-xs text-[var(--danger)]">{form.errors.descricao}</span>
                )}
              </div>
              {form.errors.form && (
                <span className="text-xs text-[var(--danger)]">{form.errors.form}</span>
              )}
            </div>
            <div
              className="flex shrink-0 items-center justify-between border-t border-[var(--modal-border)]
                p-[16px_28px_20px]"
            >
              <span className="text-[12px] text-[var(--lista-text-dim)]">
                {form.tipo === "desafio" && !form.isEdicao ? "Passo 1 de 2" : ""}
              </span>
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={fecharTudo}
                  className="h-11 rounded-[10px] border border-[var(--lista-btn-icon-border)] px-[18px]
                    text-[14px] font-medium"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={form.isSubmitting}
                  className={`flex h-11 items-center justify-center gap-2 rounded-[10px] px-5 text-[14px] font-bold
${form.tipo === "desafio" && !form.isEdicao ? "bg-[var(--hall-ouro)] text-[var(--ouro-jogo-ano-text)]" :
  "bg-[var(--lista-btn-primary-bg)] text-white"}`}
                >
                  {form.isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
                  {form.isEdicao ? "Salvar" : form.tipo === "desafio" ? "Ver jogos" : "Criar fila"}
                </button>
              </div>
            </div>
          </form>
        </DialogContent>
      </Dialog>
      {config && (
        <EscolherJogosDialog
          open={isNovaListaOpen && Boolean(config)}
          modo="criar"
          config={config}
          onClose={fecharTudo}
          onBack={voltar}
        />
      )}
    </>
  );
}
