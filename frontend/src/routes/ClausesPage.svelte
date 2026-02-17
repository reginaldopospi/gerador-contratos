<script lang="ts">
  import { onMount } from "svelte";
  import { api, APIError } from "../lib/api";
  import type { ClauseTemplate } from "../lib/types";
  import { requireAuth } from "../lib/utils/guards";

  let items: ClauseTemplate[] = [];
  let loading = false;
  let error = "";
  let success = "";

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
    try {
      await api.upsertClause({ clause_key, title, content, is_active });
      success = "Clausula salva com sucesso.";
      clause_key = "";
      title = "";
      content = "";
      is_active = true;
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
      await load();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao remover clausula";
    }
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
                  <button class="btn danger" on:click={() => remove(item.id)}>Excluir</button>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    {#if error}
      <div class="notice error">{error}</div>
    {/if}
    {#if success}
      <div class="notice success">{success}</div>
    {/if}
  </div>

  <div class="panel">
    <h2>Nova clausula / atualizar clausula</h2>
    <div class="grid">
      <div class="field">
        <label for="clause_key">Chave</label>
        <input id="clause_key" bind:value={clause_key} placeholder="ex.: entrega_chaves_24h" />
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
      <button class="btn primary" on:click={saveClause}>Salvar clausula</button>
    </div>
  </div>
</section>
