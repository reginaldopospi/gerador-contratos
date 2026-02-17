<script lang="ts">
  import { onMount } from "svelte";
  import { api, APIError } from "../lib/api";
  import type { Broker, Contract } from "../lib/types";
  import { requireAuth } from "../lib/utils/guards";

  let loading = true;
  let contracts: Contract[] = [];
  let brokers: Broker[] = [];
  let error = "";

  onMount(async () => {
    if (!requireAuth()) {
      return;
    }

    loading = true;
    error = "";
    try {
      [contracts, brokers] = await Promise.all([
        api.listContracts(),
        api.listBrokers()
      ]);
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar dashboard";
    } finally {
      loading = false;
    }
  });
</script>

<section class="section-stack">
  <div class="panel">
    <div class="page-head">
      <h1>Dashboard</h1>
      <p>Visao geral da operacao de contratos.</p>
    </div>

    {#if loading}
      <p>Carregando dados...</p>
    {:else}
      <div class="stat-grid">
        <div class="stat-card">
          <span class="stat-label">Total de contratos</span>
          <strong class="stat-value">{contracts.length}</strong>
        </div>
        <div class="stat-card">
          <span class="stat-label">Total de corretores</span>
          <strong class="stat-value">{brokers.length}</strong>
        </div>
      </div>
    {/if}

    {#if error}
      <div class="notice error">{error}</div>
    {/if}
  </div>

  <div class="panel">
    <h2>Contratos recentes</h2>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Numero</th>
            <th>Tipo</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {#if contracts.length === 0}
            <tr><td colspan="3">Nenhum contrato cadastrado.</td></tr>
          {:else}
            {#each contracts.slice(0, 8) as item}
              <tr>
                <td><a href={`#/contracts/${item.id}`}>{item.numero}</a></td>
                <td>{item.tipo}</td>
                <td>{item.status}</td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</section>
