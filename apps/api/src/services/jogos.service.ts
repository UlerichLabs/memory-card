import { and, count, desc, eq, gte, ilike, isNull, lte, ne, sql } from "drizzle-orm";
import type { SQL } from "drizzle-orm";
import type { InferSelectModel } from "drizzle-orm";

import { desafioItens, desafios, jogosZerados } from "@memory-card/db/schema";
import {
  AtualizarJogoSchema,
  CriarJogoSchema,
  type AtualizarJogoDto,
  type CriarJogoDto,
  type FiltrosJogoDto,
  type JogoZerado,
  type Nota,
  type PaginacaoResult
} from "@memory-card/types";

import { db } from "../db.js";

type JogoRow = InferSelectModel<typeof jogosZerados>;

export class JogoNotFoundError extends Error {
  constructor() {
    super("Jogo não encontrado");
  }
}

export class DestaqueDuplicadoError extends Error {
  constructor() {
    super("Já existe um destaque para este ano");
  }
}

function toJogo(row: JogoRow): JogoZerado {
  return {
    ...row,
    nota: row.nota as Nota
  };
}

function toDate(value: string | null | undefined) {
  return value ? new Date(value) : null;
}

async function validarDestaque(
  usuarioId: number,
  finalizadoEm: string | undefined,
  destaque: boolean | undefined,
  jogoId?: number
) {
  if (!destaque || !finalizadoEm) {
    return;
  }

  const ano = new Date(finalizadoEm).getUTCFullYear();
  const conditions: SQL[] = [
    eq(jogosZerados.usuarioId, usuarioId),
    eq(jogosZerados.destaque, true),
    isNull(jogosZerados.deletedAt),
    sql`extract(year from ${jogosZerados.finalizadoEm}) = ${ano}`
  ];

  if (jogoId) {
    conditions.push(ne(jogosZerados.id, jogoId));
  }

  const existing = await db.select({ id: jogosZerados.id }).from(jogosZerados).where(and(...conditions)).limit(1);

  if (existing.length > 0) {
    throw new DestaqueDuplicadoError();
  }
}

export async function criarJogo(usuarioId: number, dados: CriarJogoDto): Promise<JogoZerado> {
  const input = CriarJogoSchema.parse(dados);
  await validarDestaque(usuarioId, input.finalizadoEm, input.destaque);

  const [jogo] = await db
    .insert(jogosZerados)
    .values({
      usuarioId,
      igdbId: input.igdbId ?? null,
      nome: input.nome,
      console: input.console,
      genero: input.genero,
      tipo: input.tipo,
      iniciadoEm: toDate(input.iniciadoEm),
      finalizadoEm: new Date(input.finalizadoEm),
      tempoJogado: input.tempoJogado ?? 0,
      nota: input.nota,
      dificuldade: input.dificuldade,
      review: input.review ?? null,
      destaque: input.destaque,
      igdbCapaUrl: input.igdbCapaUrl ?? null,
      igdbDescricao: input.igdbDescricao ?? null
    })
    .returning();

  await sincronizarDesafios(usuarioId, jogo.id);

  return toJogo(jogo);
}

export async function listarJogos(
  usuarioId: number,
  filtros: FiltrosJogoDto
): Promise<PaginacaoResult<JogoZerado>> {
  const conditions: SQL[] = [eq(jogosZerados.usuarioId, usuarioId), isNull(jogosZerados.deletedAt)];

  if (filtros.console) conditions.push(eq(jogosZerados.console, filtros.console));
  if (filtros.genero) conditions.push(eq(jogosZerados.genero, filtros.genero));
  if (filtros.tipo) conditions.push(eq(jogosZerados.tipo, filtros.tipo));
  if (filtros.notaMin) conditions.push(gte(jogosZerados.nota, filtros.notaMin));
  if (filtros.notaMax) conditions.push(lte(jogosZerados.nota, filtros.notaMax));
  if (filtros.dificuldade) conditions.push(eq(jogosZerados.dificuldade, filtros.dificuldade));
  if (filtros.q) conditions.push(ilike(jogosZerados.nome, `%${filtros.q}%`));
  if (filtros.ano) conditions.push(sql`extract(year from ${jogosZerados.finalizadoEm}) = ${filtros.ano}`);

  const where = and(...conditions);
  const offset = (filtros.page - 1) * filtros.limit;
  const [totalRow] = await db.select({ total: count() }).from(jogosZerados).where(where);
  const rows = await db
    .select()
    .from(jogosZerados)
    .where(where)
    .orderBy(desc(jogosZerados.finalizadoEm))
    .limit(filtros.limit)
    .offset(offset);

  const total = totalRow?.total ?? 0;

  return {
    data: rows.map(toJogo),
    page: filtros.page,
    limit: filtros.limit,
    total,
    totalPages: Math.max(1, Math.ceil(total / filtros.limit))
  };
}

export async function buscarJogoPorId(usuarioId: number, jogoId: number): Promise<JogoZerado> {
  const [jogo] = await db
    .select()
    .from(jogosZerados)
    .where(and(eq(jogosZerados.id, jogoId), eq(jogosZerados.usuarioId, usuarioId), isNull(jogosZerados.deletedAt)))
    .limit(1);

  if (!jogo) {
    throw new JogoNotFoundError();
  }

  return toJogo(jogo);
}

export async function atualizarJogo(
  usuarioId: number,
  jogoId: number,
  dados: AtualizarJogoDto
): Promise<JogoZerado> {
  await buscarJogoPorId(usuarioId, jogoId);
  const input = AtualizarJogoSchema.parse(dados);
  await validarDestaque(usuarioId, input.finalizadoEm, input.destaque, jogoId);

  const [jogo] = await db
    .update(jogosZerados)
    .set({
      ...input,
      iniciadoEm: input.iniciadoEm === undefined ? undefined : toDate(input.iniciadoEm),
      finalizadoEm: input.finalizadoEm === undefined ? undefined : new Date(input.finalizadoEm),
      tempoJogado: input.tempoJogado === null ? 0 : input.tempoJogado,
      updatedAt: new Date()
    })
    .where(and(eq(jogosZerados.id, jogoId), eq(jogosZerados.usuarioId, usuarioId), isNull(jogosZerados.deletedAt)))
    .returning();

  await sincronizarDesafios(usuarioId, jogoId);

  return toJogo(jogo);
}

export async function deletarJogo(usuarioId: number, jogoId: number): Promise<void> {
  await buscarJogoPorId(usuarioId, jogoId);
  await db
    .update(jogosZerados)
    .set({ deletedAt: new Date(), updatedAt: new Date() })
    .where(and(eq(jogosZerados.id, jogoId), eq(jogosZerados.usuarioId, usuarioId), isNull(jogosZerados.deletedAt)));
}

export async function sincronizarDesafios(usuarioId: number, jogoId: number): Promise<void> {
  const jogo = await buscarJogoPorId(usuarioId, jogoId);

  if (!jogo.igdbId) {
    return;
  }

  const itens = await db
    .select({ itemId: desafioItens.id, desafioId: desafioItens.desafioId })
    .from(desafioItens)
    .innerJoin(desafios, eq(desafios.id, desafioItens.desafioId))
    .where(
      and(
        eq(desafioItens.igdbId, jogo.igdbId),
        eq(desafioItens.concluido, false),
        eq(desafios.usuarioId, usuarioId),
        isNull(desafios.deletedAt)
      )
    );

  for (const item of itens) {
    await db.update(desafioItens).set({ concluido: true, zeramentoId: jogoId }).where(eq(desafioItens.id, item.itemId));

    const [pendentes] = await db
      .select({ total: count() })
      .from(desafioItens)
      .where(and(eq(desafioItens.desafioId, item.desafioId), eq(desafioItens.concluido, false)));

    if ((pendentes?.total ?? 0) === 0) {
      await db
        .update(desafios)
        .set({ dataConclusao: new Date() })
        .where(and(eq(desafios.id, item.desafioId), eq(desafios.usuarioId, usuarioId), isNull(desafios.dataConclusao)));
    }
  }
}
