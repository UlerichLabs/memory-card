# Memory Card — frontend

Esqueleto com Vite, React, TypeScript, Tailwind CSS e Shadcn/ui.
A rota `/` exibe o status de `GET /api/v1/health`; não há telas funcionais.

## Docker (cenário principal)

Na raiz do repositório, configure `.env` a partir de `.env.example` e execute:

```sh
docker compose up --build -d
curl -sf http://localhost:18080/api/v1/health
curl -sf http://localhost:5173/
docker compose down
```

Abra http://localhost:5173 para verificar `Status da API: ok` no navegador.
O backend usa a porta 18080 no host porque a 8080 está ocupada pelo FlowLoop.
O PostgreSQL continua na porta 5432. `down` preserva o volume do banco; não use
`down -v` para encerrar o ambiente com dados que deseja manter.

## Desenvolvimento com hot reload

```sh
# Na raiz:
docker compose up -d postgres api
# Em apps/web (Node.js 24 LTS):
npm ci
cp .env.example .env.local
npm run dev
```

Pare o serviço `web` antes de iniciar o Vite: ambos usam 5173.

## Configuração da API

O cliente usa `fetch`, nativo do navegador, sem dependência HTTP adicional.
`VITE_API_URL` é uma URL acessível **pelo navegador**, não o hostname Docker `api`.
O padrão é `http://localhost:18080`; não inclua `/api/v1` na URL base.

No Vite local, configure `apps/web/.env.local`. No Docker, defina `VITE_API_URL`
no `.env` da raiz ou no ambiente do shell e reconstrua o serviço:

```sh
VITE_API_URL=https://api.example.com docker compose up --build -d web
```

O Dockerfile recebe a variável como argumento de build. Alterar apenas o ambiente
do container Nginx não modifica o JavaScript compilado. Variáveis `VITE_*` são
públicas e não devem conter segredos. O health público permite CORS sem credenciais;
essa configuração não se estende às futuras rotas de autenticação.

A imagem final usa Nginx Alpine, assets com cache e fallback SPA para `index.html`.
Não existe Compose de produção neste repositório; a imagem já serve build estático.

## Estrutura e comandos

- `src/components/ui`: componentes Shadcn (`button` e `input`).
- `src/pages`: página placeholder.
- `src/routes`: rotas do React Router.
- `src/services/api.ts`: cliente HTTP e consulta de health.
- `src/lib/utils.ts`: utilitário de classes do Shadcn.

```sh
npm run build
npm run lint
npx shadcn@latest add <componente>
```
