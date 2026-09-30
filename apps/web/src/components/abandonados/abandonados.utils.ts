import type { ListarAbandonadosParams } from '@/types/abandonados'

export function montarQueryStringAbandonados(params?: ListarAbandonadosParams): string {
  if (!params) return ''
  const sp = new URLSearchParams()
  if (params.pagina && params.pagina > 1) sp.set('pagina', String(params.pagina))
  if (params.por_pagina && params.por_pagina !== 12) sp.set('por_pagina', String(params.por_pagina))
  if (params.busca?.trim()) sp.set('busca', params.busca.trim())
  if (params.console?.trim()) sp.set('console', params.console.trim())
  if (params.ordenar && params.ordenar !== 'recentes') sp.set('ordenar', params.ordenar)
  return sp.toString()
}

export function formatarTempoAbandonado(s: number): string {
  if (s <= 0) return '0h'
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  return m > 0 ? `${m}m` : `${s}s`
}

export function segundosParaHms(s: number): { horas: string; minutos: string; segundos: string } {
  if (!s || s <= 0) return { horas: '', minutos: '', segundos: '' }
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  return {
    horas: h > 0 ? String(h) : '',
    minutos: m > 0 ? String(m) : '',
    segundos: sec > 0 ? String(sec) : '',
  }
}

export function ehCancelado(err: unknown, sig?: AbortSignal): boolean {
  const isStatus499 = err && typeof err === 'object' && 'status' in err && (err as { status: number }).status === 499
  return Boolean(sig?.aborted || (err instanceof DOMException && err.name === 'AbortError') || isStatus499)
}

export function hojeIso(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function formatarDataBrasileira(iso: string): string {
  if (!iso) return ''
  const [ano, mes, dia] = iso.slice(0, 10).split('-')
  if (!ano || !mes || !dia) return iso
  return `${dia}/${mes}/${ano}`
}
