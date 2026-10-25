import { ApiError } from '@/lib/api'

export const ERROS_JOGANDO: Record<string, string> = {
  'jogando.nome_obrigatorio': 'O nome do jogo é obrigatório.',
  'jogando.nome_muito_longo': 'O nome deve ter no máximo 200 caracteres.',
  'jogando.iniciado_em_obrigatorio': 'A data de início é obrigatória.',
  'jogando.iniciado_em_invalido': 'Data de início inválida.',
  'jogando.iniciado_em_futuro': 'A data de início não pode estar no futuro.',
  'jogando.entrada_invalida': 'Dados de entrada inválidos.',
  'jogando.item_duplicado': 'Este jogo já está em andamento.',
}

export const ERRO_GENERICO_JOGANDO = 'Não foi possível iniciar o jogo. Tente novamente.'
export const ERRO_REGISTRO_NAO_ENCONTRADO = 'Registro não encontrado.'
export const ERRO_DADOS_INVALIDOS = 'Dados de entrada inválidos.'

export function obterMensagemErroJogando(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 409) return ERROS_JOGANDO['jogando.item_duplicado']
    if (err.status === 404) return ERRO_REGISTRO_NAO_ENCONTRADO
    if (err.codigo && ERROS_JOGANDO[err.codigo]) return ERROS_JOGANDO[err.codigo]
    if (err.status === 400) return ERRO_DADOS_INVALIDOS
  }
  return ERRO_GENERICO_JOGANDO
}
