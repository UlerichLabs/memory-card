---
name: frontend-design-system
description: |
  Design system do Memory Card baseado nos tokens UlerichLabs.
  Usar quando criar ou editar qualquer componente visual, definir cores,
  tipografia, espaçamento, ou estilizar qualquer elemento de UI.
compatibility: Memory Card — apps/web (Tailwind CSS v3 + Shadcn/ui)
---

## Tokens de cor

Configure em `tailwind.config.ts` e `globals.css`. Nunca usar hex direto no JSX.

```css
:root {
  --background:     #0A0A0A;
  --surface:        #111111;
  --surface-raised: #181818;
  --border:         rgba(255, 255, 255, 0.07);
  --border-strong:  rgba(255, 255, 255, 0.12);
  --foreground:     #EFEFEF;
  --muted:          #666666;
  --accent:         #0EA5E9;
  --accent-dim:     rgba(14, 165, 233, 0.10);
  --gold:           #F59E0B;
  --gold-dim:       rgba(245, 158, 11, 0.10);
  --green:          #22C55E;
  --red:            #EF4444;
  --orange:         #F97316;
  --purple:         #A78BFA;
  --radius:         10px;
  --radius-sm:      6px;
}
```

## Tipografia

| Uso | Fonte | Classe Tailwind |
|-----|-------|-----------------|
| UI geral | Inter | `font-sans` |
| Dados, IDs, notas, tempo | JetBrains Mono | `font-mono` |
| Logo/títulos | A definir; usar Inter bold como placeholder | `font-sans font-bold` |

IDs sequenciais, notas e tempo sempre usam `font-mono`.

## Procedimento ao estilizar UI

1. Verifique se o token já existe no Tailwind.
2. Se não existir, adicione token semântico antes de usar.
3. Use `bg-background` na raiz da página.
4. Use `bg-surface` para cards e modais.
5. Use `bg-surface-raised` para inputs, hover de linhas e blocos elevados.
6. Aplique `font-mono` em ID, nota, tempo e estatísticas numéricas.
7. Valide nota 1 e nota 11 como casos especiais.

## GameCard

```tsx
export function GameCard() {
  return (
    <div className="rounded-xl border border-border bg-surface p-4 hover:bg-surface-raised">
      <span className="font-mono text-xs text-muted">#042</span>
      <h3 className="mt-1 font-medium text-foreground">Castlevania: SOTN</h3>
      <div className="mt-2 flex gap-2">
        <NotaBadge nota={10} />
        <DifficultyBadge dificuldade="AA" />
      </div>
    </div>
  );
}
```

## NotaBadge

```tsx
const notaConfig = {
  11: { label: "⭐ Jogo da Vida", className: "border-gold/30 bg-gold/10 text-gold" },
  1: { label: "💀 Tragédia", className: "border-red/30 bg-red/10 text-red" }
};
```

Para notas 2 a 10, usar escala visual entre `muted` e `accent`. Não tratar nota 11 como apenas "mais uma nota alta".

## DifficultyBadge

```tsx
const diffConfig = {
  C:   { label: 'Fácil',         color: '#888888' },
  B:   { label: 'Normal',        color: '#0EA5E9' },
  A:   { label: 'Difícil',       color: '#F59E0B' },
  AA:  { label: 'Muito Difícil', color: '#F97316' },
  AAA: { label: 'Platina',       color: '#EF4444' },
} as const;
```

Sempre exibir letra + label: `AA · Muito Difícil`.

## StatsHero

```tsx
export function StatsHero() {
  return (
    <div className="rounded-xl bg-surface-raised p-4">
      <p className="text-xs uppercase tracking-wide text-muted">Total Zerados</p>
      <p className="mt-1 font-mono text-3xl text-foreground">247</p>
    </div>
  );
}
```

## Gotchas

- Fundo da página sempre usa `bg-background`; nunca `bg-white` ou `bg-gray-*`.
- Cards usam `bg-surface`; nunca fundo claro.
- Nota 11 é `gold` + estrela.
- Nota 1 é `red` + caveira.
- IDs sequenciais sempre têm prefixo `#`: `#001`, `#042`, `#247`.
- Tempo jogado sempre em `font-mono`: `12:34:56`.
- `DifficultyBadge` sempre mostra letra + label.
- Não bloquear implementação por fonte de logo indefinida; usar Inter bold.

## Checklist antes de entregar

- [ ] Zero hex hardcoded no JSX.
- [ ] Fundo da página em `var(--background)` ou `bg-background`.
- [ ] IDs, notas e tempo em `font-mono`.
- [ ] Nota 11 com `gold` + estrela.
- [ ] Nota 1 com `red` + caveira.
- [ ] `DifficultyBadge` com letra + label.

## Navegação

- Tipo: barra horizontal no topo. Nunca sidebar.
- Fundo: var(--background) com borda inferior var(--border).
- Itens: Dashboard · Biblioteca · Desafios · Missões, com ícone Lucide à esquerda.
- Item ativo: fundo var(--accent-dim), texto var(--accent).
- Botão "Registrar" fixo no lado direito com fundo var(--accent).
- Avatar do usuário à direita do botão.

## Escala de notas 1-11

| Nota | Cor        | Label           |
|------|-----------|-----------------|
| 11   | #F59E0B   | ⭐ Jogo da Vida  |
| 10   | #22C55E   | Obra-prima      |
| 9    | #0EA5E9   | Excepcional     |
| 8    | #0EA5E9   | Ótimo           |
| 7    | #EFEFEF   | Bom             |
| 6    | #EFEFEF   | Decente         |
| 5    | #666666   | Médio           |
| 4    | #666666   | Abaixo da média |
| 3    | #666666   | Ruim            |
| 2    | #EF4444   | Terrível        |
| 1    | #EF4444   | 💀 Tragédia     |
