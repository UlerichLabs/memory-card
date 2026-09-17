.PHONY: all build vet test test-coverage test-coverage-html test-e2e lint

all: build vet test

build:
	cd apps/api && go build ./...

vet:
	cd apps/api && go vet ./...

test:
	cd apps/api && go test -v ./...

test-coverage:
	cd apps/api && go test -v -coverprofile=coverage.out ./...
	cd apps/api && go tool cover -func=coverage.out

test-coverage-html:
	cd apps/api && go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório HTML gerado em apps/api/coverage.html"

test-e2e:
	cd apps/api && go test -v -tags=e2e ./tests/e2e/...

lint:
	cd apps/api && golangci-lint run --config ../../.golangci-lint.yml
