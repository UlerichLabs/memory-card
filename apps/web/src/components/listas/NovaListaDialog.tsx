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
      <DialogContent showCloseButton={false} className="flex max-h-[calc(100vh-32px)] w-[640px] max-w-[calc(100vw-32px)] flex-col gap-[18px] overflow-y-auto rounded-[16px] border border-[var(--modal-border)] bg-[var(--modal-bg)] p-6 sm:p-[28px_32px]">
        <div className="flex items-start justify-between">
          <div className="flex flex-col gap-1">
            <DialogTitle className="text-[22px] font-bold text-[var(--text-primary)]">{titulo}</DialogTitle>
            <DialogDescription className="text-[14px] text-[var(--lista-text-secondary)]">Uma fila do que jogar em seguida, ou uma meta pra bater.</DialogDescription>
          </div>
          <button type="button" aria-label="Fechar" onClick={fecharModalNovaLista} className="flex h-11 w-11 items-center justify-center rounded-lg text-[var(--lista-text-muted)] hover:bg-[var(--bg-surface-alt)] hover:text-[var(--text-primary)]">
            <X className="h-5 w-5" />
          </button>
        </div>

        <form
          key={listaEmEdicao ? `edit-${listaEmEdicao.id}` : 'nova-lista'}
          onSubmit={handleSubmit}
          className="flex flex-col gap-4"
        >
          <div role="radiogroup" aria-label="Tipo" className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <button
              type="button"
              role="radio"
              aria-checked={tipo === 'fila'}
              disabled={isEdicao}
              onClick={() => setTipo('fila')}
              className={`flex flex-col gap-2 rounded-xl p-4 text-left transition-colors ${
                tipo === 'fila'
                  ? 'border border-[var(--lista-radio-fila-border)] bg-[var(--lista-item-selected-bg)]'
                  : 'border border-[var(--lista-radio-border)] bg-[var(--lista-radio-bg)] hover:border-[var(--lista-text-dim)]'
              } ${isEdicao ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}`}
            >
              <List className="h-[22px] w-[22px] text-[var(--lista-icon-fila)]" />
              <span className="text-[15px] font-bold text-[var(--text-primary)]">Fila</span>
              <span className="text-[12px] text-[var(--lista-text-secondary)]">Jogos que você quer jogar, na ordem que quiser.</span>
            </button>

            <button
              type="button"
              role="radio"
              aria-checked={tipo === 'desafio'}
              disabled={isEdicao}
              onClick={() => setTipo('desafio')}
              className={`flex flex-col gap-2 rounded-xl p-4 text-left transition-colors ${
                tipo === 'desafio'
                  ? 'border border-[var(--hall-ouro)] bg-[var(--lista-pill-desafio-bg)]'
                  : 'border border-[var(--lista-radio-border)] bg-[var(--lista-radio-bg)] hover:border-[var(--lista-text-dim)]'
              } ${isEdicao ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}`}
            >
              <Trophy className="h-[22px] w-[22px] text-[var(--hall-ouro)]" />
              <span className="text-[15px] font-bold text-[var(--text-primary)]">Desafio</span>
              <span className="text-[12px] text-[var(--lista-text-secondary)]">Uma meta com progresso automático e conquista no final.</span>
            </button>
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="lista-nome" className="text-[13px] font-medium text-[var(--lista-text-light)]">Nome *</label>
            <input
              id="lista-nome"
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              placeholder={tipo === 'fila' ? 'Ex: Próximos do backlog' : 'Ex: Zerar Franquia Mario'}
              className="h-11 rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] px-3 text-[14px] text-[var(--text-primary)] focus:outline-none"
            />
            {errors.nome && <span className="text-xs text-[var(--danger)]">{errors.nome}</span>}
          </div>

          {tipo === 'desafio' && (
            <NovaListaRegraSection
              regraTipo={regraTipo}
              setRegraTipo={setRegraTipo}
              regraValor={regraValor}
              setRegraValor={setRegraValor}
              setIgdbId={setIgdbId}
              meta={meta}
              setMeta={setMeta}
              errors={errors}
              franquias={franquias}
              setFranquias={setFranquias}
              isSearchingFranquias={isSearchingFranquias}
              plataformas={plataformas}
              generos={generos}
              buscarFranquias={buscarFranquias}
              disabled={isEdicao}
            />
          )}

          <div className="flex flex-col gap-1.5">
            <div className="flex items-center justify-between">
              <label htmlFor="lista-descricao" className="text-[13px] font-medium text-[var(--lista-text-light)]">Descrição (opcional)</label>
              <span className="text-[11px] text-[var(--lista-text-dim)]">{descricao.length}/200</span>
            </div>
            <textarea
              id="lista-descricao"
              maxLength={200}
              value={descricao}
              onChange={(e) => setDescricao(e.target.value)}
              placeholder="Adicione um contexto ou anotação para a lista"
              className="min-h-[64px] resize-none rounded-[10px] border border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] p-3 text-[14px] text-[var(--text-primary)] focus:outline-none"
            />
            {errors.descricao && <span className="text-xs text-[var(--danger)]">{errors.descricao}</span>}
          </div>

          {errors.form && <span className="text-xs text-[var(--danger)]">{errors.form}</span>}

          <div className="flex items-center justify-end gap-3 border-t border-[var(--modal-border)] pt-4">
            <button
              type="button"
              onClick={fecharModalNovaLista}
              className="h-11 rounded-[10px] border border-[var(--lista-btn-icon-border)] px-5 text-[14px] font-medium text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className={`flex h-11 items-center justify-center gap-2 rounded-[10px] px-[22px] text-[14px] font-bold transition-opacity ${
                tipo === 'desafio'
                  ? 'bg-[var(--hall-ouro)] text-[var(--ouro-jogo-ano-text)] hover:opacity-90'
                  : 'bg-[var(--lista-btn-primary-bg)] text-white hover:opacity-90'
              } ${isSubmitting ? 'cursor-not-allowed opacity-60' : ''}`}
            >
              {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
              {isEdicao ? 'Salvar' : (tipo === 'desafio' ? 'Criar desafio' : 'Criar fila')}
            </button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
