import { useState } from 'react'
import { CalendarIcon } from 'lucide-react'
import { ptBR } from 'date-fns/locale'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'

interface DateInputProps {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
  error?: string
}

function isoToDisplay(value: string) {
  if (!value) return ''
  const [year, month, day] = value.split('-')
  return year && month && day ? `${day}/${month}/${year}` : value
}

function displayToIso(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 8)
  if (digits.length < 6) return ''
  const day = digits.slice(0, 2)
  const month = digits.slice(2, 4)
  let year = digits.slice(4)
  if (year.length === 2) year = Number(year) > new Date().getFullYear() % 100 ? `19${year}` : `20${year}`
  if (year.length !== 4) return ''
  return `${year}-${month}-${day}`
}

function maskDate(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 8)
  return [digits.slice(0, 2), digits.slice(2, 4), digits.slice(4)].filter(Boolean).join('/')
}

export function DateInput({ id, label, value, onChange, error }: DateInputProps) {
  const [open, setOpen] = useState(false)
  const [display, setDisplay] = useState(isoToDisplay(value))
  const selected = value ? new Date(`${value}T00:00:00`) : undefined

  function commit(next: string) {
    const masked = maskDate(next)
    setDisplay(masked)
    const iso = displayToIso(masked)
    if (iso) onChange(iso)
    else if (!masked) onChange('')
  }

  function handleBlur() {
    const iso = displayToIso(display)
    if (iso) {
      onChange(iso)
      setDisplay(isoToDisplay(iso))
    }
  }

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm text-[var(--text-secondary)]">{label}</label>
      <div className={`flex h-9 items-center rounded-lg border bg-[var(--bg-surface-alt)] px-2 ${error ? 'border-[var(--danger)]' : 'border-[var(--border-subtle)]'}`}>
        <input id={id} value={display} onChange={(event) => commit(event.target.value)} onBlur={handleBlur} onKeyDown={(event) => { if (event.key === 'Enter') handleBlur() }} placeholder="dd/mm/aaaa" inputMode="numeric" aria-invalid={!!error} className="min-w-0 flex-1 bg-transparent text-sm text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)]" />
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger type="button" aria-label="Abrir calendário" className="rounded p-1 text-[var(--text-secondary)] focus-visible:ring-2 focus-visible:ring-[var(--accent)]"><CalendarIcon className="size-4" /></PopoverTrigger>
          <PopoverContent className="z-50 w-auto rounded-xl border border-[var(--border)] bg-[var(--bg-surface)] p-3 text-[var(--text-primary)]" align="end">
            <Calendar mode="single" selected={selected} onSelect={(date) => { if (date) { const iso = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`; onChange(iso); setDisplay(isoToDisplay(iso)); setOpen(false) } }} locale={ptBR} />
          </PopoverContent>
        </Popover>
      </div>
      {error && <span className="text-xs text-[var(--danger)]">{error}</span>}
    </div>
  )
}
