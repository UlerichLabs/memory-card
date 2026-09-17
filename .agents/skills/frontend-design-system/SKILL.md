---
name: frontend-design-system
description: Padroes obrigatorios de design visual do Memory Card. Usar sempre que criar ou editar qualquer componente, tela ou elemento de UI.
---

# Memory Card - Design System

## Hard Rules - nunca violar

- Zero cores hardcoded fora dos tokens abaixo - sempre via variaveis CSS/Tailwind config, nunca hex solto no meio de um componente
- Zero roxo como cor de acento - proibido por decisao de produto (evitar cara de app generico feito com IA)
- Zero emoji como icone - sempre SVG stroke inline (estilo lucide)
- Zero gradiente decorativo sem funcao - gradiente so em placeholder de capa de jogo (ainda sem arte real) ou em glow pontual de destaque (ex: Jogo da Vida, Game do Ano)
- Zero fonte serifada ou fonte adicional - unica familia tipografica e Inter
- Zero mais de 2-3 elementos usando a cor de acento visiveis na mesma tela ao mesmo tempo, fora a navegacao
- Zero dourado (`highlight-gold`) fora do contexto "Game do Ano" ou nota/badge de destaque - nunca usar em UI generica

## Tokens de cor (obrigatorio usar estes valores)

| Token | Hex | Uso |
|---|---|---|
| `bg-primary` | `#15161A` | Fundo principal |
| `bg-surface` | `#1A1B20` | Cards, superficies elevadas |
| `bg-surface-alt` | `#1D1F25` | Inputs, campo de busca |
| `border` | `#24262C` | Bordas padrao, divisores |
| `border-subtle` | `#2A2C33` | Bordas de inputs |
| `text-primary` | `#EDEDED` | Texto principal |
| `text-secondary` | `#9A9CA5` | Texto secundario, labels |
| `text-muted` | `#6B6D76` | Timestamps, placeholders |
| `text-faint` | `#52545C` | Icones inativos, dashed borders |
| `accent` | `#4F7CFF` | Botao primario, links, nav ativa, progress bar |
| `highlight-gold` | `#E8C15C` | Exclusivo Game do Ano e notas de destaque |
| `difficulty-c` | `#52545C` | Dificuldade C |
| `difficulty-b` | `#6B7280` | Dificuldade B |
| `difficulty-a` | `#4F7CFF` | Dificuldade A |
| `difficulty-aa` | `#E8C15C` | Dificuldade AA |
| `difficulty-aaa` | `#E05A4E` | Dificuldade AAA |
| `success` | `#4ADE80` | Confirmacoes |
| `danger` | `#E05A4E` | Erros, acoes destrutivas |

## Tipografia (obrigatorio)

Fonte unica: Inter (Google Fonts), pesos 400/500/600/700/800.

| Elemento | Tamanho | Peso |
|---|---|---|
| Titulo de pagina | 21-22px | 700 |
| Titulo de secao | 16px | 700 |
| Numero grande (stat) | 22-30px | 700-800 |
| Corpo / label | 12.5-13.5px | 400-500 |
| Texto pequeno / meta | 10.5-11.5px | 400 (cor text-muted) |

## Componentes - especificacao obrigatoria

| Componente | Regra |
|---|---|
| Card | `bg-surface` + `border: 1px solid border` + `border-radius: 10-12px` + padding 16-20px |
| Card clicavel (capa de jogo) | hover `translateY(-3px)`, sem sombra pesada |
| Botao primario | fundo `accent` solido, texto `#0E0F12` (nunca branco), `border-radius: 7px`, weight 700 |
| Badge/tag | `border-radius: 20px` (pill), fundo `bg-surface-alt`, borda `border-subtle` |
| Barra de progresso | altura 5-6px, `border-radius: 3px`, trilho `border`, preenchimento `accent` ou cor semantica |
| Capa de jogo | `aspect-ratio: 3/4`, `border-radius: 6px`, nota no canto superior direito em `highlight-gold` sobre fundo escuro semi-transparente |
| Placeholder de capa (sem arte real) | `linear-gradient(150deg, corA, corB)` - nunca cinza chapado |

## Antes de qualquer tarefa de UI

1. Confira se a cor que vai usar esta na tabela de tokens acima - se nao esta, nao use
2. Confira se ja existe um componente shadcn/ui equivalente antes de criar um novo do zero
3. Se for tela nova, compare com o artifact "Memory Card - Dashboard" (referencia canonica) antes de divergir do padrao visual
4. Se a tela usar a cor `accent` em mais de 3 lugares - pare e reavalie a hierarquia visual antes de continuar
