import { useEffect, useState } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { usePerfilStore } from '@/stores/perfilStore'
import { FotoUploadTab } from './FotoUploadTab'
import { FotoCapaTab } from './FotoCapaTab'
import { FOTO_ERRO_MENSAGENS, ERRO_GENERICO_FOTO } from '@/lib/perfilErros'
import { ApiError } from '@/lib/api'

export interface FotoPerfilDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialTab?: 'upload' | 'capa'
}

function extrairErro(err: unknown): string {
  if (err instanceof ApiError && err.codigo && FOTO_ERRO_MENSAGENS[err.codigo]) {
    return FOTO_ERRO_MENSAGENS[err.codigo]
  }
  if (err instanceof Error && err.message) {
    return err.message
  }
  return ERRO_GENERICO_FOTO
}

export function FotoPerfilDialog({ open, onOpenChange, initialTab = 'upload' }: FotoPerfilDialogProps) {
  const [tabAtiva, setTabAtiva] = useState<'upload' | 'capa'>(initialTab)
  const [erroBackend, setErroBackend] = useState<string | null>(null)
  const { isSalvando, enviarFoto, usarCapa } = usePerfilStore()

  useEffect(() => {
    if (open) {
      setTabAtiva(initialTab)
      setErroBackend(null)
    }
  }, [open, initialTab])

  async function handleUpload(arquivo: File) {
    setErroBackend(null)
    try {
      await enviarFoto(arquivo)
      onOpenChange(false)
    } catch (err) {
      setErroBackend(extrairErro(err))
    }
  }

  async function handleCapa(jogoId: number) {
    setErroBackend(null)
    try {
      await usarCapa(jogoId)
      onOpenChange(false)
    } catch (err) {
      setErroBackend(extrairErro(err))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md border-[var(--border)] bg-[var(--bg-surface)] p-5 text-[var(--text-primary)] sm:p-6">
        <DialogHeader className="gap-1">
          <DialogTitle className="text-base font-bold text-[var(--text-primary)]">Foto de perfil</DialogTitle>
          <DialogDescription className="text-xs text-[var(--text-secondary)]">
            Escolha como sua imagem de perfil será exibida no Memory Card.
          </DialogDescription>
        </DialogHeader>

        <div role="tablist" aria-label="Opções de foto de perfil" className="flex border-b border-[var(--border)]">
          <button
            type="button"
            role="tab"
            id="tab-upload"
            aria-controls="tabpanel-upload"
            aria-selected={tabAtiva === 'upload'}
            onClick={() => { setTabAtiva('upload'); setErroBackend(null) }}
            className={`flex-1 min-h-11 border-b-2 py-2.5 text-xs font-semibold transition-colors ${
              tabAtiva === 'upload'
                ? 'border-[var(--accent)] text-[var(--text-primary)]'
                : 'border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)]'
            }`}
          >
            Enviar foto
          </button>
          <button
            type="button"
            role="tab"
            id="tab-capa"
            aria-controls="tabpanel-capa"
            aria-selected={tabAtiva === 'capa'}
            onClick={() => { setTabAtiva('capa'); setErroBackend(null) }}
            className={`flex-1 min-h-11 border-b-2 py-2.5 text-xs font-semibold transition-colors ${
              tabAtiva === 'capa'
                ? 'border-[var(--accent)] text-[var(--text-primary)]'
                : 'border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)]'
            }`}
          >
            Capa de um jogo
          </button>
        </div>

        <div className="pt-2">
          {tabAtiva === 'upload' ? (
            <div role="tabpanel" id="tabpanel-upload" aria-labelledby="tab-upload">
              <FotoUploadTab
                isLoading={isSalvando}
                erroBackend={erroBackend}
                onSubmit={handleUpload}
                onCancel={() => onOpenChange(false)}
              />
            </div>
          ) : (
            <div role="tabpanel" id="tabpanel-capa" aria-labelledby="tab-capa">
              <FotoCapaTab
                isLoading={isSalvando}
                erroBackend={erroBackend}
                onSubmit={handleCapa}
                onCancel={() => onOpenChange(false)}
              />
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
