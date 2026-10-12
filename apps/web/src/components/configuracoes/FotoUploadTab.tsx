import { useEffect, useRef, useState, type DragEvent } from 'react'
import { Upload, AlertCircle } from 'lucide-react'
import { FOTO_ERRO_MENSAGENS } from '@/lib/perfilErros'

const TIPOS_VALIDOS = ['image/jpeg', 'image/png', 'image/webp']
const TAMANHO_MAXIMO = 2 * 1024 * 1024

interface FotoUploadTabProps {
  isLoading: boolean
  erroBackend: string | null
  onSubmit: (arquivo: File) => Promise<void>
  onCancel: () => void
}

export function FotoUploadTab({ isLoading, erroBackend, onSubmit, onCancel }: FotoUploadTabProps) {
  const [arquivo, setArquivo] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [erroLocal, setErroLocal] = useState<string | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!arquivo) return
    const url = URL.createObjectURL(arquivo)
    setPreviewUrl(url)
    return () => {
      URL.revokeObjectURL(url)
      setPreviewUrl(null)
    }
  }, [arquivo])

  function validarEAtribuir(file: File) {
    setErroLocal(null)
    if (!TIPOS_VALIDOS.includes(file.type)) {
      setErroLocal(FOTO_ERRO_MENSAGENS['perfil.tipo_arquivo_invalido'])
      setArquivo(null)
      setPreviewUrl(null)
      return
    }
    if (file.size > TAMANHO_MAXIMO) {
      setErroLocal(FOTO_ERRO_MENSAGENS['perfil.arquivo_muito_grande'])
      setArquivo(null)
      setPreviewUrl(null)
      return
    }
    setArquivo(file)
  }


  function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (file) validarEAtribuir(file)
    event.target.value = ''
  }

  function handleDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault()
    setIsDragging(false)
    const file = event.dataTransfer.files?.[0]
    if (file) validarEAtribuir(file)
  }

  const erroExibicao = erroLocal || erroBackend

  return (
    <div className="space-y-4">
      <div
        onDragOver={(e) => { e.preventDefault(); setIsDragging(true) }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
        className={`flex flex-col items-center justify-center rounded-[10px] border-2 border-dashed p-6 text-center cursor-pointer transition-colors ${
          isDragging ? 'border-[var(--accent)] bg-[var(--bg-surface-alt)]' : 'border-[var(--border-subtle)] hover:border-[var(--accent)] bg-[var(--bg-surface-alt)]/50'
        }`}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); inputRef.current?.click() } }}
        aria-label="Selecionar imagem de perfil"
      >
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="hidden"
          onChange={handleFileChange}
          aria-hidden="true"
        />

        {previewUrl ? (
          <div className="flex flex-col items-center gap-3">
            <img
              src={previewUrl}
              alt="Prévia da foto de perfil"
              className="h-24 w-24 rounded-full object-cover border-2 border-[var(--border)] shadow-md"
            />
            <p className="text-xs text-[var(--text-secondary)]">Clique ou arraste para trocar a imagem</p>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-2">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[var(--bg-surface-alt)] text-[var(--accent)] border border-[var(--border-subtle)]">
              <Upload className="h-5 w-5" aria-hidden="true" />
            </div>
            <p className="text-sm font-medium text-[var(--text-primary)]">Arraste uma foto ou clique para escolher</p>
            <p className="text-xs text-[var(--text-muted)]">JPG, PNG ou WebP até 2 MB</p>
          </div>
        )}
      </div>

      {erroExibicao && (
        <div role="alert" className="flex items-center gap-2 rounded-[8px] border border-[var(--danger)]/30 bg-[var(--danger)]/10 px-3 py-2 text-xs text-[var(--danger)]">
          <AlertCircle className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>{erroExibicao}</span>
        </div>
      )}

      <div className="flex items-center justify-end gap-2 pt-2">
        <button
          type="button"
          onClick={onCancel}
          disabled={isLoading}
          className="min-h-11 rounded-[7px] border border-[var(--border-subtle)] px-4 py-2 text-xs font-semibold text-[var(--text-secondary)] transition-colors hover:text-[var(--text-primary)] disabled:opacity-50"
        >
          Cancelar
        </button>
        <button
          type="button"
          onClick={() => { if (arquivo) void onSubmit(arquivo) }}
          disabled={!arquivo || isLoading || Boolean(erroLocal)}
          className="min-h-11 rounded-[7px] bg-[var(--accent)] px-5 py-2 text-xs font-bold text-[#0E0F12] transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isLoading ? 'Salvando…' : 'Usar como foto'}
        </button>
      </div>
    </div>
  )
}
