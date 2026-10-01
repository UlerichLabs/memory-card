import * as React from 'react'
import { CalendarIcon } from 'lucide-react'
import { ptBR } from 'date-fns/locale'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'
import { cn } from 'cn'

export interface DatePickerProps {
  id?: string
  name?: string
  value?: string
  onChange: (value: string) => void
  placeholder?: string
  error?: boolean
  disabled?: boolean
  ariaLabel?: string
  ariaInvalid?: boolean
}

function parseISODate(dateStr?: string): Date | undefined {
  if (!dateStr) return undefined
  const cleaned = dateStr.slice(0, 10)
  const parts = cleaned.split('-')
  if (parts.length !== 3) return undefined
  const [year, month, day] = parts.map(Number)
  if (!year || !month || !day) return undefined
  return new Date(year, month - 1, day)
}

function formatToISO(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function formatDisplayDate(dateStr?: string): string {
  if (!dateStr) return ''
  const cleaned = dateStr.slice(0, 10)
  const parts = cleaned.split('-')
  if (parts.length !== 3) return dateStr
  const [year, month, day] = parts
  return `${day.padStart(2, '0')}/${month.padStart(2, '0')}/${year}`
}

export function DatePicker({
  id,
  name,
  value,
  onChange,
  placeholder = 'dd/mm/aaaa',
  error,
  disabled,
  ariaLabel,
  ariaInvalid,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false)
  const selectedDate = parseISODate(value)

  const handleSelect = (date: Date | undefined) => {
    if (date) {
      onChange(formatToISO(date))
      setOpen(false)
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        id={id}
        name={name}
        type="button"
        disabled={disabled}
        aria-label={ariaLabel || placeholder}
        aria-invalid={ariaInvalid || !!error}
        className={cn(
          'flex h-8 w-full items-center justify-between rounded-lg border bg-[var(--bg-surface-alt)] px-2.5 py-1 text-sm text-left transition-colors outline-none cursor-pointer',
          'focus-visible:ring-2 focus-visible:ring-[var(--accent)] hover:border-[var(--border-focus)]',
          error ? 'border-[var(--danger)]' : 'border-[var(--border-subtle)]',
          disabled && 'cursor-not-allowed opacity-50'
        )}
      >
        <span className={value ? 'text-[var(--text-primary)]' : 'text-[var(--text-muted)]'}>
          {value ? formatDisplayDate(value) : placeholder}
        </span>
        <CalendarIcon className="size-4 text-[var(--text-secondary)] shrink-0" />
      </PopoverTrigger>
      <PopoverContent
        align="start"
        className="w-auto p-3 border border-[var(--border)] bg-[var(--bg-surface)] text-[var(--text-primary)] shadow-xl rounded-xl z-50"
      >
        <Calendar
          mode="single"
          selected={selectedDate}
          onSelect={handleSelect}
          locale={ptBR}
        />
      </PopoverContent>
    </Popover>
  )
}
