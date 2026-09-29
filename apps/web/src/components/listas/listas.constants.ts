export const LISTAS_ERROS: Record<string, string> = {
  'listas.nome_obrigatorio': 'O nome da lista é obrigatório.',
  'listas.nome_invalido': 'Nome inválido. Deve ter entre 1 e 100 caracteres.',
  'listas.descricao_muito_longa': 'A descrição deve ter no máximo 200 caracteres.',
  'listas.tipo_invalido': 'Tipo de lista inválido.',
  'listas.meta_invalida': 'A meta deve ser um número inteiro positivo.',
  'listas.regra_invalida': 'Selecione uma regra válida para o desafio.',
  'listas.franquia_nao_encontrada': 'Franquia não encontrada no IGDB.',
  'listas.item_duplicado': 'Esse jogo já está na lista.',
  'listas.item_nao_encontrado': 'Jogo não encontrado na lista.',
  'listas.nao_encontrada': 'Lista não encontrada.',
  'listas.ordem_invalida': 'Não foi possível salvar a nova ordem dos jogos.',
  'listas.itens_nao_permitidos': 'Esta lista não aceita adição manual de jogos.',
  'jogos.not_found': 'Jogo zerado não encontrado na Biblioteca.',
  'igdb.rate_limited': 'IGDB indisponível no momento, tente de novo.',
  'igdb.unavailable': 'IGDB indisponível no momento, tente de novo.',
}

export const LISTAS_ERRO_GENERICO = 'Ocorreu um erro. Tente novamente mais tarde.'

export function resolverMensagemErro(codigo?: string): string {
  if (!codigo) return LISTAS_ERRO_GENERICO
  return LISTAS_ERROS[codigo] ?? LISTAS_ERRO_GENERICO
}

export const ROTULOS_REGRA: Record<string, string> = {
  franquia: 'Franquia via IGDB',
  plataforma: 'Contagem por plataforma',
  genero: 'Contagem por gênero',
  manual: 'Jogos escolhidos',
}

export const ROTULOS_METABOX: Record<string, string> = {
  franquia: 'jogos da franquia',
  plataforma: 'jogos zerados na plataforma',
  genero: 'jogos zerados no gênero',
  manual: 'jogos escolhidos',
}

export const AJUDA_REGRA: Record<string, string> = {
  franquia: 'Os jogos da franquia são trazidos do IGDB. Os que você já zerou contam na hora.',
  plataforma: 'Não tem lista fixa: todo jogo que você zerar nessa plataforma conta, inclusive os que já estão na sua Biblioteca.',
  genero: 'Todo zeramento com esse gênero conta, inclusive os antigos.',
  manual: 'Você adiciona os jogos depois de criar. A meta pode ser todos ou só uma parte.',
}
