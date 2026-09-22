import { useEffect, useRef, useState } from 'react'
import { Gamepad2 } from 'lucide-react'
import { jogosService, type IGDBJogoSugestao } from '@/lib/services/jogosService'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export interface GameFormAutocompleteProps {
  nome: string
  onChangeNome: (nome: string) => void
  onSelectSugestao: (sugestao: IGDBJogoSugestao) => void
  igdbCapaUrl?: string
  error?: string
}

export function GameFormAutocomplete({
  nome, onChangeNome, onSelectSugestao, igdbCapaUrl, error,
}: GameFormAutocompleteProps) {
  const [sugestoes, setSugestoes] = useState<IGDBJogoSugestao[]>([])
  const [isOpen, setIsOpen] = useState(false)
  const [isSearching, setIsSearching] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const handleChange = (valor: string) => {
    onChangeNome(valor)
    if (valor.trim().length < 2) {
      setSugestoes([])
      setIsOpen(false)
    }
  }

  useEffect(() => {
    const termo = nome.trim()
    if (termo.length < 2) return

    const controller = new AbortController()
    const timer = setTimeout(async () => {
      setIsSearching(true)
      try {
        const res = await jogosService.buscarIGDB(termo, undefined, controller.signal)
        setSugestoes(res || [])
        setIsOpen((res || []).length > 0)
      } catch {
        setSugestoes([])
      } finally {
        setIsSearching(false)
      }
    }, 350)

    return () => {
      clearTimeout(timer)
      controller.abort()
    }
  }, [nome])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <div ref={containerRef} className="relative flex flex-col gap-1.5">
      <Label htmlFor="nome" className="text-secondary text-sm font-medium">
        Nome do jogo *
      </Label>
      <div className="flex gap-3 items-center">
        {igdbCapaUrl && (
          <img
            src={igdbCapaUrl}
            alt="Capa do jogo"
            className="w-10 h-14 object-cover rounded border border-border shrink-0"
          />
        )}
        <div className="relative flex-1">
          <Input
            id="nome"
            name="nome"
            value={nome}
            onChange={(e) => handleChange(e.target.value)}
            onFocus={() => {
              if (sugestoes.length > 0) setIsOpen(true)
            }}
            placeholder="Ex: Chrono Trigger"
            aria-invalid={!!error}
            className="bg-surface-alt border-border-subtle text-primary"
            autoComplete="off"
          />
          {isSearching && (
            <span className="absolute right-2.5 top-1/2 -translate-y-1/2 text-xs text-muted">
              Buscando...
            </span>
          )}
        </div>
      </div>
      {error && <span className="text-xs text-[#E05A4E]">{error}</span>}
      {isOpen && sugestoes.length > 0 && (
        <ul
          role="listbox"
          aria-label="Sugestões do IGDB"
          className="absolute top-full left-0 right-0 z-50 mt-1 max-h-60 overflow-y-auto rounded-lg border border-border bg-surface p-1 shadow-lg"
        >
          {sugestoes.map((sugestao) => {
            const capa = sugestao.cover?.url
              ? sugestao.cover.url.startsWith('//')
                ? `https:${sugestao.cover.url}`
                : sugestao.cover.url
              : null
            return (
              <li
                key={sugestao.id}
                role="option"
                aria-selected={false}
                onClick={() => {
                  onSelectSugestao(sugestao)
                  setIsOpen(false)
                }}
                className="flex items-center gap-3 rounded-md px-3 py-2 cursor-pointer text-sm text-primary hover:bg-surface-alt transition-colors"
              >
                {capa ? (
                  <img src={capa} alt="" className="w-7 h-10 object-cover rounded border border-border" />
                ) : (
                  <div className="w-7 h-10 rounded bg-surface-alt border border-border flex items-center justify-center text-muted">
                    <Gamepad2 className="w-4 h-4" />
                  </div>
                )}
                <div className="flex flex-col flex-1 min-w-0">
                  <span className="truncate font-medium">{sugestao.name}</span>
                  {sugestao.first_release_date && (
                    <span className="text-xs text-muted">
                      {new Date(sugestao.first_release_date * 1000).getFullYear()}
                    </span>
                  )}
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
