export interface JogoEmAndamento {
  id: number
  nome: string
  igdb_id: number | null
  igdb_capa_url: string | null
  iniciado_em: string
}

export interface CriarJogoEmAndamentoPayload {
  nome: string
  igdb_id?: number | null
  igdb_capa_url?: string | null
  iniciado_em: string
}

export interface ListarJogandoResposta {
  data: JogoEmAndamento[]
}
