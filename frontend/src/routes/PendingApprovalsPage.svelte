<script lang="ts">
  import { onMount } from "svelte";
  import { api, APIError } from "../lib/api";
  import type { PendingRegistration, TenantSummary } from "../lib/types";
  import {
    filterPendingRegistrations,
    filterTenantSummaries
  } from "../lib/utils/admin-panel";
  import { requirePlatformAdmin } from "../lib/utils/guards";

  let loading = true;
  let savingUserID = "";
  let savingTenantID = "";
  let deletingTenantID = "";
  let resettingPasswordTenantID = "";
  let editingTenantID = "";
  let items: PendingRegistration[] = [];
  let tenants: TenantSummary[] = [];
  let pendingPasswords: Record<string, string> = {};
  let tenantEditName: Record<string, string> = {};
  let tenantEditCNPJ: Record<string, string> = {};
  let tenantResetPasswords: Record<string, string> = {};
  let query = "";
  let error = "";
  let success = "";

  async function loadAdminData(): Promise<void> {
    loading = true;
    error = "";
    try {
      // Carrega os dados principais do painel com uma unica sincronizacao.
      const [pendingItems, tenantItems] = await Promise.all([
        api.listPendingRegistrations(),
        api.listTenants()
      ]);
      items = pendingItems;
      tenants = tenantItems;
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar dados administrativos";
    } finally {
      loading = false;
    }
  }

  function beginTenantEdit(tenant: TenantSummary): void {
    // Preenche os campos editaveis para manter o rascunho isolado por tenant.
    editingTenantID = tenant.tenant_id;
    tenantEditName[tenant.tenant_id] = tenant.tenant_name;
    tenantEditCNPJ[tenant.tenant_id] = tenant.tenant_cnpj ?? "";
  }

  function cancelTenantEdit(tenantID: string): void {
    editingTenantID = "";
    delete tenantEditName[tenantID];
    delete tenantEditCNPJ[tenantID];
  }

  async function saveTenant(tenantID: string): Promise<void> {
    savingTenantID = tenantID;
    error = "";
    success = "";
    try {
      const tenantName = (tenantEditName[tenantID] ?? "").trim();
      const tenantCNPJ = (tenantEditCNPJ[tenantID] ?? "").trim();
      await api.updateTenant(tenantID, { tenant_name: tenantName, tenant_cnpj: tenantCNPJ });
      // Atualiza a grade local sem obrigar recarga completa.
      tenants = tenants.map((tenant) =>
        tenant.tenant_id === tenantID
          ? { ...tenant, tenant_name: tenantName, tenant_cnpj: tenantCNPJ }
          : tenant
      );
      // Sincroniza o nome tambem na grade de pendencias.
      items = items.map((item) =>
        item.tenant_id === tenantID ? { ...item, tenant_name: tenantName } : item
      );
      cancelTenantEdit(tenantID);
      success = "Imobiliaria atualizada com sucesso.";
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao atualizar imobiliaria";
    } finally {
      savingTenantID = "";
    }
  }

  async function removeTenant(tenant: TenantSummary): Promise<void> {
    if (
      !window.confirm(
        `Confirma excluir a imobiliaria "${tenant.tenant_name}"? Esta acao remove usuarios e contratos vinculados.`
      )
    ) {
      return;
    }

    deletingTenantID = tenant.tenant_id;
    error = "";
    success = "";
    try {
      await api.deleteTenant(tenant.tenant_id);
      // Remove a imobiliaria da grade principal e quaisquer pendencias do mesmo tenant.
      tenants = tenants.filter((item) => item.tenant_id !== tenant.tenant_id);
      items = items.filter((item) => item.tenant_id !== tenant.tenant_id);
      cancelTenantEdit(tenant.tenant_id);
      delete tenantResetPasswords[tenant.tenant_id];
      success = "Imobiliaria excluida com sucesso.";
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao excluir imobiliaria";
    } finally {
      deletingTenantID = "";
    }
  }

  async function resetTenantAdminPassword(tenant: TenantSummary): Promise<void> {
    const newPassword = (tenantResetPasswords[tenant.tenant_id] ?? "").trim();
    if (!newPassword) {
      error = "Informe a nova senha do administrador da imobiliaria.";
      success = "";
      return;
    }

    resettingPasswordTenantID = tenant.tenant_id;
    error = "";
    success = "";
    try {
      await api.resetTenantAdminPassword(tenant.tenant_id, newPassword);
      // Limpa a senha digitada apos aplicacao para reduzir exposicao em tela.
      tenantResetPasswords[tenant.tenant_id] = "";
      success = "Senha do administrador redefinida com sucesso.";
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao redefinir senha do administrador";
    } finally {
      resettingPasswordTenantID = "";
    }
  }

  async function approve(userID: string): Promise<void> {
    savingUserID = userID;
    error = "";
    success = "";
    try {
      const approvedItem = items.find((item) => item.user_id === userID);
      const password = pendingPasswords[userID] ?? "";
      await api.approvePendingRegistration(userID, password);
      // Atualiza a grade para refletir a aprovacao sem precisar recarregar a pagina inteira.
      items = items.filter((item) => item.user_id !== userID);
      if (approvedItem) {
        // Mantem os totais de usuarios ativos da imobiliaria sincronizados apos aprovacao.
        tenants = tenants.map((tenant) =>
          tenant.tenant_id === approvedItem.tenant_id
            ? { ...tenant, active_users: tenant.active_users + 1 }
            : tenant
        );
      }
      delete pendingPasswords[userID];
      success = "Cadastro aprovado com sucesso.";
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao aprovar cadastro";
    } finally {
      savingUserID = "";
    }
  }

  $: filteredItems = filterPendingRegistrations(items, query);
  $: filteredTenants = filterTenantSummaries(tenants, query);

  $: totalPending = items.length;
  $: totalFiltered = filteredItems.length;
  $: totalTenants = tenants.length;
  $: totalFilteredTenants = filteredTenants.length;

  onMount(async () => {
    if (!requirePlatformAdmin()) {
      return;
    }
    await loadAdminData();
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
      <div class="stat-card">
        <span class="stat-label">Imobiliarias cadastradas</span>
        <strong class="stat-value">{totalTenants}</strong>
      </div>
      <div class="stat-card">
        <span class="stat-label">Imobiliarias no filtro</span>
        <strong class="stat-value">{totalFilteredTenants}</strong>
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
      <button class="btn ghost" on:click={loadAdminData} disabled={loading}>Atualizar</button>
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

  <div class="panel">
    <div class="page-head">
      <h2>Imobiliarias Cadastradas</h2>
      <p>Consulte os tenants ja criados e a situacao de usuarios de cada imobiliaria.</p>
    </div>

    <div class="notice info">
      A senha atual nao pode ser exibida: por seguranca ela e armazenada criptografada. Use "Redefinir senha" para definir uma nova.
    </div>

    {#if loading}
      <p>Carregando imobiliarias...</p>
    {:else}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Imobiliaria</th>
              <th>CNPJ</th>
              <th>Email admin</th>
              <th>Usuarios</th>
              <th>Usuarios ativos</th>
              <th>Cadastro</th>
              <th>Redefinir senha admin</th>
              <th>Acoes</th>
            </tr>
          </thead>
          <tbody>
            {#if filteredTenants.length === 0}
              <tr>
                <td colspan="8">Nenhuma imobiliaria encontrada para o filtro informado.</td>
              </tr>
            {:else}
              {#each filteredTenants as tenant}
                <tr>
                  <td>
                    {#if editingTenantID === tenant.tenant_id}
                      <input
                        placeholder="Nome da imobiliaria"
                        bind:value={tenantEditName[tenant.tenant_id]}
                      />
                    {:else}
                      {tenant.tenant_name}
                    {/if}
                  </td>
                  <td>
                    {#if editingTenantID === tenant.tenant_id}
                      <input
                        placeholder="CNPJ"
                        bind:value={tenantEditCNPJ[tenant.tenant_id]}
                      />
                    {:else}
                      {tenant.tenant_cnpj || "-"}
                    {/if}
                  </td>
                  <td>{tenant.admin_email || "-"}</td>
                  <td>{tenant.total_users}</td>
                  <td>{tenant.active_users}</td>
                  <td>{new Date(tenant.created_at).toLocaleString("pt-BR")}</td>
                  <td>
                    <div class="row-inline">
                      <input
                        type="password"
                        placeholder="Nova senha (min 8)"
                        bind:value={tenantResetPasswords[tenant.tenant_id]}
                      />
                      <button
                        class="btn ghost"
                        on:click={() => resetTenantAdminPassword(tenant)}
                        disabled={resettingPasswordTenantID === tenant.tenant_id || !tenant.admin_email}
                      >
                        {resettingPasswordTenantID === tenant.tenant_id ? "Aplicando..." : "Aplicar"}
                      </button>
                    </div>
                  </td>
                  <td>
                    <div class="row-actions">
                      {#if editingTenantID === tenant.tenant_id}
                        <button
                          class="btn primary"
                          on:click={() => saveTenant(tenant.tenant_id)}
                          disabled={savingTenantID === tenant.tenant_id}
                        >
                          {savingTenantID === tenant.tenant_id ? "Salvando..." : "Salvar"}
                        </button>
                        <button class="btn ghost" on:click={() => cancelTenantEdit(tenant.tenant_id)}>
                          Cancelar
                        </button>
                      {:else}
                        <button class="btn ghost" on:click={() => beginTenantEdit(tenant)}>Editar</button>
                      {/if}
                      <button
                        class="btn danger"
                        on:click={() => removeTenant(tenant)}
                        disabled={deletingTenantID === tenant.tenant_id}
                      >
                        {deletingTenantID === tenant.tenant_id ? "Excluindo..." : "Excluir"}
                      </button>
                    </div>
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

<style>
  .row-inline {
    display: flex;
    gap: 8px;
    align-items: center;
    min-width: 260px;
  }

  .row-actions {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
    min-width: 190px;
  }
</style>
