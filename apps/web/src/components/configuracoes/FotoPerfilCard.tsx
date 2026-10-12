import { useState } from 'react'
import { UserAvatar } from '@/components/layout/UserAvatar'
import { usePerfilStore } from '@/stores/perfilStore'
import { FotoPerfilDialog } from './FotoPerfilDialog'

export function FotoPerfilCard() {
  const { perfil, isSalvando, removerFoto } = usePerfilStore()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogTab, setDialogTab] = useState<'upload' | 'capa'>('upload')

  const temFoto = Boolean((perfil?.avatar_url || perfil?.foto_url) && perfil?.avatar_tipo !== 'iniciais' && perfil?.avatar_tipo !== null)

  function abrirUpload() {
    setDialogTab('upload')
    setDialogOpen(true)
  }

  function abrirCapa() {
    setDialogTab('capa')
    setDialogOpen(true)
  }

  async function handleRemover() {
    try {
      await removerFoto()
    } catch {}
  }

  return (
    <section className="space-y-4 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <header>
        <h3 className="text-sm font-bold text-[var(--text-primary)]">Foto de perfil</h3>
      </header>

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:gap-6">
        <UserAvatar nome={perfil?.nome ?? 'Jogador'} src={perfil?.avatar_url ?? perfil?.foto_url} size="xl" />

        <div className="flex flex-col gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <button
              type="button"
              onClick={abrirUpload}
              disabled={isSalvando}
              className="min-h-11 rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3.5 py-2 text-xs font-semibold text-[var(--text-primary)] transition-colors hover:border-[var(--accent)] hover:text-[var(--accent)] disabled:opacity-50"
            >
              Enviar foto
            </button>
            <button
              type="button"
              onClick={abrirCapa}
              disabled={isSalvando}
              className="min-h-11 rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3.5 py-2 text-xs font-semibold text-[var(--text-primary)] transition-colors hover:border-[var(--accent)] hover:text-[var(--accent)] disabled:opacity-50"
            >
              Usar capa de um jogo
            </button>
            {temFoto && (
              <button
                type="button"
                onClick={handleRemover}
                disabled={isSalvando}
                className="min-h-11 rounded-[7px] border border-[var(--border-subtle)] px-3.5 py-2 text-xs font-semibold text-[var(--danger)] transition-colors hover:bg-[var(--danger)]/10 disabled:opacity-50"
              >
                Remover
              </button>
            )}
          </div>

          <p className="text-xs text-[var(--text-muted)]">
            JPG, PNG ou WebP, até 2 MB. Sem foto, mostramos suas iniciais.
          </p>
        </div>
      </div>

      <FotoPerfilDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        initialTab={dialogTab}
      />
    </section>
  )
}
