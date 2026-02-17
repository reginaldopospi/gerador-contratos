# Gerador de Contratos (Go + Svelte)

Solucao completa para criacao e gestao de contratos imobiliarios.

## Stack

- Backend: Go
- Frontend: Svelte
- Banco: SQLite

## Estrutura

- `backend/` API REST, regras de negocio, migrations SQLite e testes.
- `frontend/` aplicacao web Svelte com fluxo completo de produto.
- `docs/README.md` documentacao funcional extraida do app Python legado.
- `app.py` legado Streamlit (referencia de regras originais).

## Subir o projeto

Opcao recomendada com Makefile:

```bash
make setup
make dev
```

### 1) Backend

```bash
go run ./backend/cmd/api
```

### 2) Frontend

```bash
cd frontend
npm install
npm run dev
```

## Fluxos de produto implementados

- Cadastro de imobiliaria + usuario admin
- Login / refresh / logout
- Recuperacao de senha
- CRUD de contratos com versionamento
- Previa juridica baseada em regras extraidas do legado
- CRUD de corretores
- CRUD de clausulas

## Qualidade

- Arquitetura em camadas (repository/service/use case)
- Migrations versionadas
- Testes unitarios em modulos criticos
