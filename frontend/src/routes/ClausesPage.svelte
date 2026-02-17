<script lang="ts">
  import { onMount } from "svelte";
  import { api, APIError } from "../lib/api";
  import type { ClauseTemplate } from "../lib/types";
  import { requireAuth } from "../lib/utils/guards";

  let items: ClauseTemplate[] = [];
  let loading = false;
  let error = "";
  let success = "";
  let viewingClause: ClauseTemplate | null = null;
  let editingClauseID = "";

  let clause_key = "";
  let title = "";
  let content = "";
  let is_active = true;

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      items = await api.listClauses();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar clausulas";
    } finally {
      loading = false;
    }
  }

  async function saveClause(): Promise<void> {
    error = "";
    success = "";
    if (!clause_key.trim()) {
      error = "Informe a chave da clausula.";
      return;
    }
    if (!title.trim()) {
      error = "Informe o titulo da clausula.";
      return;
    }
    if (!content.trim()) {
      error = "Informe o conteudo da clausula.";
      return;
    }

    try {
      await api.upsertClause({ clause_key, title, content, is_active });
      success = editingClauseID ? "Clausula atualizada com sucesso." : "Clausula salva com sucesso.";
      clearForm();
      await load();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao salvar clausula";
    }
  }

  async function remove(clauseID: string): Promise<void> {
    error = "";
    success = "";
    try {
      await api.deleteClause(clauseID);
      success = "Clausula removida com sucesso.";
      if (viewingClause?.id === clauseID) {
        viewingClause = null;
      }
      if (editingClauseID === clauseID) {
        clearForm();
      }
      await load();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao remover clausula";
    }
  }

  function clearForm(): void {
    editingClauseID = "";
    clause_key = "";
    title = "";
    content = "";
    is_active = true;
  }

  function startEdit(item: ClauseTemplate): void {
    editingClauseID = item.id;
    clause_key = item.clause_key;
    title = item.title;
    content = item.content;
    is_active = item.is_active;
    error = "";
    success = `Editando clausula '${item.clause_key}'.`;
  }

  function cancelEdit(): void {
    clearForm();
    success = "Edicao cancelada.";
    error = "";
  }

  function viewClause(item: ClauseTemplate): void {
    viewingClause = item;
    error = "";
  }

  onMount(async () => {
    if (!requireAuth()) {
      return;
    }
    await load();
  });
</script>

<section class="section-stack">
  <div class="panel">
    <div class="page-head">
      <h1>Clausulas</h1>
      <p>Gestao de clausulas customizaveis por imobiliaria.</p>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Chave</th>
            <th>Titulo</th>
            <th>Ativa</th>
            <th>Acoes</th>
          </tr>
        </thead>
        <tbody>
          {#if loading}
            <tr><td colspan="4">Carregando...</td></tr>
          {:else if items.length === 0}
            <tr><td colspan="4">Nenhuma clausula cadastrada.</td></tr>
          {:else}
            {#each items as item}
              <tr>
                <td><code>{item.clause_key}</code></td>
                <td>{item.title}</td>
                <td>{item.is_active ? "Sim" : "Nao"}</td>
                <td>
                  <div class="clause-actions">
                    <button class="btn ghost btn-sm" on:click={() => viewClause(item)}>Visualizar</button>
                    <button class="btn ghost btn-sm" on:click={() => startEdit(item)}>Editar</button>
                    <button class="btn danger btn-sm" on:click={() => remove(item.id)}>Excluir</button>
                  </div>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    {#if viewingClause}
      <div class="clause-preview">
        <div class="clause-preview-head">
          <h3>Visualizar clausula</h3>
          <button class="btn ghost btn-sm" on:click={() => (viewingClause = null)}>Fechar</button>
        </div>
        <p><strong>Chave:</strong> <code>{viewingClause.clause_key}</code></p>
        <p><strong>Titulo:</strong> {viewingClause.title}</p>
        <p><strong>Status:</strong> {viewingClause.is_active ? "Ativa" : "Inativa"}</p>
        <pre class="clause-content-preview">{viewingClause.content}</pre>
      </div>
    {/if}

    {#if error}
      <div class="notice error">{error}</div>
    {/if}
    {#if success}
      <div class="notice success">{success}</div>
    {/if}
  </div>

  <div class="panel">
    <h2>{editingClauseID ? "Editar clausula" : "Nova clausula / atualizar clausula"}</h2>
    <div class="grid">
      <div class="field">
        <label for="clause_key">Chave</label>
        <input
          id="clause_key"
          bind:value={clause_key}
          placeholder="ex.: entrega_chaves_24h"
          disabled={Boolean(editingClauseID)}
        />
      </div>
      <div class="field">
        <label for="clause_title">Titulo</label>
        <input id="clause_title" bind:value={title} placeholder="Titulo da clausula" />
      </div>
      <div class="field">
        <label for="clause_content">Conteudo</label>
        <textarea id="clause_content" bind:value={content}></textarea>
      </div>
      <div class="field">
        <label class="checkbox-inline">
          <input type="checkbox" bind:checked={is_active} />
          Clausula ativa
        </label>
      </div>
    </div>

    <div class="actions">
      <button class="btn primary" on:click={saveClause}>
        {editingClauseID ? "Atualizar clausula" : "Salvar clausula"}
      </button>
      {#if editingClauseID}
        <button class="btn ghost" on:click={cancelEdit}>Cancelar edicao</button>
      {/if}
    </div>
    {#if editingClauseID}
      <p class="hint">Para trocar a chave, cancele a edicao e crie uma nova clausula.</p>
    {/if}
  </div>
</section>

<style>
  .clause-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .btn-sm {
    padding: 7px 10px;
    font-size: 0.84rem;
  }

  .clause-preview {
    margin-top: 12px;
    border: 1px solid #d4e2f4;
    border-radius: 12px;
    padding: 12px;
    background: rgba(255, 255, 255, 0.72);
  }

  .clause-preview-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .clause-preview-head h3 {
    margin: 0;
  }

  .clause-content-preview {
    margin: 0;
    border: 1px solid #d6e2f3;
    border-radius: 10px;
    padding: 10px;
    white-space: pre-wrap;
    background: #f8fbff;
    color: #1b2e4c;
    font-size: 0.9rem;
    line-height: 1.5;
  }
</style>
