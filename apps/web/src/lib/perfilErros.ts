export const PERFIL_BACKEND_ERROS: Record<string, { campo?: string; mensagem: string }> = {
  'auth.nome_obrigatorio': { campo: 'nome', mensagem: 'O nome é obrigatório.' },
  'auth.nome_muito_longo': { campo: 'nome', mensagem: 'O nome deve ter no máximo 100 caracteres.' },
  'perfil.nome_obrigatorio': { campo: 'nome', mensagem: 'O nome é obrigatório.' },
  'perfil.nome_invalido': { campo: 'nome', mensagem: 'O nome deve ter no máximo 100 caracteres.' },
  'perfil.username_invalido': { campo: 'username', mensagem: 'Username deve ter entre 3 e 30 caracteres e conter apenas letras, números e _.' },
  'perfil.username_ja_existe': { campo: 'username', mensagem: 'Este nome de usuário já está em uso.' },
  'perfil.username_em_uso': { campo: 'username', mensagem: 'Este nome de usuário já está em uso.' },
  'perfil.bio_muito_longa': { campo: 'bio', mensagem: 'A bio deve ter no máximo 280 caracteres.' },
  'perfil.console_muito_longo': { campo: 'console_favorito', mensagem: 'O console favorito deve ter no máximo 100 caracteres.' },
  'perfil.console_invalido': { campo: 'console_favorito', mensagem: 'O console favorito deve ter no máximo 100 caracteres.' },
  'perfil.ano_invalido': { campo: 'jogando_desde', mensagem: 'O ano deve ser entre 1970 e o ano atual.' },
  'perfil.jogando_desde_invalido': { campo: 'jogando_desde', mensagem: 'O ano informado em jogando desde é inválido.' },
  'perfil.jogo_nao_encontrado': { campo: 'jogo_favorito_id', mensagem: 'O jogo selecionado não existe ou não pertence a você.' },
  'perfil.jogo_favorito_nao_encontrado': { campo: 'jogo_favorito_id', mensagem: 'Jogo favorito não encontrado.' },
  'perfil.jogo_sem_capa': { campo: 'jogo_favorito_id', mensagem: 'O jogo selecionado não possui capa cadastrada.' },
  'perfil.entrada_invalida': { campo: 'geral', mensagem: 'Dados de entrada inválidos.' },
}

export const FOTO_ERRO_MENSAGENS: Record<string, string> = {
  'perfil.arquivo_obrigatorio': 'Selecione um arquivo de imagem.',
  'perfil.arquivo_muito_grande': 'A foto deve ter no máximo 2 MB.',
  'perfil.tipo_arquivo_invalido': 'Apenas JPG, PNG ou WebP.',
  'perfil.imagem_invalida': 'Não foi possível processar a imagem.',
  'perfil.jogo_nao_encontrado': 'O jogo selecionado não existe ou não pertence a você.',
  'perfil.jogo_sem_capa': 'O jogo selecionado não possui capa cadastrada.',
}

export const ERRO_GENERICO_PERFIL = 'Não foi possível salvar as alterações. Tente novamente.'
export const ERRO_GENERICO_FOTO = 'Não foi possível atualizar a foto. Tente novamente.'
