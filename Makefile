SHELL := /bin/sh

GO ?= go
NPM ?= npm

.PHONY: help setup backend-deps frontend-deps backend-dev frontend-dev dev test test-backend test-frontend build build-backend build-frontend clean

help: ## Lista os comandos disponiveis
	@echo ""
	@echo "Comandos principais:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Dica (Windows): se necessario, rode com NPM=npm.cmd"
	@echo "Ex.: make setup NPM=npm.cmd"
	@echo ""

setup: backend-deps frontend-deps ## Instala dependencias locais

backend-deps: ## Gera/atualiza dependencias Go (go.sum)
	@cd backend && $(GO) mod download

frontend-deps: ## Instala dependencias do frontend
	@$(NPM) --prefix frontend install

backend-dev: backend-deps ## Sobe o backend localmente (porta 8080)
	@cd backend && $(GO) run ./cmd/api

frontend-dev: ## Sobe o frontend localmente (porta 5173)
	@$(NPM) --prefix frontend run dev

dev: backend-deps ## Sobe backend + frontend juntos (Ctrl+C encerra ambos)
	@trap 'kill 0' INT TERM EXIT; \
		cd backend && $(GO) run ./cmd/api & \
		$(NPM) --prefix frontend run dev & \
		wait

test: test-backend test-frontend ## Roda todos os testes

test-backend: ## Roda testes do backend
	@cd backend && $(GO) test ./...

test-frontend: ## Roda testes do frontend
	@$(NPM) --prefix frontend run test

build: build-backend build-frontend ## Gera build de backend e frontend

build-backend: ## Compila binario do backend em backend/bin/api
	@mkdir -p backend/bin
	@cd backend && $(GO) build -o ./bin/api ./cmd/api

build-frontend: ## Gera build de producao do frontend
	@$(NPM) --prefix frontend run build

clean: ## Remove artefatos locais (db, binarios e dist frontend)
	@rm -rf backend/bin backend/data frontend/dist
