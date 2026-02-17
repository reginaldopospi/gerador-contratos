# AGENTS.md

## Regras obrigatorias para todo desenvolvimento

1. Arquitetura e Design Patterns
- Sempre aplicar design patterns adequados ao problema.
- Padrao minimo esperado no backend: Repository + Service + Use Case.
- Evitar logica de negocio em controllers ou componentes de UI.
- Preferir composicao a heranca.

2. Clean Code
- Funcoes curtas e com responsabilidade unica.
- Nomes claros e orientados ao dominio.
- Evitar codigo duplicado; extrair modulos reutilizaveis.
- Tratar erros explicitamente com mensagens objetivas.

3. Banco de Dados
- Banco padrao do projeto: SQLite.
- Toda alteracao de schema deve usar migrations versionadas.
- Nao acessar banco sem camada de repositorio.
- Sempre definir constraints importantes (FK, UNIQUE, NOT NULL).

4. Testes
- Testes unitarios sao obrigatorios para toda feature nova e toda correcao.
- Nenhum codigo deve ser considerado concluido sem testes cobrindo regras de negocio.
- Mockar dependencias externas em testes unitarios.
- Manter cobertura de testes consistente no projeto.

5. Qualidade de entrega
- Nao enviar codigo quebrando build.
- Validar lint e testes antes de concluir tarefa.
- Documentar decisoes tecnicas relevantes quando houver trade-off.

6. Sempre coloque comentarios nos codigos que forem criados
