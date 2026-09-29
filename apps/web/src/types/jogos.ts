export type Dificuldade = 'C' | 'B' | 'A' | 'AA' | 'AAA'

export interface JogoZeradoDTO {
  id: number
  numero?: number
  usuario_id: number
  igdb_id?: number | null
  nome: string
  console: string
  genero?: string
  tipo?: string
  iniciado_em?: string | null
  finalizado_em: string
  tempo_jogado: number
  nota: number
  dificuldade: Dificuldade
  review?: string | null
  destaque: boolean
  igdb_capa_url?: string
  igdb_descricao?: string
  created_at?: string
  updated_at?: string
}

export interface ListagemMeta {
  pagina: number
  por_pagina: number
  total: number
  total_paginas: number
}

export interface ListarJogosResposta {
  data: JogoZeradoDTO[]
  meta: ListagemMeta
}

export type OrdenacaoJogos = 'recentes' | 'nota'

export interface ListarJogosParams {
  pagina?: number
  por_pagina?: number
  busca?: string
  console?: string
  genero?: string
  tipo?: string
  nota_min?: number
  nota_max?: number
  ano?: number
  dificuldade?: Dificuldade
  ordenar?: OrdenacaoJogos
}

export interface ResumoGameDoAnoItem {
  ano: number
  total_jogos: number
  game_do_ano: JogoZeradoDTO | null
}

export interface DefinirGameDoAnoResposta {
  ano: number
  anterior_id: number | null
  game_do_ano: JogoZeradoDTO
}

export interface OpcoesFiltrosDTO {
  consoles: string[]
  generos: string[]
  tipos: string[]
  anos: number[]
}

export interface SalvarJogoPayload {
  igdb_id?: number | null
  nome: string
  console: string
  genero?: string
  tipo?: string
  iniciado_em?: string | null
  finalizado_em: string
  tempo_jogado_horas?: number
  tempo_jogado_minutos?: number
  tempo_jogado_segundos?: number
  tempo_jogado?: number
  nota: number
  dificuldade: Dificuldade
  review?: string | null
  destaque: boolean
  igdb_capa_url?: string
  igdb_descricao?: string
}

export interface IGDBJogoSugestao {
  id: number
  name: string
  cover?: { id?: number; url?: string }
  first_release_date?: number
  summary?: string
  platforms?: Array<{ id: number; name: string }>
  genres?: Array<{ id: number; name: string }>
}
