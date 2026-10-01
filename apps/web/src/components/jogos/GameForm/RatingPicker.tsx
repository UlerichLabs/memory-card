interface RatingPickerProps {
  value?: number
  onChange: (value: number) => void
  error?: string
}

export function RatingPicker({ value, onChange, error }: RatingPickerProps) {
  const label = value === 11 ? '11 · Jogo da Vida' : value ? `${value} de 11` : ''
  const tokens = (item: number) => item === 11
    ? { fill: 'var(--nota-11-fill)', selectedBorder: 'var(--nota-11-fill-border)', text: 'var(--nota-11-text)', border: 'var(--nota-11-border)' }
    : { fill: `var(--nota-${item}-fill)`, selectedBorder: `var(--nota-${item}-fill-border)`, text: `var(--nota-${item}-text)`, border: `var(--nota-${item}-border)` }
  return <div className="flex flex-col gap-2">
    <div className="flex items-center justify-between"><span className="text-sm text-[var(--text-secondary)]">Nota *</span>{label && <span className="text-xs text-[var(--text-secondary)]">{label}</span>}</div>
    <div role="group" aria-label="Nota" className="flex flex-wrap gap-2 md:flex-nowrap">
      {Array.from({ length: 11 }, (_, index) => index + 1).map((item) => { const color = tokens(item); const selected = value === item; return <button key={item} type="button" aria-label={String(item)} aria-pressed={selected} onClick={() => onChange(item)} style={{ backgroundColor: selected ? color.fill : 'var(--nota-button-bg)', color: selected ? 'var(--modal-bg)' : color.text, borderColor: selected ? color.selectedBorder : color.border }} className={`h-11 w-11 shrink-0 rounded-[10px] border text-sm font-bold transition-[background,color] duration-150 focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${selected && item === 11 ? 'animate-pulso-ouro animate-brilho-ouro' : ''}`}>{item}</button> })}
    </div>
    {error && <span className="text-xs text-[var(--danger)]">{error}</span>}
  </div>
}
