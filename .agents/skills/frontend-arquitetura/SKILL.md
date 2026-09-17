---
name: frontend-arquitetura
description: Padroes obrigatorios de arquitetura para o frontend do Memory Card (apps/web). Usar sempre que criar pastas, componentes, services, rotas ou estado.
---

# Memory Card - Arquitetura de Frontend

Stack: Vite + React + TypeScript + Tailwind + Shadcn/ui + react-router-dom. Monorepo com o backend em Go (`apps/api`).

## Hard Rules - nunca violar

- Zero `fetch`/`axios` direto dentro de um componente ou pagina - sempre via funcao em `services/`
- Zero URL hardcoded - sempre `import.meta.env.VITE_API_URL`
- Zero chamada direta a IGDB do frontend - toda busca de jogo passa pelo proxy do backend Go, nunca credenciais Twitch/Amazon no cliente
- Zero `any` sem comentario justificando o motivo
- Zero state manager (Redux, Zustand, Jotai, etc.) sem justificativa documentada - comecar com `useState`/`useContext`, so introduzir lib quando a complexidade genuinamente exigir
- Zero logica de refresh de token duplicada em varios componentes - centralizada em um unico lugar (interceptor do cliente HTTP ou wrapper de `services/`)
- Zero rota privada sem passar por `PrivateRoute`

## Limites de linhas - bloqueio automatico

| Tipo | Limite |
|---|---|
| Componente `*.tsx` | 150 linhas |
| Pagina/rota `*Page.tsx` | 200 linhas |
| Hook `use*.ts` | 80 linhas |
| Service `*Service.ts` | 100 linhas |
| Utilitario `*.ts` | 60 linhas |

Se a implementacao ultrapassar o limite, propor decomposicao antes de escrever qualquer codigo.

## Estrutura de pastas (obrigatoria)

```
apps/web/src/
  components/       # componentes reutilizaveis entre telas
    ui/             # componentes do shadcn (gerados, nao editar a mao sem necessidade)
  pages/            # uma pasta ou arquivo por tela/rota
  services/         # cliente HTTP e chamadas a API, uma funcao tipada por endpoint
  routes/           # definicao de rotas e PrivateRoute
  hooks/            # hooks customizados, so quando a necessidade for real
  types/            # tipos TypeScript espelhando os DTOs da API Go
  App.tsx
  main.tsx
```

## Antes de qualquer tarefa

1. Leia os arquivos que serao editados - nunca assuma o conteudo
2. Estime as linhas do resultado - se ultrapassar o limite, proponha decomposicao primeiro
3. Verifique se existe um `*Service.ts` para o dominio da chamada - se nao existir, crie o service antes de implementar a tela/componente
4. Verifique se o tipo do dado ja existe em `types/` espelhando o DTO do backend - se nao existir, crie antes de consumir a API
5. Somente entao implemente

## Quando extrair para hook

- Logica com `useState` + `useEffect` que ultrapassa ~20 linhas dentro de um componente
- Logica reutilizada em mais de um componente (na 2a repeticao, extrair)

## Rotas (referencia inicial, expandir conforme telas forem desenhadas)

| Rota | Publica/Privada | Descricao |
|---|---|---|
| `/login` | Publica | Tela de login |
| `/cadastro` | Publica | Tela de cadastro |
| `/esqueci-senha` | Publica | Fluxo de recuperacao |
| `/` | Privada | Dashboard principal |
| `/biblioteca` | Privada | Biblioteca de jogos zerados |

## Docker

Frontend roda containerizado junto do backend via `docker-compose.yml` (build multi-stage: Node para build, Nginx alpine para servir estatico), porta 5173 em dev. Ver historia MEMOR-49 no Huly.
