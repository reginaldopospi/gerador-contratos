<script lang="ts">
  import { onMount } from "svelte";
  import { Document, HeadingLevel, Packer, Paragraph } from "docx";
  import { api, APIError } from "../lib/api";
  import type { ClauseTemplate, ContractDetails, ContractPreview, ContractVersion } from "../lib/types";
  import {
    buildContractData,
    defaultPartyRef,
    DELIVERY_OPTIONS,
    draftFromContractData,
    emptyContractDraft,
    type ContractEditorDraft,
    type DraftStringField,
    type ExtraFieldType,
    type PartyRole
  } from "../lib/utils/contract-editor";
  import { requireAuth } from "../lib/utils/guards";

  export let params: { id: string };

  let details: ContractDetails | null = null;
  let preview: ContractPreview | null = null;
  let draft: ContractEditorDraft = emptyContractDraft();
  let availableClauses: ClauseTemplate[] = [];
  let clauseSearch = "";
  let clauseLoadError = "";
  let selectedVersion = 0;
  let loading = false;
  let saving = false;
  let previewing = false;
  let downloadingDocx = false;
  let error = "";
  let success = "";

  const propertyFields: Array<{ key: DraftStringField; label: string; placeholder: string }> = [
    { key: "imovelTipo", label: "Tipo do imovel", placeholder: "Ex.: Apartamento residencial" },
    { key: "imovelEndereco", label: "Endereco completo", placeholder: "Rua, numero, bairro, cidade/UF" },
    { key: "imovelMatricula", label: "Matricula", placeholder: "Ex.: 12345" },
    { key: "imovelCartorio", label: "Cartorio", placeholder: "Ex.: 1o Registro de Imoveis" }
  ];

  const paymentFields: Array<{ key: DraftStringField; label: string; placeholder: string }> = [
    { key: "precoTotal", label: "Preco total", placeholder: "R$ 450.000,00" },
    { key: "precoFinanciamento", label: "Financiamento", placeholder: "R$ 300.000,00" },
    { key: "precoFgts", label: "FGTS", placeholder: "R$ 20.000,00" },
    { key: "precoEntrada", label: "Entrada", placeholder: "R$ 80.000,00" },
    { key: "precoSinal", label: "Sinal", placeholder: "R$ 10.000,00" },
    { key: "precoRecursoProprio", label: "Recurso proprio", placeholder: "R$ 30.000,00" },
    { key: "precoCartaCredito", label: "Carta de credito", placeholder: "R$ 0,00" },
    { key: "precoSubsidio", label: "Subsidio", placeholder: "R$ 0,00" },
    { key: "precoParcelamentoTotal", label: "Parcelamento", placeholder: "R$ 15.000,00" },
    { key: "precoOutros", label: "Outros valores", placeholder: "Detalhe complementos" }
  ];

  const commissionFields: Array<{ key: DraftStringField; label: string; placeholder: string }> = [
    { key: "quemPagaComissao", label: "Quem paga comissao", placeholder: "Ex.: Vendedor" },
    { key: "valorComissao", label: "Valor da comissao", placeholder: "Ex.: 6% ou R$ 18.000,00" },
    { key: "momentoPagto", label: "Momento do pagamento", placeholder: "Ex.: Na assinatura" }
  ];

  $: {
    const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
    if (hydratedDraft !== draft) {
      draft = hydratedDraft;
    }
  }
  $: selectedClauseKeys = draft.clausulasSelecionadas.map((item) => item.clauseKey);
  $: clauseSuggestions = getClauseSuggestions(clauseSearch, availableClauses, selectedClauseKeys);

  async function load(): Promise<void> {
    loading = true;
    error = "";
    success = "";
    try {
      details = await api.getContract(params.id);
      try {
        availableClauses = await api.listClauses();
        clauseLoadError = "";
      } catch (clauseErr) {
        availableClauses = [];
        clauseLoadError =
          clauseErr instanceof APIError
            ? clauseErr.message
            : "Falha ao carregar clausulas para o seletor";
      }
      applyVersion(details.latest_version);
      await refreshPreview(false);
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar contrato";
    } finally {
      loading = false;
    }
  }

  function applyVersion(version?: ContractVersion): void {
    draft = hydrateDraftClauseMetadata(draftFromContractData(version?.data ?? {}), availableClauses);
    selectedVersion = version?.version_number ?? 0;
  }

  async function refreshPreview(showNotice: boolean): Promise<void> {
    if (!details) {
      return;
    }

    previewing = true;
    error = "";
    if (showNotice) {
      success = "";
    }

    try {
      const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
      if (hydratedDraft !== draft) {
        draft = hydratedDraft;
      }
      const data = buildContractData(hydratedDraft);
      preview = await api.previewContractFromData({
        numero: details.contract.numero,
        tipo: details.contract.tipo,
        data
      });
      if (showNotice) {
        success = "Previa atualizada com os dados atuais do formulario.";
      }
    } catch (err) {
      if (err instanceof APIError) {
        error = err.message;
      } else if (err instanceof Error) {
        error = err.message;
      } else {
        error = "Falha ao gerar previa";
      }
    } finally {
      previewing = false;
    }
  }

  async function saveNewVersion(): Promise<void> {
    saving = true;
    error = "";
    success = "";

    try {
      const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
      if (hydratedDraft !== draft) {
        draft = hydratedDraft;
      }
      const data = buildContractData(hydratedDraft);
      await api.addContractVersion(params.id, data);
      await load();
      success = "Nova versao salva com sucesso.";
    } catch (err) {
      if (err instanceof APIError) {
        error = err.message;
      } else if (err instanceof Error) {
        error = err.message;
      } else {
        error = "Falha ao salvar versao";
      }
    } finally {
      saving = false;
    }
  }

  async function downloadContractDocx(): Promise<void> {
    if (!details) {
      return;
    }

    downloadingDocx = true;
    error = "";
    success = "";

    try {
      const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
      if (hydratedDraft !== draft) {
        draft = hydratedDraft;
      }

      const data = buildContractData(hydratedDraft);
      const latestPreview = await api.previewContractFromData({
        numero: details.contract.numero,
        tipo: details.contract.tipo,
        data
      });
      preview = latestPreview;

      const contractText = buildContractText(latestPreview);
      if (contractText.trim() === "") {
        throw new Error("Nao foi possivel montar o contrato para exportacao.");
      }

      const documentBody = new Document({
        sections: [
          {
            children: buildContractParagraphs(latestPreview.title, contractText)
          }
        ]
      });

      const blob = await Packer.toBlob(documentBody);
      triggerBlobDownload(blob, buildDocxFilename(details.contract.numero, details.contract.tipo));
      success = "Contrato DOCX baixado com sucesso.";
    } catch (err) {
      if (err instanceof APIError) {
        error = err.message;
      } else if (err instanceof Error) {
        error = err.message;
      } else {
        error = "Falha ao baixar contrato em DOCX";
      }
    } finally {
      downloadingDocx = false;
    }
  }

  onMount(async () => {
    if (!requireAuth()) {
      return;
    }
    await load();
  });

  function restoreLatestVersion(): void {
    if (!details?.latest_version) {
      draft = emptyContractDraft();
      selectedVersion = 0;
      preview = null;
      success = "Sem versao salva ainda. Formulario limpo.";
      error = "";
      return;
    }

    applyVersion(details.latest_version);
    success = `Versao mais recente v${details.latest_version.version_number} restaurada.`;
    error = "";
    void refreshPreview(false);
  }

  function loadVersion(version: ContractVersion): void {
    applyVersion(version);
    success = `Versao v${version.version_number} carregada para edicao.`;
    error = "";
    void refreshPreview(false);
  }

  function updateField(key: DraftStringField, value: string): void {
    draft = { ...draft, [key]: value } as ContractEditorDraft;
  }

  function addParty(role: PartyRole): void {
    const list = role === "vendedores" ? draft.vendedores : draft.compradores;
    const next = [...list, { ref: defaultPartyRef(role, list.length + 1), nome: "", razaoSocial: "" }];
    draft = { ...draft, [role]: next } as ContractEditorDraft;
  }

  function updateParty(role: PartyRole, index: number, key: "nome" | "razaoSocial", value: string): void {
    const list = role === "vendedores" ? draft.vendedores : draft.compradores;
    const next = list.map((item, i) => (i === index ? { ...item, [key]: value } : item));
    draft = { ...draft, [role]: next } as ContractEditorDraft;
  }

  function removeParty(role: PartyRole, index: number): void {
    const list = role === "vendedores" ? draft.vendedores : draft.compradores;
    const filtered = list.filter((_, i) => i !== index);
    const next =
      filtered.length > 0 ? filtered : [{ ref: defaultPartyRef(role, 1), nome: "", razaoSocial: "" }];
    draft = { ...draft, [role]: next } as ContractEditorDraft;
  }

  function addDeliveryClause(): void {
    draft = {
      ...draft,
      clausulasEntregaChaves: [...draft.clausulasEntregaChaves, { key: "", text: "" }]
    };
  }

  function updateDeliveryClause(index: number, key: "key" | "text", value: string): void {
    const next = draft.clausulasEntregaChaves.map((item, i) =>
      i === index ? { ...item, [key]: value } : item
    );
    draft = { ...draft, clausulasEntregaChaves: next };
  }

  function removeDeliveryClause(index: number): void {
    const next = draft.clausulasEntregaChaves.filter((_, i) => i !== index);
    draft = { ...draft, clausulasEntregaChaves: next };
  }

  function addExtraField(): void {
    draft = {
      ...draft,
      extras: [...draft.extras, { key: "", type: "text", value: "" }]
    };
  }

  function updateExtraKey(index: number, value: string): void {
    const next = draft.extras.map((item, i) => (i === index ? { ...item, key: value } : item));
    draft = { ...draft, extras: next };
  }

  function updateExtraType(index: number, value: ExtraFieldType): void {
    const next = draft.extras.map((item, i) => {
      if (i !== index) {
        return item;
      }
      if (value === "boolean") {
        return { ...item, type: value, value: item.value === "false" ? "false" : "true" };
      }
      if (value === "json" && item.value.trim() === "") {
        return { ...item, type: value, value: "{}" };
      }
      return { ...item, type: value };
    });
    draft = { ...draft, extras: next };
  }

  function updateExtraValue(index: number, value: string): void {
    const next = draft.extras.map((item, i) => (i === index ? { ...item, value } : item));
    draft = { ...draft, extras: next };
  }

  function removeExtraField(index: number): void {
    const next = draft.extras.filter((_, i) => i !== index);
    draft = { ...draft, extras: next };
  }

  function addClauseTag(clauseKey: string): void {
    const key = clauseKey.trim();
    if (key === "") {
      return;
    }
    if (draft.clausulasSelecionadas.some((item) => item.clauseKey === key)) {
      clauseSearch = "";
      return;
    }

    const match = availableClauses.find((item) => item.clause_key === key);

    draft = {
      ...draft,
      clausulasSelecionadas: [
        ...draft.clausulasSelecionadas,
        {
          clauseKey: key,
          title: match?.title ?? key,
          content: match?.content ?? "",
          index: ""
        }
      ]
    };
    clauseSearch = "";
  }

  function removeClauseTag(clauseKey: string): void {
    draft = {
      ...draft,
      clausulasSelecionadas: draft.clausulasSelecionadas.filter((item) => item.clauseKey !== clauseKey)
    };
  }

  function updateClauseTagIndex(clauseKey: string, value: string): void {
    draft = {
      ...draft,
      clausulasSelecionadas: draft.clausulasSelecionadas.map((item) =>
        item.clauseKey === clauseKey ? { ...item, index: value } : item
      )
    };
  }

  function addCustomClause(): void {
    draft = {
      ...draft,
      clausulasCustomizadas: [
        ...draft.clausulasCustomizadas,
        { title: "", content: "", index: "" }
      ]
    };
  }

  function updateCustomClause(index: number, key: "title" | "content" | "index", value: string): void {
    draft = {
      ...draft,
      clausulasCustomizadas: draft.clausulasCustomizadas.map((item, itemIndex) =>
        itemIndex === index ? { ...item, [key]: value } : item
      )
    };
  }

  function removeCustomClause(index: number): void {
    draft = {
      ...draft,
      clausulasCustomizadas: draft.clausulasCustomizadas.filter((_, itemIndex) => itemIndex !== index)
    };
  }

  function handleClauseSearchInput(event: Event): void {
    clauseSearch = inputValue(event);
  }

  function handleClauseSearchKeydown(event: KeyboardEvent): void {
    if (event.key !== "Enter") {
      return;
    }
    event.preventDefault();
    if (clauseSuggestions.length > 0) {
      addClauseTag(clauseSuggestions[0].clause_key);
    }
  }

  function getClauseSuggestions(
    query: string,
    clauses: ClauseTemplate[],
    selectedKeys: string[]
  ): ClauseTemplate[] {
    const active = clauses.filter((item) => item.is_active);
    const normalized = query.trim().toLowerCase();

    return active
      .filter((item) => !selectedKeys.includes(item.clause_key))
      .filter((item) => {
        if (normalized === "") {
          return true;
        }
        return (
          item.clause_key.toLowerCase().includes(normalized) ||
          item.title.toLowerCase().includes(normalized) ||
          item.content.toLowerCase().includes(normalized)
        );
      })
      .sort((a, b) => a.title.localeCompare(b.title))
      .slice(0, 8);
  }

  function inputValue(event: Event): string {
    return (event.currentTarget as HTMLInputElement).value;
  }

  function selectValue(event: Event): string {
    return (event.currentTarget as HTMLSelectElement).value;
  }

  function textareaValue(event: Event): string {
    return (event.currentTarget as HTMLTextAreaElement).value;
  }

  function buildContractText(contractPreview: ContractPreview): string {
    const fullText = (contractPreview.full_text ?? "").trim();
    if (fullText !== "") {
      return fullText;
    }
    return contractPreview.sections.join("\n\n").trim();
  }

  function buildContractParagraphs(title: string, text: string): Paragraph[] {
    const paragraphs: Paragraph[] = [];
    const normalizedTitle = title.trim();
    if (normalizedTitle !== "") {
      paragraphs.push(
        new Paragraph({
          heading: HeadingLevel.HEADING_1,
          text: normalizedTitle
        })
      );
      paragraphs.push(new Paragraph({ text: " " }));
    }

    for (const line of text.split(/\r?\n/)) {
      paragraphs.push(
        new Paragraph({
          text: line === "" ? " " : line
        })
      );
    }

    return paragraphs;
  }

  function triggerBlobDownload(blob: Blob, fileName: string): void {
    const blobUrl = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = blobUrl;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(blobUrl);
  }

  function buildDocxFilename(numero: string, tipo: string): string {
    const numberPart = sanitizeFilenamePart(numero) || "sem-numero";
    const typePart = sanitizeFilenamePart(tipo) || "contrato";
    return `contrato-${numberPart}-${typePart}.docx`;
  }

  function sanitizeFilenamePart(value: string): string {
    return value
      .trim()
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "")
      .slice(0, 60);
  }

  function hydrateDraftClauseMetadata(
    source: ContractEditorDraft,
    clauses: ClauseTemplate[]
  ): ContractEditorDraft {
    if (source.clausulasSelecionadas.length === 0 || clauses.length === 0) {
      return source;
    }

    const clausesByKey = new Map(clauses.map((item) => [item.clause_key, item]));
    let hasChanges = false;
    const selected = source.clausulasSelecionadas.map((item) => {
      const template = clausesByKey.get(item.clauseKey);
      if (!template) {
        return item;
      }

      const nextTitle = item.title.trim() === "" ? template.title : item.title;
      const nextContent = item.content.trim() === "" ? template.content : item.content;
      if (nextTitle === item.title && nextContent === item.content) {
        return item;
      }

      hasChanges = true;
      return { ...item, title: nextTitle, content: nextContent };
    });

    if (!hasChanges) {
      return source;
    }
    return { ...source, clausulasSelecionadas: selected };
  }
</script>

<section class="section-stack">
  <div class="panel">
    <div class="page-head">
      <h1>Editor Visual do Contrato</h1>
      <p>Edite os dados em blocos simples, sem precisar manipular JSON.</p>
    </div>

    {#if loading}
      <p>Carregando...</p>
    {:else if details}
      <div class="editor-meta">
        <article class="meta-card">
          <span class="meta-label">Numero</span>
          <strong>{details.contract.numero}</strong>
        </article>
        <article class="meta-card">
          <span class="meta-label">Tipo</span>
          <strong>{details.contract.tipo}</strong>
        </article>
        <article class="meta-card">
          <span class="meta-label">Status</span>
          <strong>{details.contract.status}</strong>
        </article>
        <article class="meta-card">
          <span class="meta-label">Versao em edicao</span>
          <strong>{selectedVersion > 0 ? `v${selectedVersion}` : "Rascunho local"}</strong>
        </article>
      </div>

      <div class="editor-toolbar">
        <button class="btn ghost" disabled={previewing} on:click={() => refreshPreview(true)}>
          {previewing ? "Gerando previa..." : "Atualizar previa"}
        </button>
        <button class="btn ghost" on:click={restoreLatestVersion}>Restaurar ultima versao</button>
        <button class="btn ghost" on:click={load}>Recarregar contrato</button>
      </div>

      <div class="editor-form">
        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Partes do contrato</h3>
            <p>Adicione vendedores/cedentes e compradores/cessionarios.</p>
          </div>

          <div class="party-columns">
            <div class="party-column">
              <div class="party-column-head">
                <h4>Parte vendedora / cedente</h4>
                <button class="btn ghost btn-sm" on:click={() => addParty("vendedores")}>
                  Adicionar
                </button>
              </div>

              {#each draft.vendedores as party, index}
                <div class="party-card">
                  <div class="grid cols-2">
                    <div class="field">
                      <label for={`seller_nome_${index}`}>Nome</label>
                      <input
                        id={`seller_nome_${index}`}
                        value={party.nome}
                        placeholder="Nome completo"
                        on:input={(event) => updateParty("vendedores", index, "nome", inputValue(event))}
                      />
                    </div>
                    <div class="field">
                      <label for={`seller_razao_${index}`}>Razao social</label>
                      <input
                        id={`seller_razao_${index}`}
                        value={party.razaoSocial}
                        placeholder="Opcional para pessoa juridica"
                        on:input={(event) =>
                          updateParty("vendedores", index, "razaoSocial", inputValue(event))}
                      />
                    </div>
                  </div>

                  <div class="inline-actions">
                    <button
                      class="btn ghost btn-sm"
                      disabled={draft.vendedores.length === 1}
                      on:click={() => removeParty("vendedores", index)}
                    >
                      Remover
                    </button>
                  </div>
                </div>
              {/each}
            </div>

            <div class="party-column">
              <div class="party-column-head">
                <h4>Parte compradora / cessionaria</h4>
                <button class="btn ghost btn-sm" on:click={() => addParty("compradores")}>
                  Adicionar
                </button>
              </div>

              {#each draft.compradores as party, index}
                <div class="party-card">
                  <div class="grid cols-2">
                    <div class="field">
                      <label for={`buyer_nome_${index}`}>Nome</label>
                      <input
                        id={`buyer_nome_${index}`}
                        value={party.nome}
                        placeholder="Nome completo"
                        on:input={(event) => updateParty("compradores", index, "nome", inputValue(event))}
                      />
                    </div>
                    <div class="field">
                      <label for={`buyer_razao_${index}`}>Razao social</label>
                      <input
                        id={`buyer_razao_${index}`}
                        value={party.razaoSocial}
                        placeholder="Opcional para pessoa juridica"
                        on:input={(event) =>
                          updateParty("compradores", index, "razaoSocial", inputValue(event))}
                      />
                    </div>
                  </div>

                  <div class="inline-actions">
                    <button
                      class="btn ghost btn-sm"
                      disabled={draft.compradores.length === 1}
                      on:click={() => removeParty("compradores", index)}
                    >
                      Remover
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Imovel</h3>
            <p>Dados essenciais do objeto contratual.</p>
          </div>

          <div class="grid cols-2">
            {#each propertyFields as field}
              <div class="field">
                <label for={`property_${field.key}`}>{field.label}</label>
                <input
                  id={`property_${field.key}`}
                  value={draft[field.key]}
                  placeholder={field.placeholder}
                  on:input={(event) => updateField(field.key, inputValue(event))}
                />
              </div>
            {/each}
          </div>
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Pagamento</h3>
            <p>Preencha somente os valores que se aplicam ao contrato.</p>
          </div>

          <div class="grid cols-3">
            {#each paymentFields as field}
              <div class="field">
                <label for={`payment_${field.key}`}>{field.label}</label>
                <input
                  id={`payment_${field.key}`}
                  value={draft[field.key]}
                  placeholder={field.placeholder}
                  on:input={(event) => updateField(field.key, inputValue(event))}
                />
              </div>
            {/each}
          </div>
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Comissao e intermediacao</h3>
          </div>

          <div class="grid cols-3">
            {#each commissionFields as field}
              <div class="field">
                <label for={`commission_${field.key}`}>{field.label}</label>
                <input
                  id={`commission_${field.key}`}
                  value={draft[field.key]}
                  placeholder={field.placeholder}
                  on:input={(event) => updateField(field.key, inputValue(event))}
                />
              </div>
            {/each}
          </div>
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Clausulas do contrato</h3>
            <p>Busque clausulas salvas e adicione em formato de tags.</p>
          </div>

          <div class="field">
            <label for="clause_search">Adicionar clausula</label>
            <input
              id="clause_search"
              value={clauseSearch}
              placeholder="Digite titulo, chave ou trecho da clausula"
              on:input={handleClauseSearchInput}
              on:keydown={handleClauseSearchKeydown}
            />
          </div>

          {#if clauseLoadError}
            <div class="notice info">{clauseLoadError}</div>
          {/if}

          {#if clauseSearch.trim() !== ""}
            <div class="clause-suggestion-list">
              {#if clauseSuggestions.length === 0}
                <p class="hint">Nenhuma clausula ativa encontrada para "{clauseSearch}".</p>
              {:else}
                {#each clauseSuggestions as clause}
                  <button
                    class="clause-suggestion-item"
                    type="button"
                    on:click={() => addClauseTag(clause.clause_key)}
                  >
                    <strong>{clause.title}</strong>
                    <small>{clause.clause_key}</small>
                  </button>
                {/each}
              {/if}
            </div>
          {/if}

          <div class="clause-tags-area">
            {#if draft.clausulasSelecionadas.length === 0}
              <p class="hint">Nenhuma clausula selecionada.</p>
            {:else}
              <div class="linked-clause-list">
                {#each draft.clausulasSelecionadas as clause}
                  <div class="linked-clause-card">
                    <div class="linked-clause-title">
                      <strong>{clause.title || clause.clauseKey}</strong>
                      <small>{clause.clauseKey}</small>
                    </div>
                    <div class="field">
                      <label for={`clause_index_${clause.clauseKey}`}>Indice no contrato</label>
                      <input
                        id={`clause_index_${clause.clauseKey}`}
                        value={clause.index}
                        placeholder="Ex.: 1.1.2"
                        on:input={(event) => updateClauseTagIndex(clause.clauseKey, inputValue(event))}
                      />
                    </div>
                    <div class="inline-actions">
                      <button
                        type="button"
                        class="btn ghost btn-sm"
                        aria-label={`Remover clausula ${clause.clauseKey}`}
                        on:click={() => removeClauseTag(clause.clauseKey)}
                      >
                        Remover
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

          <div class="editor-subsection">
            <div class="party-column-head">
              <h4>Clausulas customizadas vinculadas</h4>
              <button class="btn ghost btn-sm" on:click={addCustomClause}>Adicionar clausula customizada</button>
            </div>

            {#if draft.clausulasCustomizadas.length === 0}
              <p class="hint">Nenhuma clausula customizada vinculada.</p>
            {:else}
              <div class="delivery-clause-list">
                {#each draft.clausulasCustomizadas as customClause, customIndex}
                  <div class="delivery-clause-card">
                    <div class="grid cols-2">
                      <div class="field">
                        <label for={`custom_clause_index_${customIndex}`}>Indice no contrato</label>
                        <input
                          id={`custom_clause_index_${customIndex}`}
                          value={customClause.index}
                          placeholder="Ex.: 1.1.2"
                          on:input={(event) =>
                            updateCustomClause(customIndex, "index", inputValue(event))}
                        />
                      </div>
                      <div class="field">
                        <label for={`custom_clause_title_${customIndex}`}>Titulo da clausula</label>
                        <input
                          id={`custom_clause_title_${customIndex}`}
                          value={customClause.title}
                          placeholder="Ex.: Clausula X"
                          on:input={(event) =>
                            updateCustomClause(customIndex, "title", inputValue(event))}
                        />
                      </div>
                      <div class="field span-all">
                        <label for={`custom_clause_content_${customIndex}`}>Conteudo da clausula</label>
                        <textarea
                          id={`custom_clause_content_${customIndex}`}
                          value={customClause.content}
                          placeholder="Texto completo da clausula customizada."
                          on:input={(event) =>
                            updateCustomClause(customIndex, "content", textareaValue(event))}
                        ></textarea>
                      </div>
                    </div>

                    <div class="inline-actions">
                      <button class="btn ghost btn-sm" on:click={() => removeCustomClause(customIndex)}>
                        Remover
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Entrega de chaves</h3>
            <p>Selecione uma regra padrao ou escreva o texto personalizado.</p>
          </div>

          <div class="grid cols-2">
            <div class="field">
              <label for="entrega_chaves">Regra principal</label>
              <select
                id="entrega_chaves"
                value={draft.entregaChaves}
                on:change={(event) => updateField("entregaChaves", selectValue(event))}
              >
                <option value="">Selecione uma opcao</option>
                {#each DELIVERY_OPTIONS as option}
                  <option value={option}>{option}</option>
                {/each}
              </select>
            </div>
          </div>

          {#if draft.entregaChaves === "Escrever no contrato"}
            <div class="field">
              <label for="entrega_chaves_texto">Texto da clausula</label>
              <textarea
                id="entrega_chaves_texto"
                value={draft.entregaChavesTexto}
                placeholder="Descreva como e quando as chaves serao entregues."
                on:input={(event) => updateField("entregaChavesTexto", textareaValue(event))}
              ></textarea>
            </div>
          {/if}

          <div class="editor-subsection">
            <div class="party-column-head">
              <h4>Clausulas personalizadas de entrega</h4>
              <button class="btn ghost btn-sm" on:click={addDeliveryClause}>Adicionar clausula</button>
            </div>

            {#if draft.clausulasEntregaChaves.length === 0}
              <p class="hint">Nenhuma clausula personalizada cadastrada.</p>
            {:else}
              <div class="delivery-clause-list">
                {#each draft.clausulasEntregaChaves as clause, index}
                  <div class="delivery-clause-card">
                    <div class="grid cols-2">
                      <div class="field">
                        <label for={`delivery_key_${index}`}>Nome da opcao</label>
                        <input
                          id={`delivery_key_${index}`}
                          value={clause.key}
                          placeholder="Ex.: Na vistoria final"
                          on:input={(event) =>
                            updateDeliveryClause(index, "key", inputValue(event))}
                        />
                      </div>
                      <div class="field span-all">
                        <label for={`delivery_text_${index}`}>Texto juridico</label>
                        <textarea
                          id={`delivery_text_${index}`}
                          value={clause.text}
                          placeholder="Texto que sera utilizado quando esta opcao for escolhida."
                          on:input={(event) =>
                            updateDeliveryClause(index, "text", textareaValue(event))}
                        ></textarea>
                      </div>
                    </div>

                    <div class="inline-actions">
                      <button class="btn ghost btn-sm" on:click={() => removeDeliveryClause(index)}>
                        Remover
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Campos adicionais</h3>
            <p>Use para qualquer dado extra que nao esteja no formulario principal.</p>
          </div>

          <div class="actions">
            <button class="btn ghost" on:click={addExtraField}>Adicionar campo</button>
          </div>

          <div class="table-wrap">
            <table class="extras-table">
              <thead>
                <tr>
                  <th>Chave</th>
                  <th>Tipo</th>
                  <th>Valor</th>
                  <th>Acoes</th>
                </tr>
              </thead>
              <tbody>
                {#if draft.extras.length === 0}
                  <tr><td colspan="4">Sem campos adicionais.</td></tr>
                {:else}
                  {#each draft.extras as extra, index}
                    <tr>
                      <td>
                        <input
                          value={extra.key}
                          placeholder="exemplo__chave"
                          on:input={(event) => updateExtraKey(index, inputValue(event))}
                        />
                      </td>
                      <td>
                        <select
                          value={extra.type}
                          on:change={(event) =>
                            updateExtraType(index, selectValue(event) as ExtraFieldType)}
                        >
                          <option value="text">Texto</option>
                          <option value="number">Numero</option>
                          <option value="boolean">Booleano</option>
                          <option value="json">JSON</option>
                        </select>
                      </td>
                      <td class="value-cell">
                        {#if extra.type === "boolean"}
                          <select
                            value={extra.value === "false" ? "false" : "true"}
                            on:change={(event) => updateExtraValue(index, selectValue(event))}
                          >
                            <option value="true">true</option>
                            <option value="false">false</option>
                          </select>
                        {:else if extra.type === "json"}
                          <textarea
                            class="mini-json"
                            value={extra.value}
                            on:input={(event) => updateExtraValue(index, textareaValue(event))}
                          ></textarea>
                        {:else}
                          <input
                            value={extra.value}
                            on:input={(event) => updateExtraValue(index, inputValue(event))}
                          />
                        {/if}
                      </td>
                      <td>
                        <button class="btn danger btn-sm" on:click={() => removeExtraField(index)}>
                          Excluir
                        </button>
                      </td>
                    </tr>
                  {/each}
                {/if}
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <div class="actions">
        <button class="btn primary" disabled={saving} on:click={saveNewVersion}>
          {saving ? "Salvando..." : "Salvar nova versao"}
        </button>
        <button class="btn ghost" disabled={previewing} on:click={() => refreshPreview(true)}>
          {previewing ? "Gerando previa..." : "Gerar previa agora"}
        </button>
        <button
          class="btn ghost"
          disabled={downloadingDocx || previewing}
          on:click={downloadContractDocx}
        >
          {downloadingDocx ? "Gerando DOCX..." : "Baixar DOCX"}
        </button>
      </div>

      {#if success}
        <div class="notice success">{success}</div>
      {/if}
      {#if error}
        <div class="notice error">{error}</div>
      {/if}
    {/if}
  </div>

  <div class="panel">
    <h2>Previa juridica automatica</h2>
    {#if previewing}
      <p>Atualizando previa...</p>
    {/if}
    {#if preview}
      <h3>{preview.title}</h3>
      {#if preview.full_text}
        <article class="full-contract-preview">{preview.full_text}</article>
      {:else}
        <ol class="list-tight">
          {#each preview.sections as section}
            <li>{section}</li>
          {/each}
        </ol>
      {/if}
    {:else}
      <p>Sem previa disponivel.</p>
    {/if}
  </div>

  <div class="panel">
    <h2>Historico de versoes</h2>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Versao</th>
            <th>Criada em</th>
            <th>Acoes</th>
          </tr>
        </thead>
        <tbody>
          {#if details && details.versions.length > 0}
            {#each details.versions as version}
              <tr class:active-version={version.version_number === selectedVersion}>
                <td>v{version.version_number}</td>
                <td>{new Date(version.created_at).toLocaleString()}</td>
                <td>
                  <button class="btn ghost btn-sm" on:click={() => loadVersion(version)}>
                    Carregar
                  </button>
                </td>
              </tr>
            {/each}
          {:else}
            <tr><td colspan="3">Nenhuma versao</td></tr>
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</section>

<style>
  .editor-meta {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 10px;
    margin-bottom: 12px;
  }

  .meta-card {
    border: 1px solid #d3e0f0;
    border-radius: 12px;
    background: rgba(245, 250, 255, 0.9);
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .meta-label {
    color: #445f81;
    font-size: 0.8rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .editor-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 14px;
  }

  .editor-form {
    display: grid;
    gap: 12px;
  }

  .editor-section {
    border: 1px solid #d4e2f4;
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.55);
    padding: 14px;
    display: grid;
    gap: 12px;
  }

  .editor-section-head {
    display: grid;
    gap: 4px;
  }

  .editor-section-head h3 {
    margin: 0;
  }

  .editor-section-head p {
    margin: 0;
    color: #4a6281;
  }

  .party-columns {
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  }

  .party-column {
    display: grid;
    gap: 10px;
  }

  .party-column-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .party-column-head h4 {
    margin: 0;
    font-size: 1rem;
  }

  .party-card {
    border: 1px solid #d7e5f6;
    border-radius: 12px;
    background: #f9fbff;
    padding: 10px;
  }

  .inline-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 8px;
  }

  .clause-suggestion-list {
    border: 1px solid #d8e5f6;
    border-radius: 12px;
    background: #f8fbff;
    display: grid;
    gap: 6px;
    padding: 8px;
    max-height: 260px;
    overflow-y: auto;
  }

  .clause-suggestion-item {
    border: 1px solid #d0dff2;
    border-radius: 10px;
    padding: 8px 10px;
    background: #fff;
    color: inherit;
    text-align: left;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .clause-suggestion-item:hover {
    border-color: #22a18f;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
  }

  .clause-suggestion-item small {
    color: #4f6686;
    font-size: 0.8rem;
  }

  .clause-tags-area {
    border-top: 1px solid #dce8f6;
    padding-top: 10px;
  }

  .linked-clause-list {
    display: grid;
    gap: 8px;
  }

  .linked-clause-card {
    border: 1px solid #d7e5f6;
    border-radius: 12px;
    background: #f8fbff;
    padding: 10px;
    display: grid;
    gap: 8px;
  }

  .linked-clause-title {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .linked-clause-title small {
    color: #4f6686;
    font-size: 0.8rem;
  }

  .linked-clause-card .inline-actions {
    margin-top: 0;
  }

  .editor-subsection {
    border-top: 1px solid #dce8f6;
    padding-top: 12px;
    display: grid;
    gap: 10px;
  }

  .delivery-clause-list {
    display: grid;
    gap: 10px;
  }

  .delivery-clause-card {
    border: 1px solid #d9e6f6;
    border-radius: 12px;
    background: #f8fbff;
    padding: 10px;
  }

  .hint {
    margin: 0;
    color: #5a6f8c;
    font-size: 0.92rem;
  }

  .extras-table th,
  .extras-table td {
    min-width: 140px;
  }

  .extras-table .value-cell {
    min-width: 280px;
  }

  .mini-json {
    min-height: 88px;
    font-family: "Cascadia Code", "Consolas", monospace;
    font-size: 0.86rem;
  }

  .full-contract-preview {
    margin-top: 10px;
    border: 1px solid #d5e3f5;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.92);
    padding: 14px;
    white-space: pre-wrap;
    line-height: 1.6;
    font-size: 0.95rem;
  }

  .btn-sm {
    padding: 7px 10px;
    font-size: 0.84rem;
  }

  .active-version {
    background: rgba(15, 118, 110, 0.12);
  }

  @media (max-width: 760px) {
    .editor-section {
      padding: 12px;
    }

    .party-column-head {
      flex-wrap: wrap;
    }

    .extras-table .value-cell {
      min-width: 220px;
    }
  }
</style>
