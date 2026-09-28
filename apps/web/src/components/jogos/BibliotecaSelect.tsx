import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Check } from 'lucide-react'

export interface BibliotecaSelectOption {
  value: string
  label: string
}

export interface BibliotecaSelectProps {
  value: string
  onChange: (value: string) => void
  options: BibliotecaSelectOption[]
  placeholder?: string
  ariaLabel?: string
  compact?: boolean
}

export function BibliotecaSelect({
  value,
  onChange,
  options,
  placeholder = 'Selecione',
  ariaLabel,
  compact = false,
}: BibliotecaSelectProps) {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const selectedOption = options.find((opt) => opt.value === value)
  const isAtivo = Boolean(value && value !== '')

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
        type="button"
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        aria-label={ariaLabel}
        onClick={() => setIsOpen((prev) => !prev)}
        className={`flex w-full items-center justify-between rounded-[6px] border px-2.5 transition-colors ${
          compact ? 'h-7 text-xs' : 'h-9 text-[13px]'
        } ${
          isAtivo
            ? 'border-[var(--biblioteca-control-active-border)] bg-[var(--biblioteca-control-active-bg)] text-[var(--biblioteca-control-active-text)]'
            : 'border-[var(--biblioteca-control-border)] bg-[var(--biblioteca-control-bg)] text-[var(--biblioteca-control-text)] hover:border-[var(--biblioteca-control-border-hover)]'
        }`}
      >
        <span className="truncate">{selectedOption ? selectedOption.label : placeholder}</span>
        <ChevronDown
          className={`ml-1.5 h-3.5 w-3.5 shrink-0 text-[var(--biblioteca-control-placeholder)] transition-transform duration-150 ${
            isOpen ? 'rotate-180' : ''
          }`}
          aria-hidden="true"
        />
      </button>

      {isOpen && (
        <ul
          role="listbox"
          aria-label={ariaLabel || placeholder}
          className="absolute left-0 right-0 top-full z-50 mt-1 max-h-56 overflow-y-auto rounded-[8px] border border-[var(--biblioteca-panel-border)] bg-[var(--biblioteca-panel-bg)] p-1 shadow-xl"
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
                className={`flex w-full cursor-pointer items-center justify-between rounded-[4px] px-2.5 py-1.5 text-xs transition-colors ${
                  isSelected
                    ? 'bg-[var(--biblioteca-control-active-bg)] font-bold text-[var(--accent)]'
                    : 'text-[var(--biblioteca-control-text)] hover:bg-[var(--biblioteca-control-bg)]'
                }`}
              >
                <span className="truncate">{opt.label}</span>
                {isSelected && <Check className="h-3.5 w-3.5 shrink-0 text-[var(--accent)]" aria-hidden="true" />}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
