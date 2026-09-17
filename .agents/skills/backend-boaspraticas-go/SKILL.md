---
name: backend-boaspraticas-go
description: Padroes obrigatorios de qualidade de codigo Go para o backend do Memory Card. Usar sempre que escrever, revisar ou finalizar codigo Go.
---

# Memory Card - Boas Praticas de Backend (Go)

Repositorio publico, peca de portfolio (objetivo declarado: demonstrar qualidade de engenharia idiomatica em Go para se posicionar como engenheiro backend senior com IA como diferencial). Este e o codigo que mais vai ser julgado por quem visitar o repo.

## Hard Rules - nunca violar

- Zero `panic` para erro esperado (validacao, not found, etc.) - panic so para erro de programacao irrecuperavel na inicializacao
- Zero erro engolido (`if err != nil { }` vazio ou so log sem propagar quando quem chamou precisa saber)
- Zero `err != nil` sem `%w` no `fmt.Errorf` quando o erro e propagado (preserva a chain para `errors.Is`/`errors.As`)
- Zero nome de variavel de uma letra fora de escopos curtos (loop index e ok, mais que isso nao)
- Zero comentario descrevendo o que o codigo faz (obvio pela leitura) - comentario so para o porque de uma decisao nao-obvia
- Zero interface criada antes de ter um segundo caso de uso real ou uma necessidade de mock em teste - Go: aceite interfaces, retorne structs
- Zero `golangci-lint` rodando com warning ignorado sem justificativa em comentario

## Error handling (idiomatico)

- Erros de dominio sao tipados (`var ErrEmailJaCadastrado = errors.New("email ja cadastrado")` ou struct de erro customizado), nunca string magica comparada por `==`
- Handler mapeia erro de dominio para status HTTP com `errors.Is`/`errors.As`, nunca checando mensagem de string
- Erro de infraestrutura (banco fora do ar, etc.) nunca vaza detalhe interno na resposta HTTP - logar completo, responder generico

## Naming

| Tipo | Padrao | Exemplo |
|---|---|---|
| Pacote | minusculo, sem underscore | `service`, nao `Service` ou `my_service` |
| Struct exportada | PascalCase | `AuthService` |
| Interface | PascalCase, geralmente termina em -er quando for comportamento unico | `UsuarioRepository` (dominio) ou `Reader` (comportamento) |
| Funcao/metodo nao exportado | camelCase | `validarSenha` |
| Constante | PascalCase ou UPPER_SNAKE so se for convencao de pacote externo | `StatusAtivo` |

## Testes

- Service: teste unitario com repository mockado via interface (nunca banco real)
- Fluxo completo: teste e2e com testcontainers-go (Postgres real, migrations aplicadas)
- Nome de teste no formato `TestFuncao_Cenario` (`TestLogin_CredenciaisInvalidas`)
- Tabela de casos (`table-driven tests`) quando houver mais de 2 variacoes do mesmo cenario

## Qualidade de codigo

- `golangci-lint` rodando no CI, zero warning tolerado sem justificativa
- `go vet` e `gofmt` (ou `goimports`) sempre limpos antes de commit
- Doc comment (`// NomeDaFuncao faz X`) em toda funcao/struct exportada - e o padrao que a comunidade Go espera e o que aparece no godoc

## Antes de considerar uma feature pronta - checklist

1. `go build ./...` e `go vet ./...` limpos
2. `golangci-lint run` sem warning
3. Testes unitarios da service escritos e passando
4. Erros de dominio tipados, nunca string solta comparada
5. Nenhum `context.Context` faltando em funcao com I/O
6. Commit seguindo Conventional Commits
