import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Check } from 'lucide-react'

export interface CustomSelectOption {
  value: string
  label: string
}

export interface CustomSelectProps {
  id?: string
  name?: string
  value: string
  onChange: (value: string) => void
  options: CustomSelectOption[]
  placeholder?: string
  disabled?: boolean
  error?: boolean
  ariaLabel?: string
}

export function CustomSelect({
  id,
  name,
  value,
  onChange,
  options,
  placeholder = 'Selecione uma opção',
  disabled = false,
  error = false,
  ariaLabel,
}: CustomSelectProps) {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const selectedOption = options.find((opt) => opt.value === value)

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <div ref={containerRef} className="relative w-full">
      <button
        id={id}
        type="button"
        value={value}
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        aria-label={ariaLabel}
        onClick={() => !disabled && setIsOpen((prev) => !prev)}
        className={`flex h-8 w-full items-center justify-between rounded-lg border bg-[var(--bg-surface-alt)] px-2.5 py-1 text-sm outline-none transition-colors focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${
          error ? 'border-[var(--danger)]' : 'border-[var(--border-subtle)]'
        } ${selectedOption ? 'text-[var(--text-primary)]' : 'text-[var(--text-muted)]'} ${
          disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'
        }`}
      >
        <span className="truncate">{selectedOption ? selectedOption.label : placeholder}</span>
        <ChevronDown
          className={`h-4 w-4 shrink-0 text-[var(--text-faint)] transition-transform duration-200 ${
            isOpen ? 'rotate-180' : ''
          }`}
          aria-hidden="true"
        />
      </button>

      {name && <input type="hidden" name={name} value={value} />}

      {isOpen && (
        <ul
          role="listbox"
          aria-label={ariaLabel || placeholder}
          className="custom-scrollbar absolute top-full left-0 right-0 z-50 mt-1 max-h-52 overflow-y-auto rounded-lg border border-[var(--border)] bg-[var(--bg-surface)] p-1 shadow-2xl"
        >
          {options.map((opt) => {
            const isSelected = opt.value === value
            return (
              <li
                key={opt.value}
                role="option"
                aria-selected={isSelected}
                onClick={() => {
                  onChange(opt.value)
                  setIsOpen(false)
                }}
                className={`flex w-full items-center justify-between rounded-md px-2.5 py-1.5 text-xs cursor-pointer transition-colors ${
                  isSelected
                    ? 'bg-[var(--bg-surface-alt)] font-semibold text-[var(--accent)]'
                    : 'text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]'
                }`}
              >
                <span className="truncate">{opt.label}</span>
                {isSelected && <Check className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
