interface PerfilBarraAcoesProps {
  isDirty: boolean
  isSalvando: boolean
  onDescartar: () => void
  onSalvar: () => void
}

export function PerfilBarraAcoes({
  isDirty,
  isSalvando,
  onDescartar,
  onSalvar,
}: PerfilBarraAcoesProps) {
  return (
    <div className="sticky bottom-16 sm:static z-30 flex items-center justify-end gap-2.5 rounded-[12px] border border-[var(--border)] sm:border-0 bg-[var(--bg-surface)] sm:bg-transparent p-3 sm:p-0 shadow-lg sm:shadow-none">
      <button
        type="button"
        onClick={onDescartar}
        disabled={!isDirty || isSalvando}
        className="min-h-11 rounded-[7px] border border-[var(--border-subtle)] px-4 py-2 text-xs font-semibold text-[var(--text-secondary)] transition-colors hover:text-[var(--text-primary)] disabled:cursor-not-allowed disabled:opacity-40"
      >
        Descartar
      </button>
      <button
        type="button"
        onClick={onSalvar}
        disabled={!isDirty || isSalvando}
        className="min-h-11 rounded-[7px] bg-[var(--accent)] px-5 py-2 text-xs font-bold text-[#0E0F12] transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
      >
        {isSalvando ? 'Salvando…' : 'Salvar alterações'}
      </button>
    </div>
  )
}
