import { useState } from 'react'
import { Search } from 'lucide-react'
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

const REGRAS: Array<{ tipo: RegraTipo; titulo: string }> = [
  { tipo: 'franquia', titulo: 'Franquia' },
  { tipo: 'plataforma', titulo: 'Plataforma' },
  { tipo: 'genero', titulo: 'Gênero' },
  { tipo: 'manual', titulo: 'Eu escolho' },
]

export function NovaListaRegraSection(props: NovaListaRegraSectionProps) {
  const {
    regraTipo, setRegraTipo, regraValor, setRegraValor, setIgdbId,
    meta, setMeta, errors, franquias, setFranquias, isSearchingFranquias,
    plataformas, generos, buscarFranquias, disabled = false,
  } = props
  const [showDropdown, setShowDropdown] = useState(false)

  return (
    <div className="flex min-w-0 flex-col gap-[14px] rounded-xl border border-[var(--lista-desafio-box-border)] bg-[var(--lista-desafio-box-bg)] p-4">
      <div className="flex min-w-0 flex-col gap-2">
        <span className="text-[13px] font-medium text-[var(--lista-text-light)]">Quais jogos contam?</span>
        <div role="radiogroup" aria-label="Regra" className="grid grid-cols-2 gap-[6px] rounded-[10px] border border-[var(--lista-desafio-box-border)] bg-[var(--modal-bg)] p-1 min-[420px]:grid-cols-4">
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
                className={`min-w-0 truncate rounded-[7px] border px-2 text-[13px] font-semibold transition-colors ${
                  ativa
                    ? 'h-9 border-[var(--lista-pill-desafio-border)] bg-[var(--lista-pill-desafio-bg)] text-[var(--hall-ouro)]'
                    : 'h-9 border-transparent text-[var(--lista-text-secondary)] hover:bg-[var(--bg-surface-alt)]'
                } ${disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}`}
              >
                {r.titulo}
              </button>
            )
          })}
        </div>
      </div>

      <div className="text-[12px] text-[var(--lista-text-muted)]">{AJUDA_REGRA[regraTipo]}</div>

      <div className={`grid min-w-0 gap-3 ${regraTipo === 'manual' ? 'grid-cols-1' : 'grid-cols-1 min-[420px]:grid-cols-[minmax(0,1fr)_112px]'}`}>
        {regraTipo === 'franquia' && (
          <div className="flex min-w-0 flex-col gap-1.5">
            <label htmlFor="regra-franquia" className="text-[13px] font-medium text-[var(--lista-text-light)]">Franquia *</label>
            <div className="relative min-w-0">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--lista-text-muted)]" />
              <input
                id="regra-franquia"
                disabled={disabled}
                value={regraValor}
                onChange={(e) => { buscarFranquias(e.target.value); setShowDropdown(true) }}
                onFocus={() => { if (franquias.length) setShowDropdown(true) }}
                placeholder="Ex: The Legend of Zelda, Final Fantasy"
                autoComplete="off"
                className="h-11 w-full min-w-0 rounded-[10px] border border-[var(--lista-chip-border)] bg-[var(--lista-chip-bg)] pl-9 pr-3 text-[14px] text-[var(--text-primary)] focus:outline-none"
              />
              {isSearchingFranquias && <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-[var(--lista-text-muted)]">Buscando...</span>}
            </div>
            {showDropdown && franquias.length > 0 && (
              <ul role="listbox" className="custom-scrollbar max-h-48 overflow-y-auto rounded-lg border border-[var(--border)] bg-[var(--bg-surface)] p-1 shadow-xl">
                {franquias.map((f) => (
                  <li
                    key={f.id}
                    role="option"
                    aria-selected={false}
                  >
                    <button type="button" onClick={() => { setRegraValor(f.name); setIgdbId(f.id); setShowDropdown(false); setFranquias([]) }} className="w-full rounded px-3 py-2 text-left text-sm text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]">
                      {f.name}
                    </button>
                  </li>
                ))}
              </ul>
            )}
            {errors.regraValor && <span className="text-xs text-[var(--danger)]">{errors.regraValor}</span>}
          </div>
        )}

        {(regraTipo === 'plataforma' || regraTipo === 'genero') && (
          <div className="flex min-w-0 flex-col gap-1.5">
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
              className="h-11 w-full min-w-0 rounded-[10px] border border-[var(--lista-chip-border)] bg-[var(--lista-chip-bg)] px-3 text-[14px] text-[var(--text-primary)] focus:outline-none"
            />
            <datalist id={regraTipo === 'plataforma' ? 'plataformas-sugestoes' : 'generos-sugestoes'}>
              {(regraTipo === 'plataforma' ? plataformas : generos).map((opt) => (
                <option key={opt} value={opt} />
              ))}
            </datalist>
            {errors.regraValor && <span className="text-xs text-[var(--danger)]">{errors.regraValor}</span>}
          </div>
        )}

        <div className="flex min-w-0 flex-col gap-1.5">
          <label htmlFor="desafio-meta" className="text-[13px] font-medium text-[var(--lista-text-light)]">
            Meta {(regraTipo === 'plataforma' || regraTipo === 'genero') && '*'}
          </label>
          <input
            id="desafio-meta"
            type="text"
            inputMode="numeric"
            value={meta}
            onChange={(e) => setMeta(e.target.value)}
            placeholder={regraTipo === 'franquia' || regraTipo === 'manual' ? 'Todos' : 'Ex: 10'}
            className="no-spinner h-11 w-full min-w-0 rounded-[10px] border border-[var(--lista-chip-border)] bg-[var(--lista-chip-bg)] px-3 text-[14px] text-[var(--text-primary)] tabular-nums focus:outline-none"
          />
          {errors.meta && <span className="text-xs text-[var(--danger)]">{errors.meta}</span>}
        </div>
      </div>

    </div>
  )
}
