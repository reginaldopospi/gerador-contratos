# Backend (Go + SQLite)

API REST para autenticacao, contratos versionados, corretores e clausulas.

## Requisitos

- Go 1.23+

## Variaveis de ambiente

Use `backend/.env.example` como referencia.

- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`: habilitam envio de e-mail no fluxo de recuperacao.
- `PASSWORD_RESET_URL`: URL base usada no e-mail para montar o link com `token`.
- `REQUIRE_REGISTRATION_APPROVAL`: quando `1`, novos cadastros entram como pendentes.
- `PLATFORM_ADMIN_USERNAME`, `PLATFORM_ADMIN_EMAIL`, `PLATFORM_ADMIN_PASSWORD`, `PLATFORM_ADMIN_NAME`, `PLATFORM_TENANT_NAME`: definem o admin que aprova cadastros pendentes (login pode usar username ou email).
  - Em ambiente local, existem valores padrao para facilitar testes; em producao, configure valores proprios.

Observacao:
- A API carrega automaticamente variaveis de `.env` (na pasta `backend` ou na raiz do projeto) quando presentes.

## Como rodar

1. Instale Go 1.23+.
2. Na raiz do projeto, acesse o diretorio `backend` e execute:

```bash
cd backend
go run ./cmd/api
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
  - `GET /auth/pending-registrations` (Bearer, somente admin da plataforma)
  - `POST /auth/pending-registrations/{userID}/approve` (Bearer, somente admin da plataforma)
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
cd backend
go test ./...
```

## Observacoes

- Migrations sao executadas automaticamente ao subir a API.
- Em `APP_ENV=dev`, `forgot-password` retorna `dev_reset_token` no payload para facilitar testes.
- Se SMTP estiver configurado, `forgot-password` envia token por e-mail; em caso de falha no envio, a API retorna erro.
- Quando `REQUIRE_REGISTRATION_APPROVAL=1`, o cadastro inicial da imobiliaria nao recebe token imediatamente e precisa ser aprovado pelo admin da plataforma.
