import { Gamepad2, Joystick, Monitor, Smartphone, CircleX } from 'lucide-react'
import type { LucideProps } from 'lucide-react'
import { createElement, type ComponentType } from 'react'

export interface ConsoleTema {
  familia: string
  cor: string
  corFundo: string
  corBorda: string
  Icone: ComponentType<LucideProps>
}

function PlayStationIcon(props: LucideProps) {
  const svgProps = { ...props, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', strokeWidth: 1.8, strokeLinecap: 'round', strokeLinejoin: 'round' }
  return createElement('svg', svgProps,
    createElement('path', { d: 'm7 4 2 4-2 3-2-3 2-4Z' }),
    createElement('circle', { cx: '17', cy: '7', r: '2' }),
    createElement('path', { d: 'm6 17 3-3 3 3-3 3-3-3Z' }),
    createElement('path', { d: 'm16 14 3 3m0-3-3 3' }),
  )
}

const temas: Record<string, { cor: string; Icone: ComponentType<LucideProps> }> = {
  Nintendo: { cor: '#E60012', Icone: Gamepad2 },
  PlayStation: { cor: '#0070D1', Icone: PlayStationIcon },
  Xbox: { cor: '#107C10', Icone: CircleX },
  PC: { cor: '#8FA8C8', Icone: Monitor },
  Sega: { cor: '#3B5BDB', Icone: Joystick },
  Atari: { cor: '#F28C28', Icone: Joystick },
  Arcade: { cor: '#D946EF', Icone: Joystick },
  Mobile: { cor: '#14B8A6', Icone: Smartphone },
  Outro: { cor: 'var(--text-secondary)', Icone: Gamepad2 },
}

function normalizar(nome: string): string {
  return nome.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase()
}

function obterTokens(nome: string): string[] {
  return nome.split(/[^a-z0-9]+/).filter(Boolean)
}

function contemTermo(tokens: string[], termo: string): boolean {
  const partes = termo.split(' ')
  return tokens.some((_, indice) => partes.every((parte, deslocamento) => tokens[indice + deslocamento] === parte))
}

function obterFamilia(nome: string): string {
  const tokens = obterTokens(nome)
  if (contemTermo(tokens, 'pc') && contemTermo(tokens, 'engine')) return 'Outro'
  const familias: Array<[string, string[]]> = [
    ['Sega', ['sega', 'mega drive', 'genesis', 'master system', 'saturn', 'dreamcast', 'game gear']],
    ['Nintendo', ['nintendo', 'switch', 'wii', 'gamecube', 'game boy', 'gba', 'ds', '3ds', 'snes', 'super nintendo', 'nes', 'n64']],
    ['PlayStation', ['playstation', 'ps1', 'ps2', 'ps3', 'ps4', 'ps5', 'psp', 'vita']],
    ['Xbox', ['xbox', 'xbox 360', 'one', 'series']],
    ['PC', ['pc', 'windows', 'steam', 'mac', 'linux']],
    ['Atari', ['atari']],
    ['Arcade', ['arcade', 'fliperama']],
    ['Mobile', ['mobile', 'android', 'ios']],
  ]
  return familias.find(([, termos]) => termos.some((termo) => contemTermo(tokens, termo)))?.[0] ?? 'Outro'
}

export function getConsoleTema(nome: string): ConsoleTema {
  const familia = obterFamilia(normalizar(nome))
  const tema = temas[familia]
  return {
    familia,
    cor: tema.cor,
    corFundo: `color-mix(in srgb, ${tema.cor} 14%, transparent)`,
    corBorda: `color-mix(in srgb, ${tema.cor} 40%, transparent)`,
    Icone: tema.Icone,
  }
}
