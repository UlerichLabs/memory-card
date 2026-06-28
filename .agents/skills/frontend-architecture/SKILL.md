---
name: frontend-architecture
description: |
  Padrões de arquitetura do frontend do Memory Card (React + Vite + TypeScript).
  Usar quando criar rotas, pages, hooks, services, stores Zustand, configurar
  React.lazy, ou tomar qualquer decisão estrutural que afete mais de um componente.
compatibility: Memory Card — apps/web
---

## Estrutura de pastas

```text
apps/web/src/
├── App.tsx                    # APENAS declaração de rotas — zero lógica
├── lib/
│   ├── api.ts                 # fetch wrapper base (usa VITE_API_URL)
│   ├── formatters.ts          # formatTime(segundos → HH:MM:SS), formatDate
│   └── utils.ts
├── stores/
│   └── filtros.store.ts       # Zustand — estado dos filtros ativos
├── hooks/                     # hooks reutilizáveis globais
├── components/
│   ├── ui/                    # Shadcn components — não editar diretamente
│   ├── jogos/                 # GameCard, GameForm, GameDetail
│   ├── dashboard/             # StatsHero, Charts, Rankings
│   ├── desafios/              # ChallengeCard, ChallengeForm, Progress
│   └── layout/                # Sidebar, BottomNav, PageHeader
└── pages/
    ├── biblioteca.tsx          # lista principal + filtros
    ├── dashboard.tsx           # estatísticas e gráficos
    ├── desafios.tsx            # campanhas e desafios ativos
    ├── explorar.tsx            # explorador IGDB (Fase 2)
    └── perfil.tsx              # perfil público (Fase 2)
```

## Antes de qualquer tarefa

1. Leia `apps/web/src/App.tsx` para entender as rotas existentes.
2. Verifique se já existe hook ou service para o domínio.
3. Se não existir service, crie-o antes do hook.
4. Se não existir hook, crie-o antes da page ou componente consumidor.
5. Estime linhas; se ultrapassar o limite, decomponha antes de escrever.

## Como criar uma nova page

1. Criar o arquivo em `apps/web/src/pages/`.
2. Registrar a page em `App.tsx` com `React.lazy`.
3. Criar hook em `hooks/` se a page precisar de estado + chamada de API.
4. Criar service em `lib/` se o hook precisar de chamadas HTTP.
5. Decompor em `components/<dominio>/` quando o JSX passar de 40 linhas.

## Como criar um novo componente

1. Verificar se já existe componente similar em `components/ui/`.
2. Se for novo domínio, criar pasta `components/<dominio>/`.
3. Extrair para hook quando lógica com `useState` + `useEffect` passar de 20 linhas.
4. Extrair subcomponente quando uma seção visual passar de 40 linhas de JSX.
5. Importar tipos de domínio de `@memory-card/types`.

## Padrão de rota

```tsx
import { lazy, Suspense } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { PageLoader } from "@/components/layout/PageLoader";
import { GameDetail } from "@/components/jogos/GameDetail";
import { GameForm } from "@/components/jogos/GameForm";

const Biblioteca = lazy(() => import("@/pages/biblioteca"));
const Dashboard = lazy(() => import("@/pages/dashboard"));
const Desafios = lazy(() => import("@/pages/desafios"));

export default function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/" element={<Navigate to="/biblioteca" replace />} />
          <Route path="/biblioteca" element={<Biblioteca />} />
          <Route path="/biblioteca/:id" element={<GameDetail />} />
          <Route path="/biblioteca/novo" element={<GameForm />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/desafios" element={<Desafios />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
```

## Padrão de service

```ts
import { api } from "@/lib/api";
import type { FiltrosJogo, JogoZerado, JogoZeradoCreateInput } from "@memory-card/types";

export async function listJogos(filtros?: FiltrosJogo) {
  return api.get<JogoZerado[]>("/jogos", { params: filtros });
}

export async function createJogo(input: JogoZeradoCreateInput) {
  return api.post<JogoZerado>("/jogos", input);
}
```

## Padrão de hook

```ts
import { useQuery } from "@tanstack/react-query";
import type { FiltrosJogo } from "@memory-card/types";
import { listJogos } from "@/lib/jogos.service";

export function useJogos(filtros?: FiltrosJogo) {
  return useQuery({
    queryKey: ["jogos", filtros],
    queryFn: () => listJogos(filtros)
  });
}
```

## Limites de linhas

| Tipo | Limite |
|------|--------|
| `components/*.tsx` | 150 linhas |
| `pages/*.tsx` | 200 linhas |
| `hooks/use*.ts` | 80 linhas |
| `lib/*service.ts` | 100 linhas |
| `lib/utils.ts` | 60 linhas |

## Gotchas

- `VITE_API_URL` aponta para `http://localhost:3001` em dev; nunca hardcodar porta.
- TanStack Query é o cache; não usar `useState` para dados do servidor.
- Zustand é só para estado de UI local, como filtros e modais.
- Nunca guardar dados da API em Zustand.
- Shadcn em `components/ui/` não é editado diretamente; criar wrapper se precisar customizar.
- `formatTime()` de `@/lib/formatters` é obrigatório para exibir tempo.
- Nunca converter segundos inline dentro de componente.

## Checklist antes de entregar

- [ ] `App.tsx` sem lógica de negócio.
- [ ] Toda nova page com `React.lazy`.
- [ ] Sem chamada `fetch` direta em `.tsx`.
- [ ] Toda chamada de API passa por service + hook.
- [ ] Sem `useState` para dados do servidor.
- [ ] Nenhum arquivo ultrapassa o limite de linhas.
- [ ] Tipos importados de `@memory-card/types`.
