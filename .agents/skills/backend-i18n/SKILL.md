---
name: backend-i18n
description: Padroes obrigatorios de internacionalizacao (PT/EN) para o backend do Memory Card. Usar sempre que escrever mensagem de erro, validacao ou qualquer texto voltado ao usuario final.
---

# Memory Card - Internacionalizacao (i18n)

O Memory Card suporta multiplos idiomas de verdade (PT-BR e EN por enquanto, escopo pode crescer). Usuario escolhe o idioma; toda mensagem voltada a ele passa por traducao - nunca texto fixo em portugues (ou qualquer idioma) hardcoded na resposta da API.

## Hard Rules - nunca violar

- Zero mensagem de erro ou texto voltado ao usuario escrito direto em string no codigo Go - sempre uma chave de traducao resolvida no momento da resposta
- Zero logica de negocio dependente do idioma (idioma nunca deve mudar o que o sistema faz, so como ele se comunica)
- Zero mistura de idioma dentro da mesma resposta (nao pode ter um campo traduzido e outro nao)
- Zero data/numero formatado sem considerar o locale do usuario quando exibido como texto (ex: formato de data DD/MM/AAAA vs MM/DD/AAAA)

## Onde o idioma do usuario vive

- Coluna `idioma` (ou `locale`) na tabela `usuarios`, default `pt-BR`
- Toda resposta de erro/mensagem usa esse valor para resolver a traducao - vem do usuario autenticado (JWT claim ou lookup) ou de um header `Accept-Language` como fallback para rotas publicas (login, registro, antes de saber quem e o usuario)

## Estrutura de chaves de traducao

Convencao de chave: `<dominio>.<contexto>.<caso>`, espelhando os codigos `RN_*` ja documentados nas historias do Huly:

```
auth.register.email_taken          -> RN_CADASTRO_01
auth.register.weak_password        -> RN_CADASTRO_02
auth.login.invalid_credentials     -> RN_LOGIN_01
auth.password_reset.token_expired  -> RN_RESET_02
```

Arquivos de traducao (`internal/i18n/pt-BR.json`, `internal/i18n/en.json` ou equivalente) mantidos em paralelo - toda chave nova precisa existir nos dois arquivos antes do merge, nunca so em um idioma.

## Relacao com o formato de erro da API

O campo `codigo` do envelope de erro (ver skill `backend-api-design`) e a MESMA chave estavel usada para resolver a traducao. O campo `mensagem` e sempre o texto ja traduzido no idioma do usuario, nunca a chave crua.

## Antes de escrever uma mensagem nova

1. Nunca escreva a string direto no `service` ou `handler` - crie a chave de traducao primeiro
2. Adicione a chave em TODOS os arquivos de idioma suportados na mesma tarefa, nunca deixe para depois
3. Se a mensagem depender de uma regra de negocio ja documentada como `RN_*` no Huly, reuse esse identificador como base do nome da chave para rastreabilidade
4. Confirme que o resolvedor de idioma tem um fallback (idioma nao suportado ou ausente cai para `pt-BR`)
