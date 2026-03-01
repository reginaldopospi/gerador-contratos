<script lang="ts">
  import { onMount } from "svelte";
  import { api, APIError } from "../lib/api";
  import type { PendingRegistration } from "../lib/types";
  import { requirePlatformAdmin } from "../lib/utils/guards";

  let loading = true;
  let savingUserID = "";
  let items: PendingRegistration[] = [];
  let pendingPasswords: Record<string, string> = {};
  let query = "";
  let error = "";
  let success = "";

  async function loadPending(): Promise<void> {
    loading = true;
    error = "";
    try {
      // Carrega os cadastros pendentes exibidos no painel administrativo.
      items = await api.listPendingRegistrations();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar cadastros pendentes";
    } finally {
      loading = false;
    }
  }

  async function approve(userID: string): Promise<void> {
    savingUserID = userID;
    error = "";
    success = "";
    try {
      const password = pendingPasswords[userID] ?? "";
      await api.approvePendingRegistration(userID, password);
      // Atualiza a grade para refletir a aprovacao sem precisar recarregar a pagina inteira.
      items = items.filter((item) => item.user_id !== userID);
      delete pendingPasswords[userID];
      success = "Cadastro aprovado com sucesso.";
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao aprovar cadastro";
    } finally {
      savingUserID = "";
    }
  }

  $: filteredItems = items.filter((item) => {
    const value = query.trim().toLowerCase();
    if (!value) {
      return true;
    }
    return (
      item.tenant_name.toLowerCase().includes(value) ||
      item.name.toLowerCase().includes(value) ||
      item.email.toLowerCase().includes(value)
    );
  });

  $: totalPending = items.length;
  $: totalFiltered = filteredItems.length;

  onMount(async () => {
    if (!requirePlatformAdmin()) {
      return;
    }
    await loadPending();
  });
</script>

<section class="section-stack">
  <div class="panel">
    <div class="page-head">
      <h1>Administracao da Plataforma</h1>
      <p>Area administrativa para governar o acesso de novos cadastrados.</p>
    </div>

    <div class="stat-grid">
      <div class="stat-card">
        <span class="stat-label">Cadastros pendentes</span>
        <strong class="stat-value">{totalPending}</strong>
      </div>
      <div class="stat-card">
        <span class="stat-label">Resultados no filtro</span>
        <strong class="stat-value">{totalFiltered}</strong>
      </div>
    </div>
  </div>

  <div class="panel">
    <div class="page-head">
      <h2>Aprovacao de Cadastros</h2>
      <p>Visualize os dados do cadastrante, ajuste a senha e aprove o acesso.</p>
    </div>

    <div class="grid cols-2">
      <div class="field">
        <label for="admin_search">Busca rapida</label>
        <input id="admin_search" bind:value={query} placeholder="Imobiliaria, nome ou email" />
      </div>
    </div>

    <div class="actions">
      <button class="btn ghost" on:click={loadPending} disabled={loading}>Atualizar</button>
    </div>

    {#if loading}
      <p>Carregando cadastros pendentes...</p>
    {:else}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Imobiliaria</th>
              <th>Nome</th>
              <th>Email</th>
              <th>Data</th>
              <th>Nova senha (opcional)</th>
              <th>Acao</th>
            </tr>
          </thead>
          <tbody>
            {#if filteredItems.length === 0}
              <tr>
                <td colspan="6">Nenhum cadastro pendente para o filtro informado.</td>
              </tr>
            {:else}
              {#each filteredItems as item}
                <tr>
                  <td>{item.tenant_name}</td>
                  <td>{item.name}</td>
                  <td>{item.email}</td>
                  <td>{new Date(item.created_at).toLocaleString("pt-BR")}</td>
                  <td>
                    <input
                      type="password"
                      placeholder="Manter senha atual"
                      bind:value={pendingPasswords[item.user_id]}
                    />
                  </td>
                  <td>
                    <button
                      class="btn primary"
                      on:click={() => approve(item.user_id)}
                      disabled={savingUserID === item.user_id}
                    >
                      {savingUserID === item.user_id ? "Aprovando..." : "Aprovar"}
                    </button>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    {/if}
  </div>

  {#if error}
    <div class="notice error">{error}</div>
  {/if}

  {#if success}
    <div class="notice success">{success}</div>
  {/if}
</section>
