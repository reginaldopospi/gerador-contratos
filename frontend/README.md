# Frontend (Svelte)

Frontend SPA em Svelte integrado com a API Go.

## Requisitos

- Node.js 20+
- npm

## Variaveis de ambiente

Crie `frontend/.env` (opcional):

```bash
VITE_API_BASE_URL=/api/v1
```

Observacao:
- Em desenvolvimento, o Vite faz proxy de `/api` para `http://localhost:8080`.
- Para backend remoto, use URL absoluta em `VITE_API_BASE_URL` (ex.: `https://api.seudominio.com/api/v1`).

## Como rodar

```bash
cd frontend
npm install
npm run dev
```

Aplicacao: `http://localhost:5173`

## Funcionalidades implementadas

- Cadastro inicial da imobiliaria (tenant + admin)
- Login e logout
- Recuperacao e redefinicao de senha
- Dashboard
- Lista e criacao de contratos
- Edicao de contrato por versao (JSON)
- Previa juridica automatica
- Cadastro e exclusao de corretores
- Cadastro e exclusao de clausulas

## Testes

```bash
npm run test
```
