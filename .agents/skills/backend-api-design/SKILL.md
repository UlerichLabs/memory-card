---
name: backend-api-design
description: Convencoes obrigatorias de design de API REST para o backend do Memory Card. Usar sempre que criar ou revisar um endpoint novo.
---

# Memory Card - API Design

Convencoes de como a API Go se comunica com o frontend React. Objetivo: previsibilidade - quem consome a API nunca deveria precisar adivinhar o formato de nada.

## Hard Rules - nunca violar

- Zero rota fora do prefixo `/api/v1/` - versionamento desde o primeiro endpoint
- Zero nome de recurso em ingles quando o dominio do produto usa portugues (`/api/v1/jogos`, nunca `/api/v1/games`) - dominio em portugues, ver skill de nomenclatura do frontend para consistencia
- Zero verbo na URL (`/api/v1/criar-jogo`) - o metodo HTTP ja e o verbo (`POST /api/v1/jogos`)
- Zero resposta de sucesso e erro com formatos de JSON diferentes entre endpoints - schema de erro e unico e consistente em toda a API
- Zero status HTTP generico (200 para tudo, 500 para tudo) quando existe um status mais especifico e correto
- Zero formato de `codigo` de erro divergente do padrao definido abaixo - e a MESMA string usada como chave de traducao pela skill `backend-i18n`, nunca dois formatos diferentes

## Formato de resposta - sucesso

```json
{
  "data": { }
}
```

Listas paginadas:

```json
{
  "data": [ ],
  "meta": { "total": 247, "pagina": 1, "por_pagina": 20 }
}
```

## Formato de resposta - erro (obrigatorio, mesmo schema em toda a API)

```json
{
  "error": {
    "codigo": "auth.register.email_taken",
    "mensagem": "Este email ja esta em uso."
  }
}
```

`codigo` segue SEMPRE o formato dotted lowercase `<dominio>.<contexto>.<caso>` (ex: `auth.register.email_taken`, `auth.login.invalid_credentials`) - e uma chave estavel que o frontend pode tratar programaticamente, e e a MESMA chave que a skill `backend-i18n` usa para resolver a traducao. Nunca usar outro formato - um unico padrao em toda a API. `mensagem` e sempre o texto ja traduzido no idioma do usuario, nunca a chave crua.

## Status HTTP - referencia obrigatoria

| Situacao | Status |
|---|---|
| Sucesso com corpo | 200 |
| Recurso criado | 201 |
| Sucesso sem corpo | 204 |
| Erro de validacao / input invalido | 400 |
| Nao autenticado / token invalido | 401 |
| Autenticado mas sem permissao | 403 |
| Recurso nao encontrado | 404 |
| Conflito (ex: email duplicado) | 409 |
| Token/link expirado | 410 |
| Rate limit atingido | 429 |
| Erro interno / dependencia externa fora do ar | 500 / 502 |

## Paginacao

Query params `pagina` (default 1) e `por_pagina` (default 20, maximo 100) em toda lista que pode crescer sem limite (biblioteca de jogos, historico de atividade).

## Antes de criar um endpoint novo - checklist

1. A rota segue `/api/v1/<recurso-em-portugues>`?
2. O metodo HTTP e o verbo certo (GET/POST/PUT/PATCH/DELETE), sem verbo na URL?
3. A resposta de sucesso segue o envelope `{ "data": ... }`?
4. Os erros possiveis desse endpoint estao mapeados para status HTTP especificos (ver tabela)?
5. O campo `codigo` do erro segue o formato dotted lowercase `<dominio>.<contexto>.<caso>`, identico a chave usada em `backend-i18n`?
6. Se e uma lista que pode crescer, tem paginacao?
