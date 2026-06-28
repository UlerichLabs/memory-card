# Queries críticas

Use estes exemplos como ponto de partida para services em `apps/api/src/services/`.

## Totais do dashboard

```ts
import { sql } from "drizzle-orm";
import { jogosZerados } from "@memory-card/db/schema";
import { db } from "../db.js";

export async function buscarTotaisDashboard() {
  const [totais] = await db
    .select({
      totalJogos: sql<number>`count(*)::int`,
      tempoTotalSegundos: sql<number>`coalesce(sum(${jogosZerados.tempoSegundos}), 0)::int`,
      mediaNota: sql<number>`coalesce(avg(${jogosZerados.nota}), 0)::float`
    })
    .from(jogosZerados);

  return totais;
}
```

## Por ano

```ts
import { sql } from "drizzle-orm";
import { jogosZerados } from "@memory-card/db/schema";
import { db } from "../db.js";

export async function buscarDashboardPorAno() {
  return db
    .select({
      ano: sql<number>`extract(year from ${jogosZerados.finalizadoEm})::int`,
      total: sql<number>`count(*)::int`
    })
    .from(jogosZerados)
    .groupBy(sql`extract(year from ${jogosZerados.finalizadoEm})`)
    .orderBy(sql`extract(year from ${jogosZerados.finalizadoEm})`);
}
```

## Por plataforma

```ts
import { sql } from "drizzle-orm";
import { jogosZerados } from "@memory-card/db/schema";
import { db } from "../db.js";

export async function buscarDashboardPorPlataforma() {
  return db
    .select({
      console: jogosZerados.console,
      total: sql<number>`count(*)::int`
    })
    .from(jogosZerados)
    .groupBy(jogosZerados.console)
    .orderBy(sql`count(*) desc`);
}
```

## Progresso dinâmico de desafio QUANTIDADE

```ts
import { and, eq, gte, lte, sql } from "drizzle-orm";
import { desafios, jogosZerados } from "@memory-card/db/schema";
import { db } from "../db.js";

export async function calcularProgressoQuantidade(desafioId: number) {
  const [progresso] = await db
    .select({
      total: sql<number>`count(${jogosZerados.id})::int`
    })
    .from(desafios)
    .leftJoin(
      jogosZerados,
      and(
        gte(jogosZerados.finalizadoEm, desafios.iniciadoEm),
        lte(jogosZerados.finalizadoEm, desafios.finalizadoEm)
      )
    )
    .where(eq(desafios.id, desafioId));

  return progresso.total;
}
```

## Cruzamento automático ao registrar zeramento

```ts
import { and, eq, isNull } from "drizzle-orm";
import { desafioItens } from "@memory-card/db/schema";
import { db } from "../db.js";

export async function marcarItemDeDesafioConcluido(jogoId: number, igdbId: number | null) {
  if (igdbId === null) {
    return;
  }

  await db
    .update(desafioItens)
    .set({
      jogoZeradoId: jogoId,
      concluidoEm: new Date()
    })
    .where(and(eq(desafioItens.igdbId, igdbId), isNull(desafioItens.concluidoEm)));
}
```
