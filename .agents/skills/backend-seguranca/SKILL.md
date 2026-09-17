---
name: backend-seguranca
description: Regras obrigatorias de seguranca para autenticacao e tratamento de dados sensiveis no backend do Memory Card. Usar sempre que mexer em auth, senha, token ou dado de usuario.
---

# Memory Card - Seguranca (Backend)

Multi-usuario simples (Lucas + circulo proximo), mas o repositorio e publico - o codigo de auth vai ser lido por qualquer visitante do GitHub. Seguranca aqui nao e so protecao, e tambem prova de competencia tecnica.

## Hard Rules - nunca violar

- Zero senha em texto puro em log, resposta de API, ou commit - sempre hash bcrypt (custo minimo 10)
- Zero token JWT com segredo hardcoded no codigo - sempre de variavel de ambiente, nunca commitado
- Zero mensagem de erro de login revelando se o email existe ("email nao encontrado" vs "senha errada") - sempre mensagem generica ("credenciais invalidas")
- Zero endpoint de recuperacao de senha revelando se o email esta cadastrado - resposta sempre identica (200 generico) independente de existir ou nao
- Zero refresh token reutilizavel apos logout ou apos troca de senha - invalidar todos os tokens ativos do usuario nessas acoes
- Zero dado sensivel (senha, token) em query param de URL - sempre body ou header
- Zero validacao de senha so no frontend - toda regra de forca de senha e reforcada no backend tambem, frontend e so UX
- Zero rate limit ausente em endpoint de auth publico (login, registro, solicitar-reset) - protecao contra brute-force e abuso

## JWT

- Access token: vida curta (15min como referencia inicial, confirmar com o projeto)
- Refresh token: vida longa (7 dias como referencia inicial), armazenado de forma que permita invalidacao (nao e so verificar assinatura - precisa checar se ainda esta ativo, ex: tabela de refresh tokens ou denylist)
- Claims minimas necessarias no payload - nunca colocar dado sensivel no JWT (ele nao e criptografado, so assinado)

## Tokens de reset de senha

- Token unico, uso unico, expiracao curta (30min como referencia das historias de auth ja criadas no Huly)
- Token invalidado apos uso, mesmo que a redefinicao falhe por outro motivo
- Nunca reaproveitar o mesmo token gerado para requests diferentes

## Validacao de input

- Toda entrada de usuario validada no backend antes de tocar o banco (formato de email, forca de senha, tamanho de campos)
- Protecao contra SQL injection e automatica com sqlc (queries parametrizadas) - nunca concatenar string SQL manualmente, mesmo em caso excepcional

## Antes de considerar uma feature de auth pronta - checklist

1. Nenhuma mensagem de erro revela informacao que ajude um atacante a enumerar usuarios
2. Rate limit configurado nos endpoints publicos de auth
3. Todos os refresh tokens sao invalidados nas acoes que exigem isso (troca de senha, reset de senha)
4. Segredos (JWT secret, credenciais IGDB) vem de env var, nunca hardcoded
5. Testes cobrindo os casos de abuso (credenciais erradas, token expirado, rate limit atingido)
