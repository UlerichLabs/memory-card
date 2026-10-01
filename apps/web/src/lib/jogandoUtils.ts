export function hojeIso(data = new Date()): string {
  return `${data.getFullYear()}-${String(data.getMonth() + 1).padStart(2, '0')}-${String(data.getDate()).padStart(2, '0')}`
}

export function formatarDataJogando(data: string | null | undefined): string {
  if (!data) return 'sem data'
  const [ano, mes, dia] = data.slice(0, 10).split('-')
  return ano && mes && dia ? `${dia}/${mes}/${ano}` : 'sem data'
}

export function diasDesdeInicio(iniciadoEm: string, hoje = new Date()): number {
  const inicio = iniciadoEm.slice(0, 10).split('-').map(Number)
  const inicioUtc = Date.UTC(inicio[0], inicio[1] - 1, inicio[2])
  const hojeIsoValue = hojeIso(hoje).split('-').map(Number)
  const hojeUtc = Date.UTC(hojeIsoValue[0], hojeIsoValue[1] - 1, hojeIsoValue[2])
  return Math.max(0, Math.floor((hojeUtc - inicioUtc) / 86_400_000))
}

export function rotuloDiasDesdeInicio(iniciadoEm: string, hoje = new Date()): string {
  const dias = diasDesdeInicio(iniciadoEm, hoje)
  if (dias === 0) return 'começou hoje'
  if (dias === 1) return 'há 1 dia'
  return `há ${dias} dias`
}
