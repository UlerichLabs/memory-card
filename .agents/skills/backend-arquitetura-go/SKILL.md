---
name: backend-arquitetura-go
description: Padroes obrigatorios de arquitetura para o backend Go do Memory Card (apps/api). Usar sempre que criar pastas, handlers, services, repositories ou queries.
---

# Memory Card - Arquitetura de Backend (Go)

Stack: Gin + sqlc + PostgreSQL + golang-migrate. Monorepo com o frontend em React (`apps/web`).

## Hard Rules - nunca violar

- Zero ORM (GORM, ent, etc.) - sqlc e SQL puro, decisao de produto (idiomatico, sem reflection magica)
- Zero query SQL escrita direto no handler ou no service - toda query vive em arquivo `.sql` e passa por `sqlc generate`
- Zero logica de negocio no handler - handler so faz parse de request, chama o service, formata a resposta
- Zero acesso direto ao banco fora da camada repository/sqlc gerado - service nunca importa `database/sql` diretamente
- Zero funcao publica sem `context.Context` como primeiro parametro quando envolve I/O (banco, HTTP externo)
- Zero pacote de dominio exportado em `internal/` sendo importado de fora do modulo - e assim que o Go garante encapsulamento, nao violar
- Zero variavel de ambiente lida fora do pacote de config na inicializacao - nunca `os.Getenv` espalhado pelo codigo

## Camadas (obrigatorio seguir esta ordem de dependencia)

```
Handler (Gin)  ->  Service (regra de negocio)  ->  Repository (sqlc gerado)  ->  Postgres
```

- **Handler**: recebe `*gin.Context`, faz bind/validacao de request, chama exatamente um metodo de service, mapeia erro para status HTTP, escreve resposta JSON. Nunca contem `if` de regra de negocio.
- **Service**: recebe interfaces de repository (nunca o struct concreto do sqlc direto, para permitir mock em teste), implementa a regra de negocio, retorna erros de dominio tipados (nao `errors.New` solto).
- **Repository**: e o codigo gerado pelo sqlc a partir de `.sql` + uma interface fina que o service consome, para permitir mock.

## Estrutura de pastas (obrigatoria)

```
apps/api/
  cmd/
    api/
      main.go              # so wiring: le config, monta dependencias, sobe o servidor
  internal/
    config/                 # leitura de env vars, unico lugar que faz isso
    handler/                # um arquivo por recurso (auth_handler.go, jogos_handler.go)
    service/                # um arquivo por recurso (auth_service.go)
    repository/             # interfaces + wiring do sqlc gerado
    db/
      migrations/            # golang-migrate, up/down numeradas
      queries/                # arquivos .sql fonte do sqlc
      sqlc/                   # codigo gerado (nao editar a mao)
    middleware/              # auth, cors, logging
  sqlc.yaml
```

## Antes de qualquer tarefa

1. Leia os arquivos que serao editados - nunca assuma o schema ou a query existente
2. Se a tarefa precisa de uma query nova, escreva o `.sql` em `db/queries/` primeiro e rode `sqlc generate` antes de escrever o service
3. Se a tarefa precisa de uma tabela nova ou coluna nova, crie a migration (up + down) antes de qualquer query
4. Verifique se ja existe uma interface de repository para o dominio - se nao existir, crie antes do service
5. Somente entao implemente handler e service

## Injecao de dependencia

Sem framework de DI - wiring manual e explicito em `cmd/api/main.go`. Construtores recebem dependencias por parametro (`NewAuthService(repo AuthRepository) *AuthService`), nunca instanciam suas proprias dependencias internamente - isso e o que permite mockar em teste.
