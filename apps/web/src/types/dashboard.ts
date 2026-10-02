import type { JogoZeradoDTO } from './jogos'
import type { ListaResumo } from './listas'

export interface DashboardResumo {
  total_jogos: number
  total_segundos: number
  media_segundos_por_jogo: number
  nota_media: number
  jogos_no_ano_atual: number
  primeiro_zeramento_em: string | null
  dias_desde_primeiro: number
  anos_desde_primeiro: number
}

export interface DashboardAno {
  ano: number
  total_jogos: number
  total_segundos: number
  game_do_ano: { id: number; nome: string; console: string; igdb_capa_url: string; nota: number } | null
}
export interface DashboardGameDoAno {
  id: number
  nome: string
  console: string
  igdb_capa_url: string
  nota: number
  ano: number
  tempo_jogado?: number
}

export interface DashboardJogosDaVida {
  jogos: JogoZeradoDTO[]
  total: number
}

export interface DashboardRankingPlataforma {
  console: string
  total_jogos: number
  total_segundos: number
  percentual_jogos: number
  percentual_segundos: number
}

export interface DashboardRankingGenero {
  genero: string
  total_jogos: number
  total_segundos: number
  percentual_jogos: number
  percentual_segundos: number
}

export interface DashboardTipo { tipo: string; total_jogos: number }
export interface DashboardRecordeJogo { id: number; nome: string; console: string; ano: number; igdb_capa_url: string; tempo_jogado_segundos: number }
export interface DashboardRecordes { mais_longo: DashboardRecordeJogo | null; mais_curto: DashboardRecordeJogo | null }
export interface DashboardNota { nota: number; total: number }
export interface DashboardNotas { histograma: DashboardNota[]; nota_media: number; total_avaliados: number }
export interface DashboardDificuldade { dificuldade: 'C' | 'B' | 'A' | 'AA' | 'AAA'; total_jogos: number; percentual: number }

export interface DashboardDados {
  resumo: DashboardResumo | null
  abandonados: number
  jogoDoAno: DashboardGameDoAno | null
  jogosDaVida: JogoZeradoDTO[]
  recentes: JogoZeradoDTO[]
  desafios: ListaResumo[]
  porAno: DashboardAno[]
  plataformas: DashboardRankingPlataforma[]
  generos: DashboardRankingGenero[]
  tipos: DashboardTipo[]
  notas: DashboardNotas | null
  dificuldade: DashboardDificuldade[]
  recordes: DashboardRecordes | null
}

export type DashboardBloco = keyof DashboardDados
