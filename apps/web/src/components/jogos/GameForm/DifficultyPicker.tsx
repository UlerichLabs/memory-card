import type { Dificuldade } from '@/lib/services/jogosService'

interface DifficultyPickerProps {
  value?: Dificuldade
  onChange: (value: Dificuldade) => void
  error?: string
}

const options: Array<{ value: Dificuldade; label: string }> = [
  { value: 'C', label: 'Muito fácil' }, { value: 'B', label: 'Fácil' }, { value: 'A', label: 'Normal' }, { value: 'AA', label: 'Difícil' }, { value: 'AAA', label: 'Muito difícil' },
]

const colors: Record<Dificuldade, string> = { C: 'difficulty-c', B: 'difficulty-b', A: 'difficulty-a', AA: 'difficulty-aa', AAA: 'difficulty-aaa' }

export function DifficultyPicker({ value, onChange, error }: DifficultyPickerProps) {
  return (
    <div className="flex flex-col gap-2"><span className="text-sm text-[var(--text-secondary)]">Dificuldade *</span><div role="group" aria-label="Dificuldade" className="flex flex-wrap gap-1.5">{options.map((item) => <button key={item.value} type="button" aria-label={item.label} aria-pressed={value === item.value} onClick={() => onChange(item.value)} className={`rounded-md border px-2.5 py-2 text-xs font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${value === item.value ? `bg-[var(--${colors[item.value]})] text-[var(--bg-primary)] ${item.value === 'AAA' ? 'highlight-pulse' : ''}` : `border-[var(--${colors[item.value]})] text-[var(--${colors[item.value]})]`}`}>{item.label}</button>)}</div>{error && <span className="text-xs text-[var(--danger)]">{error}</span>}</div>
  )
}
