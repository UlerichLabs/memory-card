import { useState, useRef, useEffect } from 'react'
import { AJUDA_REGRA } from './listas.constants'
import type { RegraTipo, IGDBFranquiaSugestao } from '@/types/listas'

export interface NovaListaRegraSectionProps {
  regraTipo: RegraTipo; setRegraTipo: (r: RegraTipo) => void
  regraValor: string; setRegraValor: (v: string) => void; setIgdbId: (id: number | null) => void
  meta: string; setMeta: (m: string) => void; errors: Record<string, string>
  franquias: IGDBFranquiaSugestao[]; setFranquias: (f: IGDBFranquiaSugestao[]) => void
  isSearchingFranquias: boolean; plataformas: string[]; generos: string[]
  buscarFranquias: (termo: string) => void; disabled?: boolean
}

const REGRAS: Array<{ tipo: RegraTipo; titulo: string; desc: string }> = [
  { tipo: 'franquia', titulo: 'Franquia', desc: 'Busca a franquia no IGDB' },
  { tipo: 'plataforma', titulo: 'Plataforma', desc: 'Conta zeramentos na plataforma' },
  { tipo: 'genero', titulo: 'Gênero', desc: 'Conta zeramentos do gênero' },
  { tipo: 'manual', titulo: 'Eu escolho os jogos', desc: 'Monte a lista na mão' },
]

export function NovaListaRegraSection(props: NovaListaRegraSectionProps) {
  const {
    regraTipo, setRegraTipo, regraValor, setRegraValor, setIgdbId,
    meta, setMeta, errors, franquias, setFranquias, isSearchingFranquias,
    plataformas, generos, buscarFranquias, disabled = false,
  } = props
  const [showDropdown, setShowDropdown] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setShowDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleOutside)
    return () => document.removeEventListener('mousedown', handleOutside)
  }, [])

  return (
    <div className="flex flex-col gap-3.5 rounded-xl border border-[var(--lista-desafio-box-border)] bg-[var(--lista-desafio-box-bg)] p-4">
      <div className="flex flex-col gap-2">
        <span className="text-[13px] font-medium text-[var(--lista-text-light)]">Quais jogos contam?</span>
        <div role="radiogroup" aria-label="Regra" className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          {REGRAS.map((r) => {
            const ativa = regraTipo === r.tipo
            return (
              <button
                key={r.tipo}
                type="button"
                role="radio"
                aria-checked={ativa}
                disabled={disabled}
                onClick={() => setRegraTipo(r.tipo)}
                className={`flex flex-col gap-0.5 rounded-[10px] p-[10px_12px] text-left transition-colors ${
                  ativa
                    ? 'border border-[var(--lista-radio-regra-active-border)] bg-[var(--lista-radio-regra-active-bg)]'
                    : 'border border-[var(--lista-radio-border)] bg-[var(--lista-radio-bg)] hover:border-[var(--lista-text-dim)]'
                } ${disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}`}
              >
                <span className="text-[13px] font-semibold text-[var(--text-primary)]">{r.titulo}</span>
                <span className="text-[11px] text-[var(--lista-text-muted)]">{r.desc}</span>
              </button>
            )
          })}
        </div>
      </div>

      <div className={`grid gap-3 ${regraTipo === 'manual' ? 'grid-cols-1' : 'grid-cols-1 sm:grid-cols-[1fr_150px]'}`}>
        {regraTipo === 'franquia' && (
          <div ref={dropdownRef} className="relative flex flex-col gap-1.5">
            <label htmlFor="regra-franquia" className="text-[13px] font-medium text-[var(--lista-text-light)]">Franquia *</label>
            <input
              id="regra-franquia"
              disabled={disabled}
              value={regraValor}
              onChange={(e) => { buscarFranquias(e.target.value); setShowDropdown(true) }}
              onFocus={() => { if (franquias.length) setShowDropdown(true) }}
              placeholder="Ex: The Legend of Zelda, Final Fantasy"
              autoComplete="off"
              className="h-11 rounded-[10px] border border-[var(--lista-chip-border)] bg-[var(--lista-chip-bg)] px-3 text-[14px] text-[var(--text-primary)] focus:outline-none"
            />
            {isSearchingFranquias && <span className="absolute right-3 top-9 text-xs text-[var(--lista-text-muted)]">Buscando...</span>}
            {showDropdown && franquias.length > 0 && (
              <ul role="listbox" className="custom-scrollbar absolute left-0 right-0 top-full z-50 mt-1 max-h-48 overflow-y-auto rounded-lg border border-[var(--border)] bg-[var(--bg-surface)] p-1 shadow-xl">
                {franquias.map((f) => (
                  <li
                    key={f.id}
                    role="option"
                    aria-selected={false}
                    onClick={() => { setRegraValor(f.name); setIgdbId(f.id); setShowDropdown(false); setFranquias([]) }}
                    className="cursor-pointer rounded px-3 py-2 text-sm text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]"
                  >
                    {f.name}
                  </li>
                ))}
              </ul>
            )}
            {errors.regraValor && <span className="text-xs text-[var(--danger)]">{errors.regraValor}</span>}
          </div>
        )}

        {(regraTipo === 'plataforma' || regraTipo === 'genero') && (
          <div className="flex flex-col gap-1.5">
            <label htmlFor="regra-valor" className="text-[13px] font-medium text-[var(--lista-text-light)]">
              {regraTipo === 'plataforma' ? 'Plataforma *' : 'Gênero *'}
            </label>
            <input
              id="regra-valor"
              list={regraTipo === 'plataforma' ? 'plataformas-sugestoes' : 'generos-sugestoes'}
              disabled={disabled}
              value={regraValor}
              onChange={(e) => setRegraValor(e.target.value)}
              placeholder={regraTipo === 'plataforma' ? 'Ex: SNES, PlayStation 2' : 'Ex: RPG, Platform'}
              className="h-11 rounded-[10px] border border-[var(--lista-chip-border)] bg-[var(--lista-chip-bg)] px-3 text-[14px] text-[var(--text-primary)] focus:outline-none"
            />
            <datalist id={regraTipo === 'plataforma' ? 'plataformas-sugestoes' : 'generos-sugestoes'}>
              {(regraTipo === 'plataforma' ? plataformas : generos).map((opt) => (
                <option key={opt} value={opt} />
              ))}
            </datalist>
            {errors.regraValor && <span className="text-xs text-[var(--danger)]">{errors.regraValor}</span>}
          </div>
        )}

        <div className="flex flex-col gap-1.5">
          <label htmlFor="desafio-meta" className="text-[13px] font-medium text-[var(--lista-text-light)]">
            Meta {(regraTipo === 'plataforma' || regraTipo === 'genero') && '*'}
          </label>
          <input
            id="desafio-meta"
            type="number"
            min={1}
            max={10000}
            inputMode="numeric"
            value={meta}
            onChange={(e) => setMeta(e.target.value)}
            placeholder={regraTipo === 'franquia' || regraTipo === 'manual' ? 'Todos' : 'Ex: 10'}
            className="h-11 rounded-[10px] border border-[var(--lista-chip-border)] bg-[var(--lista-chip-bg)] px-3 text-[14px] text-[var(--text-primary)] tabular-nums focus:outline-none"
          />
          {errors.meta && <span className="text-xs text-[var(--danger)]">{errors.meta}</span>}
        </div>
      </div>

      <p className="text-[12px] text-[var(--lista-text-muted)]">{AJUDA_REGRA[regraTipo]}</p>
    </div>
  )
}
