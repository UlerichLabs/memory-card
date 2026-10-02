import { Gamepad2, Monitor, Smartphone } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

export interface ConsoleTema {
  familia: string
  cor: string
  corFundo: string
  corBorda: string
  Icone: LucideIcon
}

const temas: Record<string, { cor: string; Icone: LucideIcon }> = {
  Nintendo: { cor: '#E60012', Icone: Gamepad2 },
  PlayStation: { cor: '#0070D1', Icone: Gamepad2 },
  Xbox: { cor: '#107C10', Icone: Gamepad2 },
  PC: { cor: '#8FA8C8', Icone: Monitor },
  Sega: { cor: '#3B5BDB', Icone: Gamepad2 },
  Atari: { cor: '#F28C28', Icone: Gamepad2 },
  Arcade: { cor: '#D946EF', Icone: Gamepad2 },
  Mobile: { cor: '#14B8A6', Icone: Smartphone },
  Outro: { cor: 'var(--text-secondary)', Icone: Gamepad2 },
}

function normalizar(nome: string): string {
  return nome.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase()
}

function obterFamilia(nome: string): string {
  const familias: Array<[string, string[]]> = [
    ['Nintendo', ['nintendo', 'switch', 'wii', 'gamecube', 'game boy', 'gba', 'ds', '3ds', 'snes', 'super nintendo', 'nes', 'n64']],
    ['PlayStation', ['playstation', 'ps1', 'ps2', 'ps3', 'ps4', 'ps5', 'psp', 'vita']],
    ['Xbox', ['xbox', 'xbox 360', 'one', 'series']],
    ['PC', ['pc', 'windows', 'steam', 'mac', 'linux']],
    ['Sega', ['sega', 'mega drive', 'genesis', 'master system', 'saturn', 'dreamcast', 'game gear']],
    ['Atari', ['atari']],
    ['Arcade', ['arcade', 'fliperama']],
    ['Mobile', ['mobile', 'android', 'ios']],
  ]
  return familias.find(([, termos]) => termos.some((termo) => nome.includes(termo)))?.[0] ?? 'Outro'
}

export function getConsoleTema(nome: string): ConsoleTema {
  const familia = obterFamilia(normalizar(nome))
  const tema = temas[familia]
  return {
    familia,
    cor: tema.cor,
    corFundo: `color-mix(in srgb, ${tema.cor} 12%, transparent)`,
    corBorda: `color-mix(in srgb, ${tema.cor} 35%, transparent)`,
    Icone: tema.Icone,
  }
}
