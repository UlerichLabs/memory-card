import { Gamepad2, Joystick, Monitor, Smartphone, CircleX } from 'lucide-react'
import { siAndroid, siApple, siAtari, siPlaystation, siSega, siSteam } from 'simple-icons'
import { createElement, type ComponentType, type CSSProperties } from 'react'

interface ConsoleIconProps {
  className?: string
  style?: CSSProperties
  'aria-hidden'?: boolean | 'true' | 'false'
}

export interface ConsoleTema {
  familia: string
  cor: string
  corFundo: string
  corBorda: string
  Icone: ComponentType<ConsoleIconProps>
  IconeSuave: ComponentType<ConsoleIconProps>
  corSolida: string
  corTextoSolida: string
}

function SimpleIcon({ icon, props }: { icon: { path: string }; props: ConsoleIconProps }) {
  return createElement('svg', { ...props, viewBox: '0 0 24 24', fill: 'currentColor' }, createElement('path', { d: icon.path }))
}

function NintendoIcon(props: ConsoleIconProps) {
  return createElement('svg', { ...props, viewBox: '0 0 24 24', fill: 'currentColor' },
    createElement('rect', { x: '4', y: '7', width: '16', height: '10', rx: '4' }),
    createElement('path', { d: 'M8 10v4m-2-2h4m6-1h.01M18 13h.01' }),
  )
}

function NintendoSwitchIcon(props: ConsoleIconProps) {
  return createElement('svg', { ...props, viewBox: '0 0 24 24', fill: 'currentColor' },
    createElement('rect', { x: '4', y: '3', width: '6', height: '18', rx: '3' }),
    createElement('rect', { x: '14', y: '3', width: '6', height: '18', rx: '3' }),
    createElement('circle', { cx: '7', cy: '8', r: '1' }),
    createElement('circle', { cx: '17', cy: '16', r: '1' }),
  )
}

function PlayStationOutlineIcon(props: ConsoleIconProps) {
  const svgProps = { ...props, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', strokeWidth: 1.8, strokeLinecap: 'round', strokeLinejoin: 'round' }
  return createElement('svg', svgProps,
    createElement('path', { d: 'm7 4 2 4-2 3-2-3 2-4Z' }),
    createElement('circle', { cx: '17', cy: '7', r: '2' }),
    createElement('path', { d: 'm6 17 3-3 3 3-3 3-3-3Z' }),
    createElement('path', { d: 'm16 14 3 3m0-3-3 3' }),
  )
}

const temas: Record<string, { cor: string; corSolida: string; corTextoSolida: string; Icone: ComponentType<ConsoleIconProps>; IconeSuave: ComponentType<ConsoleIconProps> }> = {
  Nintendo: { cor: '#E60012', corSolida: '#E60012', corTextoSolida: '#fff', Icone: NintendoIcon, IconeSuave: Gamepad2 },
  PlayStation: { cor: '#0070D1', corSolida: '#0070D1', corTextoSolida: '#fff', Icone: (props) => SimpleIcon({ icon: siPlaystation, props }), IconeSuave: PlayStationOutlineIcon },
  Xbox: { cor: '#107C10', corSolida: '#107C10', corTextoSolida: '#fff', Icone: CircleX, IconeSuave: CircleX },
  PC: { cor: '#8FA8C8', corSolida: '#4B5563', corTextoSolida: '#fff', Icone: (props) => SimpleIcon({ icon: siSteam, props }), IconeSuave: Monitor },
  Sega: { cor: '#3B5BDB', corSolida: '#1D4ED8', corTextoSolida: '#fff', Icone: (props) => SimpleIcon({ icon: siSega, props }), IconeSuave: Joystick },
  Atari: { cor: '#F28C28', corSolida: '#F28C28', corTextoSolida: '#1A1B20', Icone: (props) => SimpleIcon({ icon: siAtari, props }), IconeSuave: Joystick },
  Arcade: { cor: '#D946EF', corSolida: '#A21CAF', corTextoSolida: '#fff', Icone: Joystick, IconeSuave: Joystick },
  Mobile: { cor: '#14B8A6', corSolida: '#0F9D8A', corTextoSolida: '#fff', Icone: Smartphone, IconeSuave: Smartphone },
  Outro: { cor: 'var(--text-secondary)', corSolida: '#4B5563', corTextoSolida: '#fff', Icone: Gamepad2, IconeSuave: Gamepad2 },
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
  const nomeNormalizado = normalizar(nome)
  const familia = obterFamilia(nomeNormalizado)
  const tema = temas[familia]
  const tokens = obterTokens(nomeNormalizado)
  const icone = familia === 'Nintendo' && contemTermo(tokens, 'switch')
    ? NintendoSwitchIcon
    : familia === 'Mobile' && contemTermo(tokens, 'ios')
      ? (props: ConsoleIconProps) => SimpleIcon({ icon: siApple, props })
      : familia === 'Mobile' && contemTermo(tokens, 'android')
        ? (props: ConsoleIconProps) => SimpleIcon({ icon: siAndroid, props })
        : familia === 'PC' && contemTermo(tokens, 'mac')
          ? (props: ConsoleIconProps) => SimpleIcon({ icon: siApple, props })
          : tema.Icone
  return {
    familia,
    cor: tema.cor,
    corFundo: `color-mix(in srgb, ${tema.cor} 14%, transparent)`,
    corBorda: `color-mix(in srgb, ${tema.cor} 40%, transparent)`,
    Icone: icone,
    IconeSuave: tema.IconeSuave,
    corSolida: tema.corSolida,
    corTextoSolida: tema.corTextoSolida,
  }
}
