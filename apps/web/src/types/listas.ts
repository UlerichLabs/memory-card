export type ListaTipo = "fila" | "desafio";

export type OrigemTipo = "franquia" | "plataforma" | "genero";

export interface ListaOrigem {
  tipo: OrigemTipo;
  igdb_id: number | null;
  nome: string;
}

export interface ListaProgresso {
  feitos: number;
  meta: number;
  percentual: number;
  concluido: boolean;
  concluido_em: string | null;
}

export interface ListaResumo {
  id: number;
  tipo: ListaTipo;
  nome: string;
  descricao: string | null;
  origem: ListaOrigem | null;
  total_itens: number;
  itens_pendentes: number;
  progresso: ListaProgresso | null;
  created_at: string;
  updated_at: string;
}

export interface JogoZeradoVinculado {
  id: number;
  nota: number;
  finalizado_em: string;
}

export interface ListaItem {
  id: number;
  igdb_id: number | null;
  nome: string;
  console: string | null;
  igdb_capa_url: string | null;
  ano_lancamento: number | null;
  posicao: number;
  origem: "item" | "regra";
  zerado: boolean;
  jogo_zerado: JogoZeradoVinculado | null;
}

export interface ListaDetalhada extends ListaResumo {
  itens: ListaItem[];
}

export interface CatalogoItem {
  igdb_id: number;
  nome: string;
  igdb_capa_url: string | null;
  ano_lancamento: number | null;
  sugerido: boolean;
  ja_zerado: boolean;
  jogo_zerado_id: number | null;
}

export interface CatalogoMeta {
  pagina: number;
  por_pagina: number;
  total: number;
  total_sugeridos: number | null;
}

export interface CatalogoResposta {
  itens: CatalogoItem[];
  meta: CatalogoMeta;
}

export interface OpcaoOrigem {
  id: number;
  nome: string;
}

export interface CriarListaItemPayload {
  igdb_id: number;
  nome: string;
  igdb_capa_url?: string | null;
  ano_lancamento?: number | null;
}

export interface CriarListaPayload {
  tipo: ListaTipo;
  nome: string;
  descricao?: string | null;
  origem?: ListaOrigem | null;
  itens?: CriarListaItemPayload[];
}

export interface AtualizarListaPayload {
  nome?: string;
  descricao?: string | null;
}

export interface AdicionarItemPayload {
  igdb_id?: number | null;
  nome: string;
  console?: string | null;
  igdb_capa_url?: string | null;
  ano_lancamento?: number | null;
}

export interface AdicionarItensLoteResposta {
  adicionados: number;
  ja_existentes: number;
}

export interface AdicionarItensLotePayload {
  itens: CriarListaItemPayload[];
}

export interface IGDBFranquiaSugestao {
  id: number;
  name: string;
}

export type FiltroAba = "todos" | "zerados" | "pendentes";

export type OrdenarCatalogo = "populares" | "lancamento" | "nome";

export interface ListasStore {
  listas: ListaResumo[];
  listaAberta: ListaDetalhada | null;
  isLoading: boolean;
  isLoadingDetalhe: boolean;
  error: string | null;
  filtroAba: FiltroAba;
  isNovaListaOpen: boolean;
  listaEmEdicao: ListaResumo | null;
  isExcluirListaOpen: boolean;
  listaParaExcluir: ListaResumo | null;
  setFiltroAba: (aba: FiltroAba) => void;
  limparErro: () => void;
  abrirModalCriar: () => void;
  abrirModalEditar: (lista: ListaResumo) => void;
  fecharModalNovaLista: () => void;
  abrirModalExcluir: (lista: ListaResumo) => void;
  fecharModalExcluir: () => void;
  carregarListas: (signal?: AbortSignal) => Promise<ListaResumo[]>;
  abrirLista: (id: number, signal?: AbortSignal) => Promise<ListaDetalhada>;
  criarLista: (payload: CriarListaPayload) => Promise<ListaDetalhada>;
  atualizarLista: (
    id: number,
    payload: AtualizarListaPayload,
  ) => Promise<ListaDetalhada>;
  excluirLista: (id: number) => Promise<void>;
  adicionarItem: (payload: AdicionarItemPayload) => Promise<ListaItem>;
  adicionarItensLote: (
    itens: CriarListaItemPayload[],
  ) => Promise<AdicionarItensLoteResposta>;
  removerItem: (itemId: number) => Promise<void>;
  reordenarItens: (itemIds: number[]) => Promise<void>;
  associarZeramento: (itemId: number, jogoZeradoId: number) => Promise<void>;
}

export interface FiltroCatalogo {
  origem: OrigemTipo;
  id: number;
  genero_id?: number;
  plataforma_id?: number;
  busca?: string;
  ordenar?: OrdenarCatalogo;
  pagina?: number;
  por_pagina?: number;
}
