import { z } from "zod";

import { DIFICULDADE_VALORES } from "./enums.js";

export const CriarJogoSchema = z.object({
  igdbId: z.number().int().optional().nullable(),
  nome: z.string().min(1).max(300),
  console: z.string().min(1).max(100),
  genero: z.string().min(1).max(100),
  tipo: z.string().min(1).max(100),
  iniciadoEm: z.string().datetime().optional().nullable(),
  finalizadoEm: z.string().datetime(),
  tempoJogado: z.number().int().min(0).optional().nullable(),
  nota: z.number().int().min(1).max(11),
  dificuldade: z.enum(DIFICULDADE_VALORES),
  review: z.string().max(1000).optional().nullable(),
  destaque: z.boolean().default(false),
  igdbCapaUrl: z.string().url().optional().nullable(),
  igdbDescricao: z.string().optional().nullable()
});

export type CriarJogoDto = z.infer<typeof CriarJogoSchema>;
export type AtualizarJogoDto = Partial<CriarJogoDto>;

export const AtualizarJogoSchema = CriarJogoSchema.partial();

export const FiltrosJogoSchema = z.object({
  console: z.string().optional(),
  genero: z.string().optional(),
  tipo: z.string().optional(),
  notaMin: z.coerce.number().int().min(1).max(11).optional(),
  notaMax: z.coerce.number().int().min(1).max(11).optional(),
  ano: z.coerce.number().int().optional(),
  dificuldade: z.enum(DIFICULDADE_VALORES).optional(),
  q: z.string().optional(),
  page: z.coerce.number().int().min(1).default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20)
});

export type FiltrosJogoDto = z.infer<typeof FiltrosJogoSchema>;

export interface PaginacaoResult<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}
