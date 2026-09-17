# Estratégia e Padrões de Testes — Memory Card

Este documento detalha os padrões, arquitetura, convenções e metas de testes unitários e de integração para o backend em Go (`apps/api`) do Memory Card.

---

## 1. Pirâmide e Arquitetura de Testes

O projeto segue as convenções definidas nas skills do repositório (`backend-boaspraticas-go` e `backend-arquitetura-go`):

| Camada | Tipo de Teste | Dependências Externas | Onde Vive |
|---|---|---|---|
| **Config** | Unitário | Nenhuma (usa `t.Setenv`) | `internal/config/*_test.go` |
| **Handler (Gin)** | Unitário | Service mockado via interface | `internal/handler/*_test.go` |
| **Service (Regra de Negócio)** | Unitário | Repository mockado via interface | `internal/service/*_test.go` |
| **Repository (sqlc) / Integração** | E2E / Integração | PostgreSQL real via `testcontainers-go` (MEMOR-20) | `internal/repository/*_test.go` ou `tests/e2e` |

### Hard Rules para Testes em Go
1. **Nenhum banco de dados real em testes unitários**: A camada `service` nunca acessa banco ou conecta em rede nos testes unitários.
2. **Nomenclatura**: Testes devem seguir o padrão `Test<Funcao>_<Cenario>` (ex: `TestLoad_ValoresValidos`, `TestHealthHandler_Check`, `TestHealthService_CheckDatabase`).
3. **Table-driven tests**: Sempre que houver mais de duas variações do mesmo cenário, estruture o teste em tabela com `t.Run(tc.name, ...)`.
4. **Erros tipados**: Validações de erro devem usar `errors.Is` ou `errors.As`, nunca comparação ingênua de strings.

---

## 2. Padrão de Mock de Repository (MEMOR-27)

Como o Memory Card não utiliza ORM e baseia sua persistência em interfaces geradas/definidas pelo consumidor, os testes da camada de serviço injetam mocks de repositório.

### Decisão: Mock Manual (Struct com Campos de Função)

Optamos pelo padrão de **Mock Manual** em Go (struct com campos do tipo `func`), em detrimento de geradores como `mockgen` ou `moq`:

#### Vantagens:
- **Zero Dependências Externas**: Não exige instalação de binários adicionais no CI (`github-actions`) ou na máquina do desenvolvedor.
- **Transparência e Legibilidade**: Código 100% idiomático Go, sem reflexão ou metadados mágicos gerados por ferramentas terceiras.
- **Flexibilidade com Closures**: Cada caso de teste pode customizar o retorno, simular erros específicos ou capturar argumentos passados para o repositório.
- **Aderência ao Interface Segregation Principle**: Estimula que cada service declare interfaces pequenas e estritamente necessárias (1 a 3 métodos).

### Estrutura do Padrão

```go
// 1. Interface definida no pacote consumidor (service)
type UsuarioRepository interface {
    BuscarPorEmail(ctx context.Context, email string) (*Usuario, error)
}

// 2. Struct de Mock com campos funcionais
type mockUsuarioRepository struct {
    buscarPorEmailFn func(ctx context.Context, email string) (*Usuario, error)
}

func (m *mockUsuarioRepository) BuscarPorEmail(ctx context.Context, email string) (*Usuario, error) {
    if m.buscarPorEmailFn != nil {
        return m.buscarPorEmailFn(ctx, email)
    }
    return nil, errors.New("BuscarPorEmail não implementado no mock")
}

// 3. Uso nos testes
func TestUsuarioService_BuscarPorEmail_Sucesso(t *testing.T) {
    mock := &mockUsuarioRepository{
        buscarPorEmailFn: func(ctx context.Context, email string) (*Usuario, error) {
            return &Usuario{ID: 1, Email: email}, nil
        },
    }
    svc := NewUsuarioService(mock)
    usuario, err := svc.BuscarPorEmail(context.Background(), "user@example.com")
    // asserções...
}
```

> **Exemplo Executável de Referência**: Consulte [`internal/service/exemplo_test.go`](../apps/api/internal/service/exemplo_test.go) para um exemplo completo e funcional com testes de casos felizes, erros de domínio e falhas de infraestrutura.

---

## 3. Meta de Cobertura de Código (MEMOR-28)

### Metas Definidas

- **Meta Mínima no CI (`./...` global)**: **20%**
- **Meta dos Pacotes de Regra de Negócio e Configuração**: **80%** (atualmente em **100%**)

### Racional

1. **Estágio Atual do Repositório (Sprint 1 / início Sprint 2)**:
   - Os pacotes com lógica de negócio e configuração escrita manualmente (`internal/config`, `internal/handler`, `internal/service`) possuem **100.0% de cobertura de statements**.
   - O repositório já contém as queries compiladas pelo `sqlc` em `internal/repository/db/*` (~70 statements) e o entrypoint do servidor em `cmd/server/main.go` (~24 statements).
   - Como o código do `sqlc` é gerado automaticamente para execução contra PostgreSQL real, ele **não deve ser inflado artificialmente** com mocks em testes unitários. Sua validação é de responsabilidade dos testes de integração com `testcontainers-go` (história MEMOR-20).
   - Por essa razão matemática (24 statements testados de 118 totais), a cobertura global do módulo `./...` fica em torno de **22.1%**.
   - A meta global no CI foi definida em **20%** para garantir que qualquer regressão em `config`, `handler` ou `service` quebre o build imediatamente, sem travar o pipeline artificialmente.

2. **Roadmap de Evolução da Meta**:
   - **Sprint 2 (MEMOR-21 em diante)**: À medida que os services de domínio (`auth`, `usuarios`, `jogos`) forem implementados com seus respectivos mocks, o volume de statements na camada `service` aumentará, elevando a meta global para 60-70%.
   - **MEMOR-20 (Testes de Integração)**: Com o setup do `testcontainers-go`, as queries de `internal/repository/db` passarão a ser cobertas, elevando a meta para >80%.

---

## 4. Comandos de Desenvolvimento

A partir da raiz do monorepo:

```bash
# Executar todos os testes unitários
make test

# Executar testes gerando relatório de cobertura
make test-coverage

# Visualizar cobertura detalhada no navegador
make test-coverage-html

# Executar linters e checagens estáticas
make lint
make vet
```

Ou diretamente no diretório `apps/api`:

```bash
go test -v ./...
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
