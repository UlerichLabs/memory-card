import { Crown, Loader2 } from 'lucide-react'

export interface EscolherGameDoAnoFooterProps {
  modoTrocar: boolean
  isSubmitting: boolean
  desabilitarConfirmar: boolean
  onCancelar: () => void
  onConfirmar: () => void
  onRemover: () => void
}

export function EscolherGameDoAnoFooter({
  modoTrocar,
  isSubmitting,
  desabilitarConfirmar,
  onCancelar,
  onConfirmar,
  onRemover,
}: EscolherGameDoAnoFooterProps) {
  return (
    <div className="flex items-center justify-between border-t border-[var(--hall-modal-border)] pt-[16px]">
      <div>
        {modoTrocar && (
          <button
            type="button"
            onClick={onRemover}
            disabled={isSubmitting}
            className="h-[44px] px-2 text-[14px] font-medium text-[var(--hall-btn-remover-text)] transition-opacity hover:opacity-80 disabled:opacity-50"
          >
            Remover Game do Ano
          </button>
        )}
      </div>

      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={onCancelar}
          disabled={isSubmitting}
          className="flex h-[44px] items-center rounded-[8px] border border-[var(--hall-modal-btn-cancelar-border)] px-4 text-[14px] font-medium text-[var(--text-secondary)] transition-colors hover:bg-[var(--bg-surface-alt)] disabled:opacity-50"
        >
          Cancelar
        </button>
        <button
          type="button"
          onClick={onConfirmar}
          disabled={desabilitarConfirmar}
          className="flex h-[44px] items-center gap-2 rounded-[8px] bg-[var(--hall-ouro)] px-5 text-[14px] font-bold text-[var(--hall-dark-icon)] transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isSubmitting ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Crown className="h-4 w-4 fill-current" />
          )}
          <span>Definir como Game do Ano</span>
        </button>
      </div>
    </div>
  )
}
