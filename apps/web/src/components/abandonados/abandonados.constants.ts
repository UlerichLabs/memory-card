export const ABANDONADOS_CAMPO_ERRO_MENSAGENS: Record<string, { campo: string; mensagem: string }> = {
  'abandonados.nome_obrigatorio': { campo: 'nome', mensagem: 'O nome do jogo é obrigatório.' },
  'abandonados.nome_muito_longo': { campo: 'nome', mensagem: 'O nome deve ter no máximo 200 caracteres.' },
  'abandonados.console_obrigatorio': { campo: 'console', mensagem: 'A plataforma é obrigatória.' },
  'abandonados.console_muito_longo': { campo: 'console', mensagem: 'A plataforma deve ter no máximo 100 caracteres.' },
  'abandonados.tempo_invalido': { campo: 'tempo_jogado_horas', mensagem: 'O tempo jogado é inválido.' },
  'abandonados.motivo_muito_longo': { campo: 'motivo', mensagem: 'O motivo deve ter no máximo 500 caracteres.' },
  'abandonados.data_invalida': { campo: 'abandonado_em', mensagem: 'Data inválida.' },
  'abandonados.data_futura': { campo: 'abandonado_em', mensagem: 'A data não pode estar no futuro.' },
  'abandonados.nao_encontrado': { campo: 'form', mensagem: 'Jogo abandonado não encontrado.' },
  'abandonados.entrada_invalida': { campo: 'form', mensagem: 'Dados inválidos.' },
}

export const ABANDONADOS_ERRO_GENERICO = 'Não foi possível salvar o registro. Tente novamente.'

export const ERRO_FALHA_REMOCAO_FILA =
  'Jogo abandonado, mas não foi possível removê-lo da fila. Tente novamente.'

export const AVISO_FALHA_REMOCAO_RETOMAR =
  'Zeramento registrado, mas não foi possível remover este jogo dos abandonados. Exclua-o manualmente.'

export const ERRO_JOGO_NAO_ENCONTRADO = 'Jogo não encontrado.'
