import type { ListagemMeta } from './jogos'

export type OrdenacaoAbandonados = 'recentes' | 'antigos' | 'nome' | 'tempo'
export type OrdenarAbandonados = OrdenacaoAbandonados
export type AbandonadosListagemMeta = ListagemMeta

export interface JogoAbandonado {
  id: number
  nome: string
  console: string
  igdb_id: number | null
  igdb_capa_url: string | null
  tempo_jogado: number
  motivo: string | null
  abandonado_em: string
  created_at: string
  updated_at: string
}

export interface SalvarAbandonadoPayload {
  nome: string
  console: string
  igdb_id?: number | null
  igdb_capa_url?: string | null
  tempo_jogado_horas?: number
  tempo_jogado_minutos?: number
  tempo_jogado_segundos?: number
  motivo?: string | null
  abandonado_em?: string
}

export interface ListarAbandonadosParams {
  busca?: string
  console?: string
  ordenar?: OrdenacaoAbandonados
  pagina?: number
  por_pagina?: number
}

export interface ListarAbandonadosResposta {
  data: JogoAbandonado[]
  meta: ListagemMeta
}

export interface OpcoesFiltrosAbandonados {
  consoles: string[]
}

export interface TotalAbandonadosResposta {
  total: number
}
