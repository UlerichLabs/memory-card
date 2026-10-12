import { usePerfilStore } from '@/stores/perfilStore'
import { FotoPerfilCard } from '@/components/configuracoes/FotoPerfilCard'
import { PerfilForm } from '@/components/configuracoes/PerfilForm'

export function PerfilSecao() {
  const { isLoading, perfil } = usePerfilStore()

  if (isLoading && !perfil) {
    return (
      <div
        role="status"
        aria-label="Carregando perfil"
        className="flex min-h-[300px] items-center justify-center rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] text-[var(--text-secondary)]"
      >
        <div className="h-7 w-7 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--accent)]" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <FotoPerfilCard />
      <PerfilForm key={`${perfil?.nome}-${perfil?.avatar_url}-${perfil?.updated_at ?? 'init'}`} />
    </div>
  )
}

