import { Crown, X, Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogTitle, DialogDescription, DialogClose } from '@/components/ui/dialog'
import type { JogoZeradoDTO } from '@/types/jogos'
import { GameDoAnoRadioItem } from './GameDoAnoRadioItem'
import { EscolherGameDoAnoFooter } from './EscolherGameDoAnoFooter'
import { useEscolherGameDoAno } from './useEscolherGameDoAno'

export interface EscolherGameDoAnoDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  ano: number
  jogoAtual: JogoZeradoDTO | null
  onSuccess?: () => void
}

export function EscolherGameDoAnoDialog({
  open,
  onOpenChange,
  ano,
  jogoAtual,
  onSuccess,
}: EscolherGameDoAnoDialogProps) {
  const {
    jogos,
    selectedId,
    setSelectedId,
    isLoading,
    isSubmitting,
    errorMessage,
    itemRefs,
    handleKeyDown,
    handleConfirmar,
    handleRemover,
    modoTrocar,
    desabilitarConfirmar,
    mostrarAvisoTroca,
  } = useEscolherGameDoAno({
    open,
    ano,
    jogoAtual,
    onSuccess,
    onClose: () => onOpenChange(false),
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className="flex h-full w-full max-w-full flex-col justify-between gap-[18px] overflow-hidden rounded-none border border-[var(--hall-modal-border)] bg-[var(--hall-modal-bg)] p-6 sm:h-auto sm:max-h-[90vh] sm:max-w-[720px] sm:rounded-[16px] sm:p-[28px_32px]"
      >
        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Crown className="h-5 w-5 fill-current text-[var(--hall-ouro)]" aria-hidden="true" />
              <DialogTitle className="text-[22px] font-bold text-[var(--text-primary)]">
                Game do Ano de {ano}
              </DialogTitle>
            </div>
            <DialogClose
              aria-label="Fechar"
              className="flex h-11 w-11 items-center justify-center rounded-[8px] text-[var(--hall-subtitulo)] transition-colors hover:text-[var(--text-primary)]"
            >
              <X className="h-5 w-5" />
            </DialogClose>
          </div>
          <DialogDescription className="text-[14px] text-[var(--hall-subtitulo)]">
            Escolha o jogo que mais marcou o seu {ano}. Só um por ano.
          </DialogDescription>
        </div>

        <div
          role="radiogroup"
          aria-label={`Jogos de ${ano}`}
          className="custom-scrollbar flex max-h-[420px] min-h-[160px] flex-1 flex-col gap-2 overflow-y-auto pr-1"
        >
          {isLoading ? (
            <div className="flex min-h-[160px] items-center justify-center text-[var(--hall-muted)]">
              <Loader2 className="h-6 w-6 animate-spin text-[var(--hall-ouro)]" />
            </div>
          ) : jogos.length === 0 ? (
            <div className="flex min-h-[160px] items-center justify-center text-center text-[13px] text-[var(--hall-muted)]">
              Nenhum jogo encontrado para o ano {ano}.
            </div>
          ) : (
            jogos.map((j, idx) => (
              <GameDoAnoRadioItem
                key={j.id}
                jogo={j}
                selecionado={selectedId === j.id}
                tabIndex={selectedId === j.id || (selectedId === null && idx === 0) ? 0 : -1}
                onSelect={() => setSelectedId(j.id)}
                onKeyDown={(e) => handleKeyDown(e, idx)}
                itemRef={(el) => { if (el) itemRefs.current.set(j.id, el) }}
              />
            ))
          )}
        </div>

        {mostrarAvisoTroca && jogoAtual && (
          <div className="rounded-[10px] border border-[var(--hall-aviso-border)] bg-[var(--hall-aviso-bg)] p-[10px_12px] text-[13px] text-[var(--hall-aviso-text)]">
            <strong>{jogoAtual.nome}</strong> deixa de ser o Game do Ano de {ano}.
          </div>
        )}

        {errorMessage && (
          <div className="rounded-[8px] border border-[var(--danger)]/50 bg-[var(--danger)]/10 px-3 py-2 text-[13px] text-[var(--danger)]">
            {errorMessage}
          </div>
        )}

        <EscolherGameDoAnoFooter
          modoTrocar={modoTrocar}
          isSubmitting={isSubmitting}
          desabilitarConfirmar={desabilitarConfirmar}
          onCancelar={() => onOpenChange(false)}
          onConfirmar={handleConfirmar}
          onRemover={handleRemover}
        />
      </DialogContent>
    </Dialog>
  )
}
