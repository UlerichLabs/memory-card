export interface IgdbJogo {
  igdbId: number;
  nome: string;
  capaUrl: string | null;
  ano: number | null;
  generos: string[];
  plataformas: string[];
}

export interface IgdbJogoDetalhe extends IgdbJogo {
  descricao: string | null;
  desenvolvedor: string | null;
}
