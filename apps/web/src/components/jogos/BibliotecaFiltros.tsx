import { useEffect, useState } from 'react'
import { Search, X } from 'lucide-react'
import type { Dificuldade } from '@/types/jogos'
import { BibliotecaSelect, type BibliotecaSelectOption } from './BibliotecaSelect'
import { BibliotecaFiltrosMobile } from './BibliotecaFiltrosMobile'

const DIFICULDADES: Dificuldade[] = ['C', 'B', 'A', 'AA', 'AAA']
const DIFICULDADE_VAR: Record<Dificuldade, string> = {
  C: 'var(--difficulty-c)', B: 'var(--difficulty-b)', A: 'var(--difficulty-a)', AA: 'var(--difficulty-aa)', AAA: 'var(--difficulty-aaa)',
}
const NOTAS_OPCOES: BibliotecaSelectOption[] = Array.from({ length: 11 }, (_, i) => ({ value: String(i + 1), label: String(i + 1) }))

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

  const filtrosAtivos = (props.consoleVal ? 1 : 0) + (props.generoVal ? 1 : 0) + (props.tipoVal ? 1 : 0) +
    (props.anoVal ? 1 : 0) + (props.notaMinVal > 1 || props.notaMaxVal < 11 ? 1 : 0) + (props.dificuldadeVal ? 1 : 0)

  return (
    <div className="rounded-[10px] border border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-[14px_16px]">
      <div className="grid grid-cols-1 gap-[10px] md:grid-cols-[1fr_auto_auto_auto_auto]">
        <div className="relative flex items-center">
          <Search className="absolute left-3 h-3.5 w-3.5 text-[var(--biblioteca-control-placeholder)]" aria-hidden="true" />
          <input
            type="search"
            value={buscaLocal}
            onChange={(e) => setBuscaLocal(e.target.value)}
            placeholder="Buscar por nome…"
            aria-label="Filtrar jogos na biblioteca"
            className="h-9 w-full rounded-[6px] border border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] pl-8 pr-8 text-[13px] text-[var(--biblioteca-text-primary)] placeholder:text-[var(--biblioteca-control-placeholder)] transition-colors hover:border-[var(--biblioteca-control-border-hover)] focus:border-[var(--biblioteca-control-border-focus)] focus:outline-none"
          />
          {buscaLocal && (
            <button
              type="button"
              onClick={() => { setBuscaLocal(''); props.onBuscaChange('') }}
              aria-label="Limpar busca"
              className="absolute right-2.5 text-[var(--biblioteca-control-placeholder)] hover:text-[var(--biblioteca-text-primary)]"
            >
              <X className="h-3.5 w-3.5" aria-hidden="true" />
            </button>
          )}
        </div>

        <BibliotecaFiltrosMobile
          filtrosAtivosCount={filtrosAtivos}
          consoleVal={props.consoleVal}
          generoVal={props.generoVal}
          tipoVal={props.tipoVal}
          anoVal={props.anoVal}
          notaMinVal={props.notaMinVal}
          notaMaxVal={props.notaMaxVal}
          dificuldadeVal={props.dificuldadeVal}
          opcoesConsole={props.opcoesConsole}
          opcoesGenero={props.opcoesGenero}
          opcoesTipo={props.opcoesTipo}
          opcoesAno={props.opcoesAno}
          onConsoleChange={props.onConsoleChange}
          onGeneroChange={props.onGeneroChange}
          onTipoChange={props.onTipoChange}
          onAnoChange={props.onAnoChange}
          onNotaMinChange={(min) => props.onNotaChange(min, Math.max(min, props.notaMaxVal))}
          onNotaMaxChange={(max) => props.onNotaChange(Math.min(max, props.notaMinVal), max)}
          onDificuldadeToggle={props.onDificuldadeToggle}
          onLimparFiltros={props.onLimparFiltros}
        />

        <div className="hidden w-44 md:block"><BibliotecaSelect value={props.consoleVal} onChange={props.onConsoleChange} options={props.opcoesConsole} ariaLabel="Filtrar por console" /></div>
        <div className="hidden w-44 md:block"><BibliotecaSelect value={props.generoVal} onChange={props.onGeneroChange} options={props.opcoesGenero} ariaLabel="Filtrar por gênero" /></div>
        <div className="hidden w-36 md:block"><BibliotecaSelect value={props.tipoVal} onChange={props.onTipoChange} options={props.opcoesTipo} ariaLabel="Filtrar por tipo" /></div>
        <div className="hidden w-28 md:block"><BibliotecaSelect value={props.anoVal} onChange={props.onAnoChange} options={props.opcoesAno} ariaLabel="Filtrar por ano" /></div>
      </div>

      <div className="hidden items-center justify-between border-t border-[var(--biblioteca-divider)] pt-[10px] mt-[10px] md:flex">
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-2">
            <span className="text-xs text-[var(--biblioteca-text-muted)]">Nota:</span>
            <div className="w-[60px]"><BibliotecaSelect compact value={String(props.notaMinVal)} onChange={(v) => { const n = Number(v); props.onNotaChange(n, Math.max(n, props.notaMaxVal)) }} options={NOTAS_OPCOES} ariaLabel="Nota mínima" /></div>
            <span className="text-xs text-[var(--biblioteca-text-muted)]">a</span>
            <div className="w-[60px]"><BibliotecaSelect compact value={String(props.notaMaxVal)} onChange={(v) => { const n = Number(v); props.onNotaChange(Math.min(n, props.notaMinVal), n) }} options={NOTAS_OPCOES} ariaLabel="Nota máxima" /></div>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-xs text-[var(--biblioteca-text-muted)]">Dificuldade:</span>
            <div className="flex gap-1">
              {DIFICULDADES.map((dif) => {
                const ativo = props.dificuldadeVal === dif
                const cor = DIFICULDADE_VAR[dif]
                return (
                  <button
                    key={dif}
                    type="button"
                    onClick={() => props.onDificuldadeToggle(dif)}
                    aria-pressed={ativo}
                    style={ativo ? { borderColor: cor, color: cor, backgroundColor: 'var(--biblioteca-control-active-bg)' } : undefined}
                    className={`h-6 min-w-[28px] rounded-[4px] border px-1.5 text-[11px] font-bold transition-colors ${
                      ativo ? '' : 'border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] text-[var(--biblioteca-text-muted)] hover:border-[var(--biblioteca-control-border-hover)]'
                    }`}
                  >
                    {dif}
                  </button>
                )
              })}
            </div>
          </div>
        </div>

        {filtrosAtivos > 0 && (
          <div className="flex items-center gap-2.5">
            <span className="text-[11px] text-[var(--biblioteca-text-muted)]">
              {filtrosAtivos} {filtrosAtivos === 1 ? 'filtro ativo' : 'filtros ativos'}
            </span>
            <button
              type="button"
              onClick={props.onLimparFiltros}
              className="text-xs text-[var(--biblioteca-text-muted)] underline decoration-dotted transition-colors hover:text-[var(--biblioteca-text-primary)]"
            >
              Limpar filtros
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
