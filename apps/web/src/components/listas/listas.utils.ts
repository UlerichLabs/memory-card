import type { ListaResumo } from '@/types/listas'
import { LISTAS_ERROS, LISTAS_ERRO_GENERICO } from './listas.constants'

export function obterIniciaisJogo(nome: string): string {
  const limpo = nome.trim().replace(/^the\s+/i, '').trim()
  const palavras = limpo.split(/\s+/).filter(Boolean)
  if (palavras.length === 0) return '?'
  if (palavras.length === 1) return palavras[0].slice(0, 2).toUpperCase()
  return (palavras[0][0] + palavras[1][0]).toUpperCase()
}

export function ordenarListas(listas: ListaResumo[]): ListaResumo[] {
  const desafios = listas
    .filter((l) => l.tipo === 'desafio')
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  const filas = listas
    .filter((l) => l.tipo === 'fila')
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  return [...desafios, ...filas]
}

export function formatarFaltam(faltam: number): string {
  if (faltam === 1) {
    return 'Falta 1 · o progresso conta sozinho quando você registra um zeramento que se encaixa'
  }
  return `Faltam ${faltam} · o progresso conta sozinho quando você registra um zeramento que se encaixa`
}

export function formatarSubFila(pendentes: number): string {
  if (pendentes === 1) return '1 jogo na fila'
  return `${pendentes} jogos na fila`
}

export function mapearErroApiParaCampo(codigo: string): { campo: string; mensagem: string } {
  const campos: Record<string, string> = {
    'listas.nome_obrigatorio': 'nome',
    'listas.nome_invalido': 'nome',
    'listas.descricao_muito_longa': 'descricao',
    'listas.meta_invalida': 'meta',
    'listas.regra_invalida': 'regraValor',
    'listas.franquia_nao_encontrada': 'regraValor',
  }
  const campo = campos[codigo] ?? 'form'
  return { campo, mensagem: LISTAS_ERROS[codigo] ?? LISTAS_ERRO_GENERICO }
}

export function versaoLabelIGDB(ano?: number | null, plataformas?: Array<{ name: string }>): string {
  const plat = plataformas?.map((p) => p.name).join(', ')
  return [ano, plat].filter(Boolean).join(' · ')
}
