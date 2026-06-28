---
name: frontend-components
description: |
  Padrões de componentes React para o Memory Card.
  Usar quando criar ou editar qualquer arquivo .tsx: formulários, modais,
  cards, listas, sheets, ou qualquer elemento de UI com lógica.
compatibility: Memory Card — apps/web
---

## Antes de qualquer tarefa

1. Leia o arquivo que será editado; nunca assuma o conteúdo.
2. Verifique hooks, services e componentes vizinhos do mesmo domínio.
3. Estime linhas; se ultrapassar o limite, decomponha primeiro.
4. Somente então implemente.

## Hard rules

- Zero chamada `fetch` ou API direta em `.tsx`; sempre via hook com TanStack Query.
- Zero formatadores inline; sempre usar `@/lib/formatters`.
- Zero comentários no código.
- Zero hex hardcoded; usar tokens do design system.
- Tempo sempre em segundos internamente; exibir com `formatTime()`.
- Nota sempre integer 1-11; nunca float, nunca string.
- Dificuldade sempre enum `C/B/A/AA/AAA`; nunca string livre.
- Todo formulário reseta ao fechar modal ou sheet.
- Tipos de domínio vêm de `@memory-card/types`.

## Padrão de formulário

```tsx
import { useState } from "react";
import type { Dificuldade } from "@memory-card/types";

type GameFormFields = {
  nome: string;
  console: string;
  genero: string;
  tipo: string;
  iniciadoEm: string | null;
  finalizadoEm: string;
  horas: number;
  minutos: number;
  segundos: number;
  nota: number;
  dificuldade: Dificuldade;
  condicaoZeramento: string;
  destaque: boolean;
};

const initialFields: GameFormFields = {
  nome: "",
  console: "",
  genero: "",
  tipo: "",
  iniciadoEm: null,
  finalizadoEm: "",
  horas: 0,
  minutos: 0,
  segundos: 0,
  nota: 8,
  dificuldade: "B",
  condicaoZeramento: "",
  destaque: false
};

export function GameForm() {
  const [fields, setFields] = useState<GameFormFields>(initialFields);

  function handleSubmit() {
    const tempoSegundos = fields.horas * 3600 + fields.minutos * 60 + fields.segundos;
    createJogo({ ...fields, tempoSegundos });
  }
}
```

O campo `tipo` deve ser um select dependente do `genero` selecionado. Importar gêneros e tipos de `@memory-card/types` para popular selects.

## Padrão de campo com erro

```tsx
<div className="flex flex-col gap-1.5">
  <Label className="text-xs uppercase tracking-wide text-muted">Nome do jogo</Label>
  <Input
    value={fields.nome}
    onChange={(event) => setFields((current) => ({ ...current, nome: event.target.value }))}
    aria-invalid={Boolean(errors.nome)}
    className={errors.nome ? "border-red focus-visible:ring-red" : ""}
  />
  {errors.nome && <p className="text-xs text-red">{errors.nome}</p>}
</div>
```

## Padrão de modal ou sheet com reset

```tsx
import { useEffect } from "react";

type Props = {
  open: boolean;
  onClose: () => void;
};

export function GameFormSheet({ open, onClose }: Props) {
  useEffect(() => {
    if (!open) {
      setFields(initialFields);
      setErrors({});
    }
  }, [open]);
}
```

## Padrão de lista com estados

```tsx
{isLoading ? (
  <LoadingSpinner />
) : jogos.length === 0 ? (
  <EmptyState message="Nenhum jogo registrado ainda." />
) : (
  jogos.map((jogo) => <GameCard key={jogo.id} jogo={jogo} />)
)}
```

## Decomposição de referência — Biblioteca

1. `pages/biblioteca.tsx` orquestra filtros, query e layout.
2. `components/jogos/GameList.tsx` renderiza loading, empty e lista.
3. `components/jogos/GameCard.tsx` renderiza um item.
4. `components/jogos/GameFilters.tsx` renderiza filtros e atualiza Zustand.
5. `hooks/useJogos.ts` busca dados.
6. `lib/jogos.service.ts` chama a API.

## Gotchas

- `React Hook Form` e Zod devem ser preferidos quando o formulário crescer.
- Não duplicar parsing de tempo em componente; mover para `@/lib/formatters`.
- Não criar `useEffect` para copiar dados da API para estado local.
- Em modal, resetar no fechamento, não no submit.
- `destaque` é regra de negócio: apenas 1 por ano; a API deve validar também.

## Checklist antes de entregar

- [ ] Nenhum `.tsx` chama API diretamente.
- [ ] Nenhum formatador inline foi criado.
- [ ] Nenhum comentário foi adicionado ao código.
- [ ] Sem hex hardcoded.
- [ ] Tempo armazenado em segundos.
- [ ] Tempo exibido via `formatTime()`.
- [ ] Nota é integer 1-11.
- [ ] Dificuldade vem de enum.
- [ ] Formulários resetam ao fechar.
