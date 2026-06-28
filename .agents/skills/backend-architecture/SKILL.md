---
name: backend-architecture
description: |
  Padrões de arquitetura da API do Memory Card (Fastify + TypeScript + Drizzle).
  Usar quando criar rotas, services, plugins, validações Zod, health checks,
  ou qualquer fluxo HTTP em apps/api.
compatibility: Memory Card — apps/api
---

## Estrutura alvo

```text
apps/api/src/
├── routes/
│   ├── jogos.routes.ts
│   ├── dashboard.routes.ts
│   ├── desafios.routes.ts
│   ├── campanhas.routes.ts
│   └── igdb.routes.ts
├── services/
│   ├── jogos.service.ts
│   ├── dashboard.service.ts
│   ├── desafios.service.ts
│   └── igdb.service.ts
├── plugins/
│   ├── cors.ts
│   └── swagger.ts
└── server.ts
```

## Antes de qualquer tarefa

1. Leia `server.ts` para entender plugins e rotas já registradas.
2. Leia o service do domínio antes de alterar uma route.
3. Leia o schema em `packages/db/src/schema.ts` antes de persistir dados.
4. Verifique tipos em `@memory-card/types`.
5. Se o endpoint recebe `body`, crie schema Zod antes da implementação.

## Hard rules

- Routes apenas declaram endpoints e validação; zero lógica de negócio.
- Services têm toda a lógica; sem acesso direto a `request` ou `reply`.
- Todo endpoint com body usa Zod.
- Tempo jogado sempre armazenado em segundos.
- MVP sem autenticação; `usuario_id` fica reservado para Fase 2.
- Todas as rotas usam prefixo `/api/v1`.
- Zero `console.log` em produção; usar Fastify logger.
- Sucesso retorna `{ data: T }`.
- Erro retorna `{ error: { message: string, code?: string } }`.
- Routes até 80 linhas, services até 150, plugins até 60.

## Como criar rota

1. Criar schema Zod no arquivo da route ou em arquivo próximo se for reutilizado.
2. Registrar endpoint com prefixo `/api/v1`.
3. Fazer parse de `request.body`, `request.params` ou `request.query`.
4. Chamar service com dados já validados.
5. Retornar `{ data }`.
6. Logar contexto útil com `request.log`, sem payload sensível.

## Route

```ts
import type { FastifyInstance } from "fastify";
import { z } from "zod";
import { dificuldades } from "@memory-card/types";
import { criarJogoZerado, listarJogosZerados } from "../services/jogos.service.js";

const criarJogoSchema = z.object({
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

export async function jogosRoutes(app: FastifyInstance) {
  app.get("/api/v1/jogos", async () => {
    const jogos = await listarJogosZerados();
    return { data: jogos };
  });

  app.post("/api/v1/jogos", async (request, reply) => {
    const input = criarJogoSchema.parse(request.body);
    const jogo = await criarJogoZerado(input);
    request.log.info({ jogoId: jogo.id }, "jogo zerado registrado");
    return reply.code(201).send({ data: jogo });
  });
}
```

## Service

```ts
import { jogosZerados } from "@memory-card/db/schema";
import type { JogoZeradoCreateInput } from "@memory-card/types";
import { db } from "../db.js";

export async function listarJogosZerados() {
  return db.select().from(jogosZerados);
}

export async function criarJogoZerado(input: JogoZeradoCreateInput) {
  const [jogo] = await db
    .insert(jogosZerados)
    .values(input)
    .returning();

  return jogo;
}
```

## Gotchas

- Não passar `request` para service.
- Não validar o mesmo body duas vezes; route valida, service assume input tipado.
- Não misturar nomes snake_case da API com camelCase do TypeScript sem decisão explícita.
- `tempoSegundos` é contrato interno; conversão `HH:MM:SS` fica no frontend.
- Health check pode consultar `select 1`, mas não deve depender de tabelas da aplicação.

## Checklist antes de entregar

- [ ] Rotas usam `/api/v1`.
- [ ] Route sem regra de negócio.
- [ ] Service sem `request` ou `reply`.
- [ ] Body validado com Zod.
- [ ] Respostas seguem `{ data }` e `{ error }`.
- [ ] Nenhum `console.log`.
- [ ] Tipos vêm de `@memory-card/types`.
- [ ] Schema vem de `@memory-card/db`.
