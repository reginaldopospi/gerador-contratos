# Documentacao da aplicacao atual e proposta de solucao completa

## 1) O que a aplicacao atual e

A aplicacao atual e um sistema monolitico em `app.py` (Streamlit) para montar contratos imobiliarios com regras juridicas dinamicas.

Objetivo principal:
- Coletar dados do contrato em formato wizard.
- Gerar uma previa textual juridica com clausulas condicionais.
- Salvar contrato versionado por imobiliaria no Supabase.

Tecnologias atuais identificadas:
- UI e fluxo: Streamlit.
- Persistencia: Supabase (`corretores` e `contratos`).
- APIs externas: ViaCEP (endereco), ReceitaWS (CNPJ).
- Autenticacao: usuarios e senhas em `st.secrets`.

## 2) Leitura funcional do app.py (o que ele faz hoje)

### 2.1 Autenticacao e sessao

- Login simples por usuario/senha em `st.secrets[auth][users]`.
- Estado de sessao em `st.session_state`.
- Flag `contrato_dirty` para detectar alteracoes nao salvas.
- Logout com confirmacao para sair sem salvar.

### 2.2 Wizard de etapas

Etapas visiveis:
1. Localizar contrato (oculta no fluxo normal, usada para buscar contrato salvo).
2. Identificacao do contrato.
3. Imovel.
4. Parte Vendedora.
5. Parte Compradora.
6. Preco e Chaves.
7. Parcelamento (aparece de forma dinamica).
8. Permutas/Daçao (aparece de forma dinamica).
9. Previa de Contrato (clausulas).

Etapas ocultas:
- Cadastro de corretor.
- Senha Admin.
- Admin de corretores.
- Admin de clausulas (step existe, mas tela nao foi implementada no fluxo final).

### 2.3 Cadastros e formularios

- Endereco com CEP e autopreenchimento via ViaCEP.
- Parte PF:
  - Nome, nacionalidade, RG, CPF, profissao, estado civil, regime de bens.
  - Endereco.
  - Conjuge/companheiro obrigatorio para casado(a)/uniao estavel.
- Parte PJ:
  - CNPJ com busca ReceitaWS.
  - Razao social e endereco da empresa.
  - Representante legal.
- Listas dinamicas:
  - Multiplos vendedores.
  - Multiplos compradores.
  - Multiplos corretores de venda e captacao.

### 2.4 Corretor e comissao

- Corretor selecionado a partir do cadastro persistido.
- Cadastro rapido de corretor durante o fluxo.
- Tela admin para editar/excluir corretores.
- Percentual por corretor e dados bancarios.

### 2.5 Regra de preco, pagamento e clausulas

- Campos de composicao do preco:
  - Total, financiamento, FGTS, entrada, sinal, recurso proprio, carta de credito, subsidio, parcelamento, outros.
- Regras juridicas condicionais:
  - Com e sem financiamento.
  - Com e sem FGTS.
  - Com e sem sinal.
  - Com e sem parcelamento.
  - Com e sem alienacao fiduciaria.
  - Tipo de contrato (venda/compra x cessao).
- Montagem de clausulas por funcoes dedicadas (muitas regras vindas de planilha).
- Numeracao dinamica de clausulas/subclausulas.

### 2.6 Persistencia e versionamento

- Salva contrato no Supabase em `contratos` com versao incremental (`max+1`).
- Contrato e salvo como JSON completo (`dados`).
- Busca da ultima versao por imobiliaria + numero do contrato.
- Carregamento de contrato salvo para o estado da sessao.

## 3) Estrutura de dados inferida

### 3.1 Entidades principais

- Usuario (atual: apenas em secret).
- Imobiliaria/Tenant (derivado do usuario logado).
- Contrato.
- Versao de contrato.
- Imovel.
- Parte (PF/PJ).
- Corretor.
- Clausulas e configuracoes de clausula.

### 3.2 Tabelas atuais no Supabase (pelo codigo)

- `corretores`
  - `id`, `imobiliaria`, `nome`, `cpf`, `banco`, `agencia`, `conta`, `pix`.
- `contratos`
  - `id`, `imobiliaria`, `numero_contrato`, `versao`, `numero_versao_label`, `dados`, `created_at`, `updated_at`.

## 4) Lacunas encontradas para uma solucao profissional

### 4.1 Seguranca e identidade

- Sem cadastro de usuario.
- Sem recuperacao de senha.
- Sem hash de senha no fluxo atual (senha em secret).
- Sem MFA, sem auditoria de login, sem bloqueio por tentativa.

### 4.2 Produto e operacao

- Nao existe painel de contratos (lista, filtros, status, ordenacao).
- Nao existe workflow de status (rascunho, em revisao, aprovado, assinado).
- Nao existe trilha de auditoria completa por campo alterado.
- Nao existe notificacao por email para eventos importantes.

### 4.3 Documento final

- Existem templates `.docx` no repositorio, mas o fluxo atual nao gera arquivo final DOCX/PDF.
- Dependencia `python-docx` aparece em `requirements.txt`, mas nao esta sendo usada no `app.py` atual.

### 4.4 Qualidade tecnica

- Codigo em arquivo unico muito grande.
- Funcoes duplicadas e sobreposicoes.
- Step `admin_clausulas` declarado, mas sem tela final implementada.
- Possivel inconsistência de chave em regra de clausula (`tipo_imovel` vs `imovel__tipo`).
- Sem suite de testes automatizados.

## 5) Solucao alvo completa (Backend Go + Frontend Svelte)

## 5.1 Objetivo da nova arquitetura

Separar a aplicacao em camadas claras para escalar manutencao, seguranca e evolucao:
- Backend Go: API de negocio, autenticacao, versionamento, regras e geracao de documentos.
- Frontend Svelte: UX moderna, wizard robusto, validacoes ricas, painel operacional.
- Banco SQLite: armazenamento principal com migrations e constraints.

## 5.2 Arquitetura proposta

### Backend (Go)

Padrao recomendado:
- Clean Architecture + DDD leve.
- Camadas: `domain`, `application`, `infrastructure`, `interfaces/http`.

Pacotes principais:
- `auth`: login, refresh token, logout, recuperacao de senha.
- `users`: cadastro, perfil, roles e permissoes.
- `tenants`: imobiliarias e isolamento de dados.
- `contracts`: contrato, versoes, status e historico.
- `parties`: PF/PJ e validacoes.
- `properties`: dados do imovel e endereco.
- `brokers`: cadastro/admin de corretores.
- `clauses`: motor de clausulas parametrizadas.
- `documents`: geracao DOCX/PDF com templates.
- `audit`: trilha de auditoria.

Tecnologias sugeridas:
- Router: `chi`.
- ORM/Query: `sqlc` ou `gorm` (preferencia por sqlc para tipagem forte).
- Migrations: `golang-migrate`.
- Auth: JWT access + refresh rotativo.
- Senhas: `bcrypt`.

### Frontend (Svelte)

Recomendado:
- `SvelteKit` + TypeScript.
- Estado por feature (stores locais por modulo).
- Formularios com validacao de schema (Zod).
- Componentizacao por dominio.

Modulos de tela:
- Login.
- Cadastro de usuario.
- Recuperacao de senha.
- Dashboard.
- Lista de contratos.
- Wizard de contrato (autosave + validacao por etapa).
- Admin de corretores.
- Admin de clausulas.
- Perfil e seguranca da conta.

## 5.3 Banco SQLite (modelo base)

Tabelas essenciais:
- `users`
  - id, tenant_id, name, email (unique), password_hash, role, is_active, created_at, updated_at.
- `password_reset_tokens`
  - id, user_id, token_hash, expires_at, used_at, created_at.
- `sessions`
  - id, user_id, refresh_token_hash, ip, user_agent, expires_at, revoked_at.
- `tenants`
  - id, nome_fantasia, cnpj, created_at.
- `brokers`
  - id, tenant_id, nome, cpf, banco, agencia, conta, pix, created_at, updated_at.
- `contracts`
  - id, tenant_id, numero, tipo, status, created_by, created_at, updated_at.
- `contract_versions`
  - id, contract_id, version_number, data_json, created_by, created_at.
- `clause_templates`
  - id, tenant_id (nullable para global), key, title, content, is_active, updated_at.
- `audit_logs`
  - id, tenant_id, actor_user_id, entity, entity_id, action, before_json, after_json, created_at.

## 5.4 Fluxos de autenticacao e acesso (completos)

### Cadastro de usuario
- Admin cria usuario interno da imobiliaria.
- Opcional: auto cadastro com convite por email.

### Login
- Email + senha.
- JWT access curto e refresh com rotacao.
- Bloqueio progressivo por tentativas falhas.

### Recuperacao de senha
- Tela "Esqueci minha senha".
- Token unico com expiracao curta.
- Token armazenado com hash no banco.
- Tela de redefinicao + invalidacao de sessoes antigas.

### Controle de acesso
- RBAC minimo:
  - `admin` (gestao completa).
  - `gestor` (contratos + corretores + clausulas).
  - `operador` (cria/edita contratos, sem admin global).

## 5.5 UX/UI pro fluxo profissional

### Estrutura de navegacao
- Sidebar fixa com:
  - Dashboard.
  - Contratos.
  - Novo contrato.
  - Corretores.
  - Clausulas.
  - Configuracoes.

### Wizard de contrato (novo)
- Stepper superior com progresso.
- Autosave por etapa.
- Indicadores de erro por etapa.
- Painel lateral de resumo em tempo real.
- Acao "Salvar rascunho" sempre visivel.
- Acoes finais:
  - Gerar previa.
  - Gerar DOCX.
  - Gerar PDF.
  - Criar nova versao.

### Experiencia operacional
- Lista de contratos com filtros:
  - numero, parte, corretor, status, data.
- Acoes rapidas:
  - duplicar versao,
  - comparar versoes,
  - arquivar,
  - exportar.

## 5.6 Regras de negocio migradas do app.py

Motor de regras deve manter:
- Qualificacao PF/PJ.
- Regras de conjuge por estado civil.
- Regras de composicao de pagamento.
- Regras condicionais de clausulas (financiamento, FGTS, sinal, alienacao, parcelamento etc).
- Numeracao dinamica de clausulas.
- Multi vendedores/compradores/corretores.

## 5.7 Requisitos nao funcionais

- Logs estruturados.
- Auditoria por alteracao sensivel.
- Backup automatizado do SQLite.
- Healthcheck de API.
- Monitoramento basico (latencia/erros).
- Politica LGPD:
  - minimizacao,
  - trilha de acesso,
  - anonimização quando aplicavel.

## 6) Roadmap sugerido

Fase 1 (MVP tecnico):
- Auth completo (login, refresh, recovery).
- CRUD de contratos com versionamento.
- CRUD de corretores.
- Wizard basico com validacao por etapa.

Fase 2 (paridade de regras):
- Migrar todas as regras de clausulas do Python para Go.
- Previa HTML fiel ao modelo atual.
- Admin de clausulas funcional.

Fase 3 (producao):
- Geracao DOCX/PDF.
- Auditoria completa.
- Dashboard e filtros avancados.
- Hardening de seguranca.

## 7) Resultado esperado

Com Go + Svelte + SQLite, a aplicacao deixa de ser um script monolitico e vira um produto com:
- arquitetura escalavel,
- autenticacao profissional,
- experiencia de uso consistente,
- governanca de dados,
- base segura para crescer funcionalidades.

## 8) Status de implementacao no repositorio

Ja implementado nesta migracao:
- `backend/` com API Go, SQLite, migrations, auth (cadastro/login/refresh/forgot/reset), contratos versionados, corretores e clausulas.
- `frontend/` em Svelte com fluxos de produto (login, cadastro, recuperacao, dashboard, contratos, corretores, clausulas).
- testes unitarios no backend (auth, contracts, rules) e frontend (utils com Vitest).

Pendente para evolucao de produto:
- migrar 100% de todas as clausulas juridicas do Python para o motor novo de regras.
- geracao final DOCX/PDF com templates juridicos oficiais.
- trilha de auditoria detalhada por campo alterado no frontend.
