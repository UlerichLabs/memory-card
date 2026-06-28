---
name: typescript-best-practices
description: |
  Boas práticas TypeScript do Memory Card com Node.js 22, TypeScript 5 e ESM.
  Usar ao criar tipos compartilhados, schemas Zod, services, hooks, validações,
  ou ao revisar qualquer código TypeScript do monorepo.
compatibility: Memory Card — apps/web, apps/api, packages/types, packages/db
---

## Antes de qualquer tarefa

1. Verifique se o tipo já existe em `packages/types/src/`.
2. Verifique se enum já existe em `packages/types/src/enums.ts`.
3. Se a entrada vem de usuário ou HTTP, modele com Zod.
4. Se o dado vem do banco, confira o schema em `@memory-card/db`.
5. Organize imports antes de entregar.

## Hard rules

- Zero `any`; usar `unknown` com type guard ou tipo explícito.
- Types de domínio ficam em `packages/types/src/`.
- Nunca redefinir tipo de domínio inline.
- Enums de consoles, gêneros e notas sempre vêm de `packages/types/src/enums.ts`.
- Usar `async/await`; zero `.then().catch()` encadeado.
- Zod para validação de entrada em todo endpoint.
- ESM modules com `import` e `export`; zero `require()`.
- Imports organizados: node built-ins, libs externas, `@packages` internos, relativos.
- Schema de banco vem de `@memory-card/db`.

## Tipos obrigatórios em packages/types

- `JogoZerado`
- `JogoZeradoCreateInput`
- `JogoZeradoUpdateInput`
- `Campanha`
- `Desafio`
- `DesafioItem`
- `Conquista`
- `DashboardTotais`
- `DashboardPorAno`
- `DashboardPorPlataforma`

Conteúdo longo de enums e tipos completos deve ficar em arquivo de referência da skill ou direto em `packages/types`, não duplicado em cada feature.

## Padrão de enum compartilhado

```ts
export const dificuldades = ["C", "B", "A", "AA", "AAA"] as const;
export type Dificuldade = (typeof dificuldades)[number];

export const notas = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11] as const;
export type Nota = (typeof notas)[number];
```

## Padrão de tipo de domínio

```ts
import type { Dificuldade, Nota } from "./enums";

export type JogoZerado = {
  id: number;
  igdbId: number | null;
  nome: string;
  console: string;
  genero: string | null;
  tipo: string | null;
  iniciadoEm: Date | null;
  finalizadoEm: Date;
  tempoSegundos: number;
  nota: Nota;
  dificuldade: Dificuldade;
  condicaoZeramento: string | null;
  destaque: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type JogoZeradoCreateInput = Omit<JogoZerado, "id" | "createdAt" | "updatedAt">;
export type JogoZeradoUpdateInput = Partial<JogoZeradoCreateInput>;
```

## Type guard com unknown

```ts
type ApiError = {
  error: {
    message: string;
    code?: string;
  };
};

export function isApiError(value: unknown): value is ApiError {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  if (!("error" in value)) {
    return false;
  }

  const error = value.error;

  return typeof error === "object" && error !== null && "message" in error;
}
```

## Zod em endpoint

```ts
import { z } from "zod";
import { dificuldades } from "@memory-card/types";

export const criarJogoZeradoSchema = z.object({
  nome: z.string().min(1).max(200),
  console: z.string().min(1).max(50),
  genero: z.string().max(50).nullable(),
  tipo: z.string().max(50).nullable(),
  iniciadoEm: z.coerce.date().nullable(),
  finalizadoEm: z.coerce.date(),
  tempoSegundos: z.number().int().positive(),
  nota: z.number().int().min(1).max(11),
  dificuldade: z.enum(dificuldades),
  condicaoZeramento: z.string().max(500).nullable(),
  destaque: z.boolean()
});

export type CriarJogoZeradoInput = z.infer<typeof criarJogoZeradoSchema>;
```

## Ordem de imports

```ts
import { readFile } from "node:fs/promises";

import { z } from "zod";

import { jogosZerados } from "@memory-card/db/schema";
import type { JogoZeradoCreateInput } from "@memory-card/types";

import { db } from "../db.js";
```

## Gotchas

- `unknown` só ajuda se houver narrowing antes do uso.
- `as` não substitui validação de runtime.
- `z.infer` evita divergência entre schema e tipo.
- Em Node ESM, imports relativos compilados precisam de extensão `.js`.
- Não usar enum TypeScript quando `as const` resolve melhor para compartilhamento frontend/backend.

## Checklist antes de entregar

- [ ] Nenhum `any`.
- [ ] Tipos de domínio vieram de `packages/types`.
- [ ] Enums vieram de `packages/types/src/enums.ts`.
- [ ] Código usa ESM.
- [ ] Código async usa `async/await`.
- [ ] Entradas de endpoint usam Zod.
- [ ] Imports estão na ordem definida.
