---
name: backend-database
description: |
  Padrões de banco do Memory Card com Drizzle ORM, PostgreSQL e docker-compose.
  Usar ao alterar schema, criar migrations, escrever queries críticas,
  ou persistir regras de jogos, campanhas, desafios e conquistas.
compatibility: Memory Card — packages/db + apps/api
---

## Antes de qualquer tarefa

1. Leia `packages/db/src/schema.ts`.
2. Leia a migration mais recente antes de gerar outra.
3. Confirme se o tipo de domínio já existe em `packages/types`.
4. Para query crítica, procure referência em `references/queries.md`.
5. Rode geração/migração somente depois de revisar o SQL gerado.

## Hard rules

- Tempo sempre em segundos no banco, usando `tempo_segundos integer`.
- Nota sempre integer 1-11; validar no Zod antes de persistir.
- Dificuldade como enum PostgreSQL `C/B/A/AA/AAA`.
- Zero `deleteById`; usar soft delete com `deleted_at` quando aplicável.
- Progresso de desafio `QUANTIDADE` sempre calculado via `COUNT`.
- Nunca salvar snapshot de progresso `QUANTIDADE`.
- Cruzamento de zeramento com `desafio_itens` via update automático ao registrar jogo.
- Toda migration fica em `packages/db/src/migrations/` com nome sequencial.
- Schema compartilhado fica em `packages/db/src/schema.ts`.
- Tipos e enums TypeScript vêm de `packages/types`.

## Schema principal esperado

O schema compartilhado deve evoluir para conter:

- `jogos_zerados`
- `campanhas`
- `desafios`, com `tipoDesafioEnum`
- `desafio_itens`
- `conquistas`
- `favoritos`

Conteúdo longo de schema completo deve ficar em `references/`, não no `SKILL.md`.

## Como alterar schema

1. Alterar `packages/db/src/schema.ts`.
2. Garantir que enums TypeScript relacionados existam em `packages/types`.
3. Gerar migration.
4. Revisar SQL gerado, especialmente constraints, defaults e enum diffs.
5. Aplicar migration em banco local.
6. Atualizar services que dependem da mudança.

## Exemplo de constraint

```ts
import { sql } from "drizzle-orm";
import { check, integer, pgEnum, pgTable, serial } from "drizzle-orm/pg-core";
import { dificuldades } from "@memory-card/types";

export const dificuldadeEnum = pgEnum("dificuldade", dificuldades);

export const jogosZerados = pgTable(
  "jogos_zerados",
  {
    id: serial("id").primaryKey(),
    nota: integer("nota").notNull(),
    dificuldade: dificuldadeEnum("dificuldade").notNull()
  },
  () => [check("jogos_zerados_nota_range", sql`nota between 1 and 11`)]
);
```

## Como escrever query crítica

1. Escreva a query no service, não na route.
2. Use schema importado de `@memory-card/db/schema`.
3. Tipar retorno com tipos de `@memory-card/types`.
4. Para agregações, converter retorno SQL para número (`::int`, `::float`).
5. Documentar variações novas em `references/queries.md`.

## Dashboard totals

```ts
import { sql } from "drizzle-orm";
import { jogosZerados } from "@memory-card/db/schema";

export const dashboardTotalsQuery = {
  totalJogos: sql<number>`count(*)::int`,
  tempoTotalSegundos: sql<number>`coalesce(sum(${jogosZerados.tempoSegundos}), 0)::int`,
  mediaNota: sql<number>`coalesce(avg(${jogosZerados.nota}), 0)::float`
};
```

## Gotchas

- `updated_at` não muda sozinho com `defaultNow()`; service deve setar em updates ou usar trigger.
- `timestamp without time zone` exige consistência no backend; não misturar timezone sem decisão.
- Drizzle pode gerar SQL inválido se constraint qualificar coluna dentro de `CHECK`; revisar SQL.
- `COUNT` no PostgreSQL pode vir como string dependendo do driver; usar cast explícito.
- Soft delete exige filtros em todas as queries de leitura do domínio.

## Checklist antes de entregar

- [ ] Schema alterado em `packages/db/src/schema.ts`.
- [ ] Migration sequencial criada no local correto.
- [ ] SQL gerado revisado.
- [ ] Nota validada 1-11 no Zod e no banco quando aplicável.
- [ ] Tempo salvo em segundos.
- [ ] Soft delete usado quando aplicável.
- [ ] Progresso `QUANTIDADE` usa `COUNT`.
- [ ] Queries importam schema de `@memory-card/db`.
