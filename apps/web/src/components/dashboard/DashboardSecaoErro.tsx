export function DashboardSecaoErro({ onRetry }: { onRetry: () => void }) {
  return <div role="alert" className="rounded-[10px] border border-[var(--danger)]/50 bg-[var(--bg-surface)] p-4 text-sm text-[var(--text-secondary)]"><p>Não foi possível carregar este bloco.</p><button type="button" onClick={onRetry} className="btn-primario mt-2 px-3 py-1.5 text-xs">Tentar novamente</button></div>
}
