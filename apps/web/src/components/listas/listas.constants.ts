export const LISTAS_ERROS: Record<string, string> = {
  "listas.nome_obrigatorio": "O nome da lista é obrigatório.",
  "listas.nome_invalido": "Nome inválido. Deve ter entre 1 e 100 caracteres.",
  "listas.descricao_muito_longa":
    "A descrição deve ter no máximo 200 caracteres.",
  "listas.tipo_invalido": "Tipo de lista inválido.",
  "listas.item_duplicado": "Esse jogo já está na lista.",
  "listas.item_nao_encontrado": "Jogo não encontrado na lista.",
  "listas.nao_encontrada": "Lista não encontrada.",
  "listas.ordem_invalida": "Não foi possível salvar a nova ordem dos jogos.",
  "listas.desafio_sem_jogos": "Escolha pelo menos um jogo.",
  "listas.itens_demais": "Máximo de 1.000 jogos por desafio.",
  "listas.campo_nao_permitido": "Este campo não é permitido.",
  "listas.franquia_nao_encontrada": "Franquia não encontrada.",
  "catalogo.parametro_invalido": "Parâmetro de catálogo inválido.",
  "catalogo.paginacao_invalida": "Paginação do catálogo inválida.",
  "igdb.rate_limited":
    "Não foi possível carregar os jogos agora. Tente de novo.",
  "igdb.unavailable":
    "Não foi possível carregar os jogos agora. Tente de novo.",
  "jogos.not_found": "Jogo zerado não encontrado na Biblioteca.",
};

export const LISTAS_ERRO_GENERICO =
  "Ocorreu um erro. Tente novamente mais tarde.";

export function resolverMensagemErro(codigo?: string): string {
  return (codigo && LISTAS_ERROS[codigo]) || LISTAS_ERRO_GENERICO;
}

export const ROTULOS_ORIGEM: Record<string, string> = {
  franquia: "Franquia",
  plataforma: "Plataforma",
  genero: "Gênero",
};

export const AJUDA_ORIGEM: Record<string, string> = {
  franquia:
    "Na próxima tela aparecem todos os jogos da franquia pra você escolher.",
  plataforma:
    "Na próxima tela aparecem os jogos da plataforma, dos mais populares aos menos.",
  genero:
    "Na próxima tela aparecem os jogos do gênero. Dá pra filtrar por plataforma lá.",
};
