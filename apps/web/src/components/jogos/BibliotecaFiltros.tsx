import { useEffect, useState } from 'react'
import { Search, X } from 'lucide-react'
import type { Dificuldade } from '@/types/jogos'
import { BibliotecaSelect, type BibliotecaSelectOption } from './BibliotecaSelect'
import { BibliotecaFiltrosMobile } from './BibliotecaFiltrosMobile'

const DIFICULDADES: Array<{ nivel: Dificuldade; label: string }> = [
  { nivel: 'C', label: 'Muito fácil' }, { nivel: 'B', label: 'Fácil' }, { nivel: 'A', label: 'Normal' },
  { nivel: 'AA', label: 'Difícil' }, { nivel: 'AAA', label: 'Muito difícil' },
]

const NOTAS_OPCOES: BibliotecaSelectOption[] = Array.from({ length: 11 }, (_, i) => ({
  value: String(i + 1), label: String(i + 1),
}))

export interface BibliotecaFiltrosProps {
  busca: string; consoleVal: string; generoVal: string; tipoVal: string; anoVal: string
  notaMinVal: number; notaMaxVal: number; dificuldadeVal?: Dificuldade
  opcoesConsole: BibliotecaSelectOption[]; opcoesGenero: BibliotecaSelectOption[]
  opcoesTipo: BibliotecaSelectOption[]; opcoesAno: BibliotecaSelectOption[]
  onBuscaChange: (termo: string) => void; onConsoleChange: (val: string) => void
  onGeneroChange: (val: string) => void; onTipoChange: (val: string) => void
  onAnoChange: (val: string) => void; onNotaChange: (min: number, max: number) => void
  onDificuldadeToggle: (dif: Dificuldade) => void; onLimparFiltros: () => void
}

export function BibliotecaFiltros(props: BibliotecaFiltrosProps) {
  const [buscaLocal, setBuscaLocal] = useState(props.busca)

  useEffect(() => { setBuscaLocal(props.busca) }, [props.busca])

  useEffect(() => {
    const timer = setTimeout(() => {
      if (buscaLocal.trim() !== props.busca.trim()) props.onBuscaChange(buscaLocal.trim())
    }, 300)
    return () => clearTimeout(timer)
  }, [buscaLocal, props.busca])

  const filtrosAtivos = (props.busca.trim() ? 1 : 0) + (props.consoleVal ? 1 : 0) + (props.generoVal ? 1 : 0) +
    (props.tipoVal ? 1 : 0) + (props.anoVal ? 1 : 0) + (props.notaMinVal > 1 || props.notaMaxVal < 11 ? 1 : 0) + (props.dificuldadeVal ? 1 : 0)

  return (
    <div className="rounded-[12px] border border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-4 space-y-3">
      <div className="flex flex-col gap-2.5 xl:grid xl:grid-cols-[1fr_210px_180px_150px_140px]">
        <div className="relative flex items-center w-full">
          <Search className="absolute left-3.5 h-4 w-4 text-[var(--biblioteca-control-placeholder)]" aria-hidden="true" />
          <input
            type="search"
            value={buscaLocal}
            onChange={(e) => setBuscaLocal(e.target.value)}
            placeholder="Buscar pelo nome do jogo…"
            aria-label="Filtrar jogos na biblioteca"
            className="h-[44px] w-full rounded-[10px] border border-[var(--biblioteca-input-border)] bg-[var(--biblioteca-input-bg)] pl-10 pr-9 text-[13px] text-[var(--biblioteca-input-text)] placeholder:text-[var(--biblioteca-control-placeholder)] transition-colors hover:border-[var(--biblioteca-control-border-hover)] focus:border-[var(--biblioteca-input-active-border)] focus:outline-none"
          />
          {buscaLocal && (
            <button
              type="button"
              onClick={() => { setBuscaLocal(''); props.onBuscaChange('') }}
              aria-label="Limpar busca"
              className="absolute right-3 text-[var(--biblioteca-control-placeholder)] hover:text-[var(--biblioteca-text-primary)]"
            >
              <X className="h-4 w-4" aria-hidden="true" />
            </button>
          )}
        </div>

        <BibliotecaFiltrosMobile
          filtrosAtivosCount={filtrosAtivos}
          consoleVal={props.consoleVal} generoVal={props.generoVal} tipoVal={props.tipoVal} anoVal={props.anoVal}
          notaMinVal={props.notaMinVal} notaMaxVal={props.notaMaxVal} dificuldadeVal={props.dificuldadeVal}
          opcoesConsole={props.opcoesConsole} opcoesGenero={props.opcoesGenero}
          opcoesTipo={props.opcoesTipo} opcoesAno={props.opcoesAno}
          onConsoleChange={props.onConsoleChange} onGeneroChange={props.onGeneroChange}
          onTipoChange={props.onTipoChange} onAnoChange={props.onAnoChange}
          onNotaMinChange={(min) => props.onNotaChange(min, Math.max(min, props.notaMaxVal))}
          onNotaMaxChange={(max) => props.onNotaChange(Math.min(max, props.notaMinVal), max)}
          onDificuldadeToggle={props.onDificuldadeToggle} onLimparFiltros={props.onLimparFiltros}
        />

        <div className="hidden md:flex md:flex-wrap xl:contents gap-2.5">
          <div className="flex-1 min-w-[180px] xl:w-auto xl:flex-none"><BibliotecaSelect value={props.consoleVal} onChange={props.onConsoleChange} options={props.opcoesConsole} ariaLabel="Filtrar por console" /></div>
          <div className="flex-1 min-w-[160px] xl:w-auto xl:flex-none"><BibliotecaSelect value={props.generoVal} onChange={props.onGeneroChange} options={props.opcoesGenero} ariaLabel="Filtrar por gênero" /></div>
          <div className="flex-1 min-w-[140px] xl:w-auto xl:flex-none"><BibliotecaSelect value={props.tipoVal} onChange={props.onTipoChange} options={props.opcoesTipo} ariaLabel="Filtrar por tipo" /></div>
          <div className="flex-1 min-w-[130px] xl:w-auto xl:flex-none"><BibliotecaSelect value={props.anoVal} onChange={props.onAnoChange} options={props.opcoesAno} ariaLabel="Filtrar por ano" /></div>
        </div>
      </div>

      <div className="hidden items-center justify-between border-t border-[var(--biblioteca-divider)] pt-3 md:flex">
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-2">
            <span className="text-[13px] text-[var(--biblioteca-text-muted)]">Nota:</span>
            <BibliotecaSelect compact ativo={props.notaMinVal > 1} value={String(props.notaMinVal)} onChange={(v) => { const n = Number(v); props.onNotaChange(n, Math.max(n, props.notaMaxVal)) }} options={NOTAS_OPCOES} ariaLabel="Nota mínima" />
            <span className="text-[13px] text-[var(--biblioteca-text-muted)]">a</span>
            <BibliotecaSelect compact ativo={props.notaMaxVal < 11} value={String(props.notaMaxVal)} onChange={(v) => { const n = Number(v); props.onNotaChange(Math.min(n, props.notaMinVal), n) }} options={NOTAS_OPCOES} ariaLabel="Nota máxima" />
          </div>

          <div className="flex items-center gap-2">
            <span className="text-[13px] text-[var(--biblioteca-text-muted)]">Dificuldade:</span>
            <div className="flex gap-1.5">
              {DIFICULDADES.map(({ nivel, label }) => {
                const ativo = props.dificuldadeVal === nivel
                const code = nivel.toLowerCase()
                const style = ativo
                  ? { backgroundColor: `var(--dif-${code}-fill)`, color: 'var(--nota-badge-text)', borderColor: `var(--dif-${code}-fill)` }
                  : { backgroundColor: 'var(--dif-btn-bg)', color: `var(--dif-${code}-text)`, borderColor: `var(--dif-${code}-border)` }
                return (
                  <button
                    key={nivel}
                    type="button"
                    onClick={() => props.onDificuldadeToggle(nivel)}
                    aria-pressed={ativo}
                    style={style}
                    className="h-[36px] rounded-[8px] border px-3 text-[13px] font-semibold transition-colors"
                  >
                    {label}
                  </button>
                )
              })}
            </div>
          </div>
        </div>

        {filtrosAtivos > 0 && (
          <div className="flex items-center gap-2.5">
            <span className="flex h-[22px] w-[22px] items-center justify-center rounded-full bg-[var(--biblioteca-badge-filtro)] text-[12px] font-bold text-white">
              {filtrosAtivos}
            </span>
            <span className="text-[13px] text-[var(--biblioteca-text-filtros-ativos)]">
              {filtrosAtivos === 1 ? 'filtro ativo' : 'filtros ativos'}
            </span>
            <button
              type="button"
              onClick={props.onLimparFiltros}
              className="h-[36px] rounded-[8px] border border-[var(--biblioteca-btn-limpar-border)] px-3 text-[13px] font-medium text-[var(--biblioteca-btn-limpar-text)] transition-colors hover:text-[var(--biblioteca-text-primary)]"
            >
              Limpar filtros
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
