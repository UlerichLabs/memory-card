import { useContext, useEffect, useRef, useState } from 'react'
import { Gamepad2 } from 'lucide-react'
import { jogosService, type IGDBJogoSugestao } from '@/lib/services/jogosService'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AuthContext } from '@/store/authStore'
import { formatarCapaIGDB } from '@/lib/utils'

export interface GameFormAutocompleteProps {
  nome: string
  onChangeNome: (nome: string) => void
  onSelectSugestao: (sugestao: IGDBJogoSugestao) => void
  igdbCapaUrl?: string
  igdbDescricao?: string
  error?: string
}

export function GameFormAutocomplete({
  nome, onChangeNome, onSelectSugestao, igdbCapaUrl, igdbDescricao, error,
}: GameFormAutocompleteProps) {
  const auth = useContext(AuthContext)
  const token = auth?.sessao?.access_token
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
        const res = await jogosService.buscarIGDB(termo, token, controller.signal)
        setSugestoes(res || [])
        setIsOpen(true)
      } catch {
        setSugestoes([])
      } finally {
        setIsSearching(false)
      }
    }, 350)
    return () => { clearTimeout(timer); controller.abort() }
  }, [nome, token])

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
    <div ref={containerRef} className="relative flex flex-col gap-2">
      <Label htmlFor="nome" className="text-[var(--text-secondary)] text-sm font-medium">
        Nome do jogo *
      </Label>
      <div className="flex gap-3 items-center">
        {igdbCapaUrl ? (
          <img src={igdbCapaUrl} alt="Capa do jogo" className="h-[60px] w-[44px] shrink-0 rounded-[6px] border border-[var(--border)] object-cover shadow-sm" />
        ) : (
          <div className="flex h-[60px] w-[44px] shrink-0 items-center justify-center rounded-[6px] border-2 border-dashed border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] text-[var(--text-faint)]" aria-label="Placeholder de capa">
            <Gamepad2 className="h-5 w-5 opacity-40" aria-hidden="true" />
          </div>
        )}
        <div className="relative flex-1">
          <Input
            id="nome"
            name="nome"
            value={nome}
            onChange={(e) => handleChange(e.target.value)}
            onFocus={() => { if (sugestoes.length > 0) setIsOpen(true) }}
            placeholder="Ex: God of War, Chrono Trigger"
            aria-invalid={!!error}
            autoComplete="off"
            className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)]"
          />
          {isSearching && (
            <span className="absolute right-2.5 top-1/2 -translate-y-1/2 text-xs text-[var(--text-muted)]">Buscando...</span>
          )}
          {isOpen && (
            <ul role="listbox" aria-label="Sugestões do IGDB" className="custom-scrollbar absolute top-full left-0 right-0 z-50 mt-1 max-h-60 overflow-y-auto rounded-lg border border-[var(--border)] bg-[var(--bg-surface)] p-1 shadow-2xl">
              {isSearching ? (
                <li className="py-3 text-center text-xs text-[var(--text-muted)]">Buscando no IGDB...</li>
              ) : sugestoes.length === 0 ? (
                <li className="py-3 text-center text-xs text-[var(--text-muted)]">Nenhum jogo encontrado no IGDB</li>
              ) : (
                sugestoes.slice(0, 5).map((sugestao) => {
                  const capa = formatarCapaIGDB(sugestao.cover?.url)
                  const ano = sugestao.first_release_date ? new Date(sugestao.first_release_date * 1000).getFullYear() : null
                  const plats = sugestao.platforms?.map((p) => p.name).join(', ')
                  const subtitulo = [ano, plats].filter(Boolean).join(' · ')
                  return (
                    <li
                      key={sugestao.id}
                      role="option"
                      aria-selected={false}
                      onClick={() => { onSelectSugestao(sugestao); setIsOpen(false) }}
                      className="flex items-center gap-3 rounded-md px-3 py-2 cursor-pointer text-sm text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)] transition-colors"
                    >
                      {capa ? (
                        <img src={capa} alt="" className="h-[44px] w-[32px] shrink-0 rounded object-cover border border-[var(--border)]" />
                      ) : (
                        <div className="flex h-[44px] w-[32px] shrink-0 items-center justify-center rounded border border-[var(--border)] bg-[var(--bg-surface-alt)] text-[var(--text-muted)]">
                          <Gamepad2 className="h-4 w-4" />
                        </div>
                      )}
                      <div className="flex flex-col flex-1 min-w-0">
                        <span className="truncate font-semibold text-sm">{sugestao.name}</span>
                        {subtitulo && <span className="truncate text-xs text-[var(--text-muted)]">{subtitulo}</span>}
                      </div>
                    </li>
                  )
                })
              )}
            </ul>
          )}
        </div>
      </div>
      {error && <span className="text-xs text-[var(--danger)]">{error}</span>}
      {igdbCapaUrl && (
        <div className="flex items-center gap-3 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] p-2.5">
          <img src={igdbCapaUrl} alt="Capa selecionada" className="h-[48px] w-[36px] shrink-0 rounded object-cover border border-[var(--border)] shadow-sm" />
          <div className="min-w-0 flex-1 space-y-0.5">
            <p className="truncate text-xs font-bold text-[var(--text-primary)]">{nome}</p>
            {igdbDescricao && <p className="line-clamp-2 text-[11px] text-[var(--text-muted)]">{igdbDescricao}</p>}
          </div>
        </div>
      )}
    </div>
  )
}
