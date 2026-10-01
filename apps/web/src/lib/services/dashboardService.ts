import { ApiError } from '@/lib/api'
import { abandonadosService } from './abandonadosService'
import { jogosService } from './jogosService'
import { listasService } from './listasService'
import type {
  DashboardAno, DashboardDificuldade, DashboardNotas, DashboardRankingGenero,
  DashboardRankingPlataforma, DashboardRecordes, DashboardResumo, DashboardTipo,
} from '@/types/dashboard'

const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')

async function direto<T>(path: string, token?: string, signal?: AbortSignal): Promise<T> {
  const headers = new Headers({ Accept: 'application/json' })
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const response = await fetch(`${baseUrl}/api/v1${path}`, { headers, signal })
  if (!response.ok) {
    let codigo = 'fallback'
    let mensagem = ''
    try {
      const body = await response.json()
      codigo = typeof body?.error?.codigo === 'string' ? body.error.codigo : codigo
      mensagem = typeof body?.error?.mensagem === 'string' ? body.error.mensagem : mensagem
    } catch { }
    throw new ApiError(codigo, mensagem, response.status)
  }
  return response.json() as Promise<T>
}

export const dashboardService = {
  resumo: (token?: string, signal?: AbortSignal) => direto<DashboardResumo>('/dashboard/resumo', token, signal),
  porAno: (token?: string, signal?: AbortSignal) => direto<DashboardAno[]>('/dashboard/por-ano', token, signal),
  plataformas: (token?: string, signal?: AbortSignal) => direto<DashboardRankingPlataforma[]>('/dashboard/ranking-plataformas?limite=5', token, signal),
  generos: (token?: string, signal?: AbortSignal) => direto<DashboardRankingGenero[]>('/dashboard/ranking-generos?limite=5', token, signal),
  tipos: (genero: string, token?: string, signal?: AbortSignal) => direto<DashboardTipo[]>(`/dashboard/breakdown-tipo?genero=${encodeURIComponent(genero)}`, token, signal),
  recordes: (token?: string, signal?: AbortSignal) => direto<DashboardRecordes>('/dashboard/recordes', token, signal),
  notas: (token?: string, signal?: AbortSignal) => direto<DashboardNotas>('/dashboard/notas', token, signal),
  dificuldade: (token?: string, signal?: AbortSignal) => direto<DashboardDificuldade[]>('/dashboard/dificuldade', token, signal),
  abandonados: (token?: string, signal?: AbortSignal) => abandonadosService.obterTotal(token, signal).then((res) => res.total),
  jogoDoAno: (token?: string, signal?: AbortSignal) => jogosService.obterResumoGameDoAno(token, signal),
  jogosDaVida: (token?: string, signal?: AbortSignal) => jogosService.listar({ nota_min: 11, por_pagina: 5, ordenar: 'nota' }, token, signal).then((res) => res.data),
  recentes: (token?: string, signal?: AbortSignal) => jogosService.listar({ por_pagina: 4, ordenar: 'recentes' }, token, signal).then((res) => res.data),
  desafios: (token?: string, signal?: AbortSignal) => listasService.listar(token, signal).then((listas) => listas.filter((lista) => lista.tipo === 'desafio' && lista.progresso && !lista.progresso.concluido).slice(0, 3)),
}
