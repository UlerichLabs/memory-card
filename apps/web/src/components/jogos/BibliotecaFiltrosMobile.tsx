import { useState } from 'react'
import { SlidersHorizontal, X } from 'lucide-react'
import type { Dificuldade } from '@/types/jogos'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { BibliotecaSelect, type BibliotecaSelectOption } from './BibliotecaSelect'

const DIFICULDADES: Array<{ nivel: Dificuldade; label: string }> = [
  { nivel: 'C', label: 'Muito fácil' },
  { nivel: 'B', label: 'Fácil' },
  { nivel: 'A', label: 'Normal' },
  { nivel: 'AA', label: 'Difícil' },
  { nivel: 'AAA', label: 'Muito difícil' },
]
const NOTAS_OPCOES: BibliotecaSelectOption[] = Array.from({ length: 11 }, (_, i) => ({ value: String(i + 1), label: String(i + 1) }))

export interface BibliotecaFiltrosMobileProps {
  filtrosAtivosCount: number
  consoleVal: string; generoVal: string; tipoVal: string; anoVal: string
  notaMinVal: number; notaMaxVal: number; dificuldadeVal?: Dificuldade
  opcoesConsole: BibliotecaSelectOption[]; opcoesGenero: BibliotecaSelectOption[]
  opcoesTipo: BibliotecaSelectOption[]; opcoesAno: BibliotecaSelectOption[]
  onConsoleChange: (val: string) => void; onGeneroChange: (val: string) => void
  onTipoChange: (val: string) => void; onAnoChange: (val: string) => void
  onNotaMinChange: (val: number) => void; onNotaMaxChange: (val: number) => void
  onDificuldadeToggle: (dif: Dificuldade) => void; onLimparFiltros: () => void
}

export function BibliotecaFiltrosMobile(p: BibliotecaFiltrosMobileProps) {
  const [aberto, setAberto] = useState(false)
  const chips: Array<{ id: string; label: string; onRemover: () => void }> = []
  if (p.consoleVal) chips.push({ id: 'console', label: p.consoleVal, onRemover: () => p.onConsoleChange('') })
  if (p.generoVal) chips.push({ id: 'genero', label: p.generoVal, onRemover: () => p.onGeneroChange('') })
  if (p.tipoVal) chips.push({ id: 'tipo', label: p.tipoVal, onRemover: () => p.onTipoChange('') })
  if (p.anoVal) chips.push({ id: 'ano', label: p.anoVal, onRemover: () => p.onAnoChange('') })
  if (p.notaMinVal > 1 || p.notaMaxVal < 11) chips.push({ id: 'nota', label: `Nota: ${p.notaMinVal} a ${p.notaMaxVal}`, onRemover: () => { p.onNotaMinChange(1); p.onNotaMaxChange(11) } })
  if (p.dificuldadeVal) {
    const nomeDif = DIFICULDADES.find((d) => d.nivel === p.dificuldadeVal)?.label || p.dificuldadeVal
    chips.push({ id: 'dif', label: `Dificuldade: ${nomeDif}`, onRemover: () => p.onDificuldadeToggle(p.dificuldadeVal!) })
  }

  return (
    <div className="flex flex-col gap-2 md:hidden">
      <Sheet open={aberto} onOpenChange={setAberto}>
        <SheetTrigger
          type="button"
          aria-label="Abrir filtros"
          className="flex h-[44px] items-center justify-center gap-2 rounded-[10px] border border-[var(--biblioteca-input-border)] bg-[var(--biblioteca-input-bg)] px-3 text-[13px] font-medium text-[var(--biblioteca-input-text)] transition-colors hover:border-[var(--biblioteca-control-border-hover)]"
        >
          <SlidersHorizontal className="h-4 w-4 text-[var(--biblioteca-control-placeholder)]" aria-hidden="true" />
          <span>Filtros {p.filtrosAtivosCount > 0 ? `(${p.filtrosAtivosCount})` : ''}</span>
        </SheetTrigger>
        <SheetContent side="bottom" className="max-h-[85vh] overflow-y-auto rounded-t-[14px] border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-4">
          <SheetHeader className="pb-2">
            <SheetTitle className="text-sm font-bold text-[var(--biblioteca-text-primary)]">Filtros</SheetTitle>
          </SheetHeader>
          <div className="space-y-4 pt-2">
            <div className="space-y-1"><label className="text-xs text-[var(--biblioteca-text-muted)]">Plataforma</label><BibliotecaSelect value={p.consoleVal} onChange={p.onConsoleChange} options={p.opcoesConsole} ariaLabel="Filtrar por console" /></div>
            <div className="space-y-1"><label className="text-xs text-[var(--biblioteca-text-muted)]">Gênero</label><BibliotecaSelect value={p.generoVal} onChange={p.onGeneroChange} options={p.opcoesGenero} ariaLabel="Filtrar por gênero" /></div>
            <div className="space-y-1"><label className="text-xs text-[var(--biblioteca-text-muted)]">Tipo</label><BibliotecaSelect value={p.tipoVal} onChange={p.onTipoChange} options={p.opcoesTipo} ariaLabel="Filtrar por tipo" /></div>
            <div className="space-y-1"><label className="text-xs text-[var(--biblioteca-text-muted)]">Ano</label><BibliotecaSelect value={p.anoVal} onChange={p.onAnoChange} options={p.opcoesAno} ariaLabel="Filtrar por ano" /></div>
            <div className="space-y-1">
              <label className="text-xs text-[var(--biblioteca-text-muted)]">Faixa de nota</label>
              <div className="flex items-center gap-2">
                <BibliotecaSelect compact ativo={p.notaMinVal > 1} value={String(p.notaMinVal)} onChange={(v) => p.onNotaMinChange(Number(v))} options={NOTAS_OPCOES} ariaLabel="Nota mínima" />
                <span className="text-xs text-[var(--biblioteca-text-muted)]">a</span>
                <BibliotecaSelect compact ativo={p.notaMaxVal < 11} value={String(p.notaMaxVal)} onChange={(v) => p.onNotaMaxChange(Number(v))} options={NOTAS_OPCOES} ariaLabel="Nota máxima" />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs text-[var(--biblioteca-text-muted)]">Dificuldade</label>
              <div className="grid grid-cols-2 gap-1.5 sm:grid-cols-3">
                {DIFICULDADES.map(({ nivel, label }) => {
                  const ativo = p.dificuldadeVal === nivel
                  const code = nivel.toLowerCase()
                  const style = ativo
                    ? { backgroundColor: `var(--dif-${code}-fill)`, color: 'var(--nota-badge-text)', borderColor: `var(--dif-${code}-fill)` }
                    : { backgroundColor: 'var(--dif-btn-bg)', color: `var(--dif-${code}-text)`, borderColor: `var(--dif-${code}-border)` }
                  return (
                    <button
                      key={nivel}
                      type="button"
                      onClick={() => p.onDificuldadeToggle(nivel)}
                      aria-pressed={ativo}
                      style={style}
                      className="h-8 rounded-[6px] border text-xs font-semibold transition-colors"
                    >
                      {label}
                    </button>
                  )
                })}
              </div>
            </div>
            {p.filtrosAtivosCount > 0 && (
              <button
                type="button"
                onClick={() => { p.onLimparFiltros(); setAberto(false) }}
                className="w-full rounded-[8px] border border-[var(--biblioteca-btn-limpar-border)] bg-transparent py-2.5 text-xs font-semibold text-[var(--biblioteca-btn-limpar-text)] transition-colors hover:text-[var(--biblioteca-text-primary)]"
              >
                Limpar todos os filtros
              </button>
            )}
          </div>
        </SheetContent>
      </Sheet>
      {chips.length > 0 && (
        <div className="flex flex-wrap gap-1.5 pt-1">
          {chips.map((chip) => (
            <span key={chip.id} className="inline-flex items-center gap-1 rounded-[4px] border border-[var(--biblioteca-input-border)] bg-[var(--biblioteca-input-bg)] px-2 py-0.5 text-[11px] text-[var(--biblioteca-text-secondary)]">
              <span>{chip.label}</span>
              <button type="button" onClick={chip.onRemover} aria-label={`Remover filtro ${chip.label}`} className="text-[var(--biblioteca-control-placeholder)] hover:text-[var(--biblioteca-text-primary)]">
                <X className="h-3 w-3" aria-hidden="true" />
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
