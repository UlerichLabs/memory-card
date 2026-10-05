import type { JogoEmAndamento } from '@/types/jogando'

export function hojeIso(data = new Date()): string {
  return `${data.getFullYear()}-${String(data.getMonth() + 1).padStart(2, '0')}-${String(data.getDate()).padStart(2, '0')}`
}

export function formatarDataJogando(data: string | null | undefined): string {
  if (!data) return 'sem data'
  const [ano, mes, dia] = data.slice(0, 10).split('-')
  return ano && mes && dia ? `${dia}/${mes}/${ano}` : 'sem data'
}

export function diasDesdeInicio(iniciadoEm: string, hoje = new Date()): number {
  const ini = iniciadoEm.slice(0, 10).split('-').map(Number), h = hojeIso(hoje).split('-').map(Number)
  return Math.max(0, Math.floor((Date.UTC(h[0], h[1] - 1, h[2]) - Date.UTC(ini[0], ini[1] - 1, ini[2])) / 86_400_000))
}

export function rotuloDiasDesdeInicio(iniciadoEm: string, hoje = new Date()): string {
  const dias = diasDesdeInicio(iniciadoEm, hoje)
  return dias === 0 ? 'começou hoje' : dias === 1 ? 'há 1 dia' : `há ${dias} dias`
}

export function normalizarNomeJogo(nome: string): string {
  return nome.trim().toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '')
}

export interface CriterioBuscaAndamento {
  nome?: string | null
  igdbId?: number | null
}

export function encontrarJogoEmAndamento(
  jogos: JogoEmAndamento[],
  criterio: CriterioBuscaAndamento,
): JogoEmAndamento | undefined {
  const alvoIgdbId = criterio.igdbId ?? null
  const alvoNomeNorm = criterio.nome?.trim() ? normalizarNomeJogo(criterio.nome) : null
  const matches: { jogo: JogoEmAndamento; prioridade: number }[] = []
  for (const jogo of jogos) {
    if (alvoIgdbId !== null && jogo.igdb_id !== null) {
      if (jogo.igdb_id === alvoIgdbId) matches.push({ jogo, prioridade: 0 })
      continue
    }
    if (alvoNomeNorm && normalizarNomeJogo(jogo.nome) === alvoNomeNorm) {
      matches.push({ jogo, prioridade: 1 })
    }
  }
  if (matches.length === 0) return undefined
  matches.sort((a, b) => {
    if (a.prioridade !== b.prioridade) return a.prioridade - b.prioridade
    const tempoA = new Date(a.jogo.iniciado_em).getTime(), tempoB = new Date(b.jogo.iniciado_em).getTime()
    const diff = !Number.isNaN(tempoA) && !Number.isNaN(tempoB) ? tempoA - tempoB : a.jogo.iniciado_em.localeCompare(b.jogo.iniciado_em)
    return diff || a.jogo.id - b.jogo.id
  })
  return matches[0].jogo
}
