export function formatarHoras(segundos: number): string {
  const horas = Math.round(segundos / 3600)
  if (horas >= 1) return `${horas.toLocaleString('pt-BR')}h`
  return `${Math.max(0, Math.round(segundos / 60))}m`
}

export function formatarDuracao(segundos: number): string {
  const minutos = Math.max(0, Math.round(segundos / 60))
  const horas = Math.floor(minutos / 60)
  const restantes = minutos % 60
  return horas > 0 ? `${horas}h ${restantes.toString().padStart(2, '0')}m` : `${restantes}m`
}

export function formatarPercentual(valor: number): string {
  return Number.isInteger(valor) ? `${valor}` : valor.toFixed(1)
}

export function formatarData(data: string | null): string {
  if (!data) return 'sem data'
  return new Intl.DateTimeFormat('pt-BR', { timeZone: 'UTC' }).format(new Date(data))
}

export function iniciais(nome: string): string {
  return nome.split(' ').filter(Boolean).slice(0, 2).map((parte) => parte[0]).join('').toUpperCase()
}
