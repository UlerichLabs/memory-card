import { and, asc, desc, eq, isNull, sql } from "drizzle-orm";
import type { InferSelectModel } from "drizzle-orm";

import { jogosZerados } from "@memory-card/db/schema";
import type { Dificuldade, Nota } from "@memory-card/types";

import { db } from "../db.js";

type JogoRow = InferSelectModel<typeof jogosZerados>;

export type DashboardJogo = {
  id: number;
  nome: string;
  console: string;
  nota: Nota;
  dificuldade: Dificuldade;
  finalizadoEm: Date;
  tempoJogado: number;
  igdbCapaUrl: string | null;
};

function baseWhere(usuarioId: number) {
  return and(eq(jogosZerados.usuarioId, usuarioId), isNull(jogosZerados.deletedAt));
}

function toDashboardJogo(row: JogoRow): DashboardJogo {
  return {
    id: row.id,
    nome: row.nome,
    console: row.console,
    nota: row.nota as Nota,
    dificuldade: row.dificuldade as Dificuldade,
    finalizadoEm: row.finalizadoEm,
    tempoJogado: row.tempoJogado,
    igdbCapaUrl: row.igdbCapaUrl
  };
}

export async function getTotais(usuarioId: number) {
  const [row] = await db
    .select({
      totalJogos: sql<number>`count(*)::int`,
      totalSegundos: sql<number>`coalesce(sum(${jogosZerados.tempoJogado}), 0)::int`,
      notaMedia: sql<number>`coalesce(avg(${jogosZerados.nota}), 0)::float`,
      totalConsoles: sql<number>`count(distinct ${jogosZerados.console})::int`,
      primeiroZeramento: sql<Date | null>`min(${jogosZerados.finalizadoEm})`
    })
    .from(jogosZerados)
    .where(baseWhere(usuarioId));

  return row;
}

export async function getPorAno(usuarioId: number) {
  const anoExpr = sql<number>`extract(year from ${jogosZerados.finalizadoEm})::int`;
  const rows = await db
    .select({
      ano: anoExpr,
      totalJogos: sql<number>`count(*)::int`,
      totalSegundos: sql<number>`coalesce(sum(${jogosZerados.tempoJogado}), 0)::int`
    })
    .from(jogosZerados)
    .where(baseWhere(usuarioId))
    .groupBy(sql`extract(year from ${jogosZerados.finalizadoEm})`)
    .orderBy(asc(anoExpr));

  const destaques = await db
    .select()
    .from(jogosZerados)
    .where(and(baseWhere(usuarioId), eq(jogosZerados.destaque, true)))
    .orderBy(asc(jogosZerados.finalizadoEm));
  const destaquesPorAno = new Map(destaques.map((jogo) => [jogo.finalizadoEm.getUTCFullYear(), toDashboardJogo(jogo)]));

  return rows.map((row) => ({
    ...row,
    destaque: destaquesPorAno.get(row.ano) ?? null
  }));
}

export async function getPorPlataforma(usuarioId: number) {
  return db
    .select({
      console: jogosZerados.console,
      total: sql<number>`count(*)::int`
    })
    .from(jogosZerados)
    .where(baseWhere(usuarioId))
    .groupBy(jogosZerados.console)
    .orderBy(desc(sql`count(*)`))
    .limit(5);
}

export async function getRecordes(usuarioId: number) {
  const where = and(baseWhere(usuarioId), sql`${jogosZerados.tempoJogado} > 0`);
  const [maisLongo] = await db
    .select()
    .from(jogosZerados)
    .where(where)
    .orderBy(desc(jogosZerados.tempoJogado))
    .limit(1);
  const [maisCurto] = await db
    .select()
    .from(jogosZerados)
    .where(where)
    .orderBy(asc(jogosZerados.tempoJogado))
    .limit(1);

  return {
    maisLongo: maisLongo ? toDashboardJogo(maisLongo) : null,
    maisCurto: maisCurto ? toDashboardJogo(maisCurto) : null
  };
}
