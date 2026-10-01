import type { Dificuldade } from '@/lib/services/jogosService'

interface DifficultyPickerProps {
  value?: Dificuldade
  onChange: (value: Dificuldade) => void
  error?: string
}

const options: Array<{ value: Dificuldade; label: string }> = [
  { value: 'C', label: 'Muito fácil' }, { value: 'B', label: 'Fácil' }, { value: 'A', label: 'Normal' }, { value: 'AA', label: 'Difícil' }, { value: 'AAA', label: 'Muito difícil' },
]

const tokens: Record<Dificuldade, { fill: string; selectedBorder: string; text: string; border: string }> = {
  C: { fill: 'var(--dif-c-fill)', selectedBorder: 'var(--dif-c-fill-border)', text: 'var(--dif-c-text)', border: 'var(--dif-c-border)' },
  B: { fill: 'var(--dif-b-fill)', selectedBorder: 'var(--dif-b-fill-border)', text: 'var(--dif-b-text)', border: 'var(--dif-b-border)' },
  A: { fill: 'var(--dif-a-fill)', selectedBorder: 'var(--dif-a-fill-border)', text: 'var(--dif-a-text)', border: 'var(--dif-a-border)' },
  AA: { fill: 'var(--dif-aa-fill)', selectedBorder: 'var(--dif-aa-fill-border)', text: 'var(--dif-aa-text)', border: 'var(--dif-aa-border)' },
  AAA: { fill: 'var(--dif-aaa-fill)', selectedBorder: 'var(--dif-aaa-fill-border)', text: 'var(--dif-aaa-text)', border: 'var(--dif-aaa-border)' },
}

export function DifficultyPicker({ value, onChange, error }: DifficultyPickerProps) {
  return <div className="flex flex-col gap-2"><span className="text-sm text-[var(--text-secondary)]">Dificuldade *</span><div role="group" aria-label="Dificuldade" className="grid grid-cols-1 gap-2 sm:grid-cols-5">{options.map((item) => { const color = tokens[item.value]; const selected = value === item.value; return <button key={item.value} type="button" aria-label={item.label} aria-pressed={selected} onClick={() => onChange(item.value)} style={{ backgroundColor: selected ? color.fill : 'var(--nota-button-bg)', color: selected ? 'var(--modal-bg)' : color.text, borderColor: selected ? color.selectedBorder : color.border }} className={`h-11 w-full rounded-[10px] border px-2 text-sm font-semibold transition-[background,color] duration-150 focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${selected && item.value === 'AAA' ? 'animate-pulso-vermelho' : ''}`}>{item.label}</button> })}</div>{error && <span className="text-xs text-[var(--danger)]">{error}</span>}</div>
}
