---
name: frontend-boaspraticas
description: Padroes obrigatorios de qualidade de codigo para o frontend do Memory Card. Usar sempre que escrever, revisar ou finalizar codigo.
---

# Memory Card - Boas Praticas de Frontend

Repositorio publico, peca de portfolio (objetivo declarado: demonstrar qualidade de engenharia para se posicionar como engenheiro backend senior com IA como diferencial). O frontend precisa refletir esse mesmo padrao de cuidado.

## Hard Rules - nunca violar

- Zero `div`/`span` com `onClick` fazendo o papel de botao - sempre `<button>` real
- Zero botao so-icone sem `aria-label`
- Zero commit fora do padrao Conventional Commits (`feat:`, `fix:`, `chore:`, `refactor:`, `test:`)
- Zero codigo morto ou comentado esquecido no repositorio
- Zero `strict: false` no `tsconfig.json`
- Zero teste testando detalhe de implementacao (classe CSS, estrutura de DOM interna) - testar comportamento observavel pelo usuario
- Zero termo de dominio em ingles quando ja existe o termo em portugues no backend (`jogoZerado`, nunca `completedGame`) - dominio do produto e em portugues, codigo tecnico generico pode ficar em ingles

## Nomenclatura (obrigatoria)

| Tipo | Padrao | Exemplo |
|---|---|---|
| Componente | PascalCase | `GameCard.tsx` |
| Hook | camelCase, prefixo `use` | `useAuth.ts` |
| Funcao/variavel | camelCase | `buscarJogos` |
| Tipo/Interface | PascalCase, sem prefixo `I` | `Usuario` |
| Arquivo de service | camelCase | `gamesService.ts` |

## Limites de linhas

Ver skill `frontend-arquitetura` - os limites por tipo de arquivo sao os mesmos, nao duplicar aqui.

## Testes

- Componente com logica nao-trivial (formulario, validacao, fluxo condicional) exige teste - Vitest + Testing Library
- Fluxos criticos (login, cadastro, recuperacao de senha) exigem cobertura e2e quando a suite e2e do projeto existir

## Qualidade de codigo

- ESLint + Prettier configurados e rodando no CI (mesmo pipeline do GitHub Actions que valida o backend Go)
- Componente que passa de ~150-200 linhas (ver `frontend-arquitetura`) deve ser quebrado antes de seguir

## Antes de considerar uma tela pronta - checklist

1. Rodar lint - zero warning
2. Rodar testes - todos passando
3. Conferir responsividade basica (desktop + mobile)
4. Conferir que segue os tokens da skill `frontend-design-system` - nenhuma cor hardcoded fora da tabela
5. Conferir acessibilidade basica: botao/link/input reais, `aria-label` em icone-so
6. Commit seguindo Conventional Commits
