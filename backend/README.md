# Backend (Go + SQLite)

API REST para autenticacao, contratos versionados, corretores e clausulas.

## Requisitos

- Go 1.23+

## Variaveis de ambiente

Use `backend/.env.example` como referencia.

## Como rodar

1. Instale Go 1.23+.
2. Na raiz do projeto, execute:

```bash
go run ./backend/cmd/api
```

3. Healthcheck:

```bash
GET http://localhost:8080/health
```

## Rotas principais

Base: `/api/v1`

- Auth
  - `POST /auth/register`
  - `POST /auth/login`
  - `POST /auth/refresh`
  - `POST /auth/forgot-password`
  - `POST /auth/reset-password`
  - `GET /auth/me` (Bearer)
  - `POST /auth/users` (Bearer)
- Contratos
  - `GET /contracts`
  - `POST /contracts`
  - `GET /contracts/{id}`
  - `POST /contracts/{id}/versions`
  - `GET /contracts/{id}/versions/{versionNumber}`
  - `GET /contracts/{id}/preview`
  - `POST /contracts/preview`
- Corretores
  - `GET /brokers`
  - `POST /brokers`
  - `PUT /brokers/{id}`
  - `DELETE /brokers/{id}`
- Clausulas
  - `GET /clauses`
  - `POST /clauses`
  - `DELETE /clauses/{id}`
- CNPJ (Receita Federal via integracao)
  - `GET /cnpj/{cnpj}`

## Testes

```bash
go test ./backend/...
```

## Observacoes

- Migrations sao executadas automaticamente ao subir a API.
- Em `APP_ENV=dev`, `forgot-password` retorna `dev_reset_token` no payload para facilitar testes.
