interface RatingPickerProps {
  value?: number
  onChange: (value: number) => void
  error?: string
}

export function RatingPicker({ value, onChange, error }: RatingPickerProps) {
  const label = value === 11 ? '11 · Jogo da Vida' : value ? `${value} de 11` : 'Selecione'
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between"><span className="text-sm text-[var(--text-secondary)]">Nota *</span><span className="text-xs text-[var(--highlight-gold)]">{label}</span></div>
      <div role="group" aria-label="Nota" className="flex flex-wrap gap-1.5">
        {Array.from({ length: 11 }, (_, index) => index + 1).map((item) => { const color = item === 11 ? 'highlight-gold' : item <= 3 ? 'danger' : item <= 7 ? 'highlight-gold' : 'success'; return <button key={item} type="button" aria-label={String(item)} aria-pressed={value === item} onClick={() => onChange(item)} className={`h-11 w-11 rounded-md border text-sm font-bold transition-colors focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${value === item ? `bg-[var(--${color})] text-[var(--bg-primary)] ${item === 11 ? 'highlight-pulse' : ''}` : `border-[var(--${color})] text-[var(--${color})]`}`}>{item}</button> })}
      </div>
      {error && <span className="text-xs text-[var(--danger)]">{error}</span>}
    </div>
  )
}
