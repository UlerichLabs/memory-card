import { List, Trophy, X, Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { useListasStore } from '@/stores/listasStore'
import { useNovaListaForm } from './useNovaListaForm'
import { NovaListaRegraSection } from './NovaListaRegraSection'

export function NovaListaDialog() {
  const { isNovaListaOpen, fecharModalNovaLista, listaEmEdicao } = useListasStore()
  const form = useNovaListaForm(fecharModalNovaLista)
  const {
    tipo, setTipo, nome, setNome, descricao, setDescricao,
    regraTipo, setRegraTipo, regraValor, setRegraValor, setIgdbId,
    meta, setMeta, errors, isSubmitting, franquias, setFranquias,
    isSearchingFranquias, plataformas, generos, buscarFranquias,
    handleSubmit, isEdicao,
  } = form
  const titulo = isEdicao ? (tipo === 'desafio' ? 'Editar desafio' : 'Editar lista') : 'Nova lista ou desafio'

  return (
    <Dialog open={isNovaListaOpen} onOpenChange={(open) => !open && fecharModalNovaLista()}>
      <DialogContent showCloseButton={false} className="flex max-h-[calc(100dvh-48px)] w-[640px] max-w-[calc(100vw-32px)] flex-col gap-0 overflow-hidden rounded-[16px] border border-[var(--modal-border)] bg-[var(--modal-bg)] p-0 text-[var(--text-primary)] sm:max-w-none">
        <div className="flex shrink-0 items-start justify-between p-[24px_28px_16px]">
          <div className="flex min-w-0 flex-col gap-1">
            <DialogTitle className="text-[20px] font-bold text-[var(--text-primary)]">{titulo}</DialogTitle>
            <DialogDescription className="text-[13px] text-[var(--lista-text-secondary)]">Uma fila do que jogar em seguida, ou uma meta pra bater.</DialogDescription>
          </div>
          <button type="button" aria-label="Fechar" onClick={fecharModalNovaLista} className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-[var(--lista-text-muted)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)]">
            <X className="h-[18px] w-[18px]" />
          </button>
        </div>

        <form key={listaEmEdicao ? `edit-${listaEmEdicao.id}` : 'nova-lista'} onSubmit={handleSubmit} className="flex min-h-0 flex-1 flex-col">
          <div className="custom-scrollbar flex min-h-0 flex-1 flex-col gap-[18px] overflow-x-hidden overflow-y-auto p-[4px_28px_20px]">
            <div role="radiogroup" aria-label="Tipo" className="grid min-w-0 grid-cols-1 gap-[10px] min-[420px]:grid-cols-2">
              <button type="button" role="radio" aria-checked={tipo === 'fila'} disabled={isEdicao} onClick={() => setTipo('fila')} className={`flex min-w-0 items-center gap-3 rounded-xl border p-[12px_14px] text-left transition-colors ${tipo === 'fila' ? 'border-[var(--lista-radio-fila-border)] bg-[var(--lista-item-selected-bg)]' : 'border-[var(--lista-radio-border)] bg-[var(--lista-radio-bg)] hover:border-[var(--lista-text-dim)]'} ${isEdicao ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}`}>
                <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[9px] bg-[var(--hall-andamento-bg)]"><List className="h-[18px] w-[18px] text-[var(--lista-icon-fila)]" /></span>
                <span className="flex min-w-0 flex-col gap-0.5"><span className="truncate text-[14px] font-bold text-[var(--text-primary)]">Fila</span><span className="truncate whitespace-nowrap text-[12px] text-[var(--lista-text-secondary)]">O que jogar, na sua ordem</span></span>
              </button>
              <button type="button" role="radio" aria-checked={tipo === 'desafio'} disabled={isEdicao} onClick={() => setTipo('desafio')} className={`flex min-w-0 items-center gap-3 rounded-xl border p-[12px_14px] text-left transition-colors ${tipo === 'desafio' ? 'border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)]' : 'border-[var(--lista-radio-border)] bg-[var(--lista-radio-bg)] hover:border-[var(--lista-text-dim)]'} ${isEdicao ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}`}>
                <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[9px] bg-[var(--lista-pill-desafio-bg)]"><Trophy className="h-[18px] w-[18px] text-[var(--hall-ouro)]" /></span>
                <span className="flex min-w-0 flex-col gap-0.5"><span className="truncate text-[14px] font-bold text-[var(--text-primary)]">Desafio</span><span className="truncate whitespace-nowrap text-[12px] text-[var(--lista-text-secondary)]">Meta com progresso automático</span></span>
              </button>
            </div>

            <div className="flex min-w-0 flex-col gap-1.5">
              <label htmlFor="lista-nome" className="text-[13px] font-medium text-[var(--lista-text-light)]">Nome *</label>
              <input id="lista-nome" value={nome} onChange={(e) => setNome(e.target.value)} placeholder={tipo === 'fila' ? 'Ex: Próximos do backlog' : 'Ex: Zerar Franquia Mario'} className="h-11 w-full min-w-0 rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] px-3 text-[14px] text-[var(--text-primary)] focus:outline-none" />
              {errors.nome && <span className="text-xs text-[var(--danger)]">{errors.nome}</span>}
            </div>

            {tipo === 'desafio' && <NovaListaRegraSection regraTipo={regraTipo} setRegraTipo={setRegraTipo} regraValor={regraValor} setRegraValor={setRegraValor} setIgdbId={setIgdbId} meta={meta} setMeta={setMeta} errors={errors} franquias={franquias} setFranquias={setFranquias} isSearchingFranquias={isSearchingFranquias} plataformas={plataformas} generos={generos} buscarFranquias={buscarFranquias} disabled={isEdicao} />}

            <div className="flex min-w-0 flex-col gap-1.5">
              <div className="flex items-center justify-between"><label htmlFor="lista-descricao" className="text-[13px] font-medium text-[var(--lista-text-light)]">Descrição (opcional)</label><span className="text-[11px] text-[var(--lista-text-dim)]">{descricao.length}/200</span></div>
              <textarea id="lista-descricao" maxLength={200} value={descricao} onChange={(e) => setDescricao(e.target.value)} placeholder="Adicione um contexto ou anotação para a lista" className="h-[68px] w-full min-w-0 resize-none rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] p-[10px_12px] text-[14px] text-[var(--text-primary)] focus:outline-none" />
              {errors.descricao && <span className="text-xs text-[var(--danger)]">{errors.descricao}</span>}
            </div>
            {errors.form && <span className="text-xs text-[var(--danger)]">{errors.form}</span>}
          </div>

          <div className="flex shrink-0 items-center justify-end gap-3 border-t border-[var(--modal-border)] p-[16px_28px_20px]">
            <button type="button" onClick={fecharModalNovaLista} className="h-11 rounded-[10px] border border-[var(--lista-btn-icon-border)] px-[18px] text-[14px] font-medium text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]">Cancelar</button>
            <button type="submit" disabled={isSubmitting} className={`flex h-11 items-center justify-center gap-2 rounded-[10px] px-5 text-[14px] font-bold transition-opacity ${tipo === 'desafio' ? 'bg-[var(--hall-ouro)] text-[var(--ouro-jogo-ano-text)] hover:opacity-90' : 'bg-[var(--lista-btn-primary-bg)] text-white hover:opacity-90'} ${isSubmitting ? 'cursor-not-allowed opacity-60' : ''}`}>
              {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
              {isEdicao ? 'Salvar' : (tipo === 'desafio' ? 'Criar desafio' : 'Criar fila')}
            </button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
