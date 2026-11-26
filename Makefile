# Recipes usam apenas sh POSIX (sem bashismos), para funcionar tanto em
# macOS/Linux quanto no Windows via Git Bash. Veja CONTRIBUTING.md.

# .env is optional (CI has none). The wildcard guard matters: with a plain
# "-include .env", make would remake .env from the rule below (cp .env.example)
# before running ANY target, including in CI.
ifneq (,$(wildcard .env))
include .env
export
endif

COMPOSE        := docker compose
DEPS           := postgres elasticsearch kibana apm-server logstash filebeat
MIGRATIONS_DIR := migrations
MIGRATION_DSN  := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

.DEFAULT_GOAL := help

.PHONY: help
help: ## Lista os comandos disponíveis
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## ---------- Setup ----------

.env:
	cp .env.example .env
	@echo ">> .env criado a partir de .env.example: troque as senhas antes de subir o ambiente"

.PHONY: setup
setup: .env tools hooks ## Setup completo do ambiente de desenvolvimento (rode uma vez)

.PHONY: tools
tools: ## Instala via mise as ferramentas do mise.toml (Go, migrate, air, swag)
	mise install
	go mod download

.PHONY: hooks
hooks: ## Ativa os git hooks versionados em .githooks (pre-commit com go fmt)
	git config core.hooksPath .githooks
	@echo ">> git hooks ativados (.githooks)"

## ---------- Docker ----------

.PHONY: up
up: .env ## Sobe o ambiente completo (API + infra) via docker compose
	$(COMPOSE) up -d --build

.PHONY: deps-up
deps-up: .env ## Sobe apenas as dependências (Postgres + Elastic Stack + Logstash)
	$(COMPOSE) up -d $(DEPS)

.PHONY: down
down: ## Para os containers (os volumes nomeados são preservados)
	$(COMPOSE) down

.PHONY: logs
logs: ## Acompanha os logs dos containers
	$(COMPOSE) logs -f

## ---------- API local ----------

.PHONY: run
run: .env ## Sobe apenas a API localmente com hot reload (Air)
	air -c .air.toml

## ---------- Migrações ----------

.PHONY: migrate-up
migrate-up: ## Aplica todas as migrações pendentes
	migrate -path $(MIGRATIONS_DIR) -database "$(MIGRATION_DSN)" up

.PHONY: migrate-down
migrate-down: ## Desfaz todas as migrações
	migrate -path $(MIGRATIONS_DIR) -database "$(MIGRATION_DSN)" down -all

.PHONY: migrate-create
migrate-create: ## Cria uma nova migração (uso: make migrate-create name=add_x)
	@test -n "$(name)" || (echo "uso: make migrate-create name=<nome>" && exit 1)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## ---------- Geração de código ----------

.PHONY: mocks
mocks: ## Regenera os mocks do gomock das Ports
	go generate ./internal/application/port/...

.PHONY: swagger
swagger: ## Regenera a documentação Swagger
	swag fmt -g cmd/api/main.go
	swag init -g cmd/api/main.go -o docs --parseInternal

## ---------- Qualidade ----------

.PHONY: fmt
fmt: ## Formata todo o código Go (go fmt); falha se algum arquivo não compilar
	go fmt ./...

.PHONY: fmt-check
fmt-check: ## Falha se houver arquivo Go fora do padrão (não altera nada; útil no CI)
	@unformatted="$$(gofmt -l cmd internal)"; \
	if [ -n "$$unformatted" ]; then \
		echo "Arquivos não formatados (rode 'make fmt'):"; echo "$$unformatted"; exit 1; \
	fi

.PHONY: test-unit
test-unit: ## Roda todos os testes unitários (race detector, ordem aleatória, sem cache)
	go test -race -shuffle=on -count=1 ./...

.PHONY: test
test: test-unit ## Alias de test-unit

.PHONY: cover
cover: ## Roda os testes com relatório de cobertura
	go test -count=1 -coverprofile=coverage.out ./internal/domain/... ./internal/application/...
	go tool cover -func=coverage.out

.PHONY: vet
vet: ## Análise estática
	go vet ./...

## ---------- Carga ----------

.PHONY: load-test
load-test: ## Teste de carga com k6 (opcional: DURATION=5m RATE=50 DELAY_VUS=10)
	k6 run -e BASE_URL=http://localhost:$(or $(HTTP_PORT),8080) \
		$(if $(DURATION),-e DURATION=$(DURATION)) \
		$(if $(RATE),-e RATE=$(RATE)) \
		$(if $(DELAY_VUS),-e DELAY_VUS=$(DELAY_VUS)) \
		scripts/load.js
