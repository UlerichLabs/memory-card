import { useRef, type KeyboardEvent } from 'react'

interface TimeInputProps {
  horas: string
  minutos: string
  segundos: string
  setHoras: (value: string) => void
  setMinutos: (value: string) => void
  setSegundos: (value: string) => void
}

export function TimeInput({ horas, minutos, segundos, setHoras, setMinutos, setSegundos }: TimeInputProps) {
  const refs = [useRef<HTMLInputElement>(null), useRef<HTMLInputElement>(null), useRef<HTMLInputElement>(null)]
  const values = [horas, minutos, segundos]
  const setters = [setHoras, setMinutos, setSegundos]

  function move(index: number) {
    if (index < 2) refs[index + 1].current?.focus()
    else refs[index].current?.form?.querySelector<HTMLElement>('[aria-label="1"]')?.focus()
  }

  function handleKey(index: number, event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter') { event.preventDefault(); move(index) }
  }

  return (
    <div className="flex h-11 min-w-0 items-center justify-between rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2">
      {values.map((value, index) => (
        <div key={index} className="flex items-center gap-1">
          <input ref={refs[index]} value={value} onChange={(event) => { const next = event.target.value; if (!/^\d*$/.test(next)) return; if (index > 0 && Number(next) > 59) return; setters[index](next); if (index > 0 && next.length === 2) move(index) }} onKeyDown={(event) => handleKey(index, event)} placeholder={index === 0 ? '0' : '00'} inputMode="numeric" aria-label={['Horas', 'Minutos', 'Segundos'][index]} className="w-10 bg-transparent text-center text-sm text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]" />
          <span className="text-xs font-semibold text-[var(--text-muted)]">{['h', 'm', 's'][index]}</span>
        </div>
      ))}
    </div>
  )
}
