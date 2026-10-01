import type { Dificuldade } from '@/lib/services/jogosService'
import type { CustomSelectOption } from '@/components/ui/CustomSelect'

export const DIFICULDADE_OPCOES: Array<{ value: Dificuldade; label: string }> = [
  { value: 'C', label: 'C — Muito fácil' },
  { value: 'B', label: 'B — Fácil' },
  { value: 'A', label: 'A — Normal' },
  { value: 'AA', label: 'AA — Difícil' },
  { value: 'AAA', label: 'AAA — Muito difícil' },
]

export const CONSOLES_PADRAO: CustomSelectOption[] = [
  'PC', 'PlayStation 5', 'PlayStation 4', 'PlayStation 3', 'PlayStation 2', 'PlayStation',
  'Xbox Series X/S', 'Xbox One', 'Xbox 360', 'Nintendo Switch', 'Nintendo Wii U',
  'Nintendo Wii', 'Nintendo GameCube', 'Nintendo 64', 'Super Nintendo', 'NES',
  'Nintendo DS', 'Nintendo 3DS', 'Game Boy Advance', 'PSP', 'PS Vita', 'Outro',
].map((c) => ({ value: c, label: c }))
