import { DIFICULDADE_VALORES } from "./enums.js";

export const dificuldades = DIFICULDADE_VALORES;

export type Dificuldade = (typeof DIFICULDADE_VALORES)[number];

export * from "./enums.js";

export type Nota = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11;

export interface JogoZerado {
  id: number;
  usuarioId: number;
  igdbId: number | null;
  nome: string;
  console: string;
  genero: string | null;
  tipo: string | null;
  iniciadoEm: Date | null;
  finalizadoEm: Date;
  tempoJogado: number;
  nota: Nota;
  dificuldade: Dificuldade;
  review: string | null;
  destaque: boolean;
  igdbCapaUrl: string | null;
  igdbDescricao: string | null;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
}

export type JogoZeradoCreateInput = Omit<JogoZerado, "id" | "createdAt" | "updatedAt" | "deletedAt">;

export type JogoZeradoUpdateInput = Partial<JogoZeradoCreateInput>;

export interface Campanha {
  id: number;
  usuarioId: number;
  nome: string;
  descricao: string | null;
  createdAt: Date;
  icone: string | null;
  concluida: boolean;
  deletedAt: Date | null;
}

export interface Desafio {
  id: number;
  usuarioId: number;
  campanhaId: number | null;
  nome: string;
  tipo: "QUANTIDADE" | "LISTA_COMPLETA" | "FRANQUIA" | "EMPRESA" | "GENERO" | "MANUAL";
  filtroValor: string | null;
  meta: number | null;
  dataConclusao: Date | null;
  createdAt: Date;
  deletedAt: Date | null;
}

export interface DesafioItem {
  id: number;
  desafioId: number;
  igdbId: number;
  nome: string;
  concluido: boolean;
  zeramentoId: number | null;
}

export interface Conquista {
  id: number;
  usuarioId: number;
  desafioId: number | null;
  nome: string;
  descricao: string | null;
  dataConquista: Date;
  createdAt: Date;
}

export interface Favorito {
  id: number;
  usuarioId: number;
  tipo: "JOGO" | "GENERO";
  jogoId: string;
  posicao: number;
  createdAt: Date;
}

export type Usuario = {
  id: number;
  nome: string;
  username: string;
  email: string;
  avatar?: string | null;
  bio?: string | null;
  redesSociais?: {
    youtube?: string;
    instagram?: string;
    twitch?: string;
    x?: string;
  } | null;
  createdAt: Date;
  updatedAt: Date;
};

export type AuthTokens = {
  accessToken: string;
};

export type JwtPayload = {
  sub: number;
  usuario_id: number;
  email: string;
  username: string;
};

export * from "./igdb.js";
export * from "./jogos.schemas.js";
