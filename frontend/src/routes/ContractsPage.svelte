<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { api, APIError } from "../lib/api";
  import type { Contract } from "../lib/types";
  import { requireAuth } from "../lib/utils/guards";

  let loading = false;
  let creating = false;
  let error = "";
  let items: Contract[] = [];

  let q = "";
  let numero = "";
  let tipo = "Compromisso de Venda e Compra de Imovel";

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      items = await api.listContracts(q);
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar contratos";
    } finally {
      loading = false;
    }
  }

  async function createContract(): Promise<void> {
    creating = true;
    error = "";
    try {
      const details = await api.createContract({
        numero,
        tipo,
        // Mantem o contrato inicial em rascunho sem expor esse controle na UI.
        status: "rascunho",
        data: {
          contrato__numero: numero,
          contrato__tipo: tipo
        }
      });
      numero = "";
      await load();
      push(`/contracts/${details.contract.id}`);
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao criar contrato";
    } finally {
      creating = false;
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
      <h1>Contratos</h1>
      <p>Gerencie contratos e versoes.</p>
    </div>

    <div class="grid cols-2">
      <div class="field">
        <label for="search_contract">Buscar contrato</label>
        <input id="search_contract" bind:value={q} placeholder="Numero ou tipo" />
      </div>
      <div class="actions">
        <button class="btn ghost" on:click={load}>Pesquisar</button>
      </div>
    </div>

    {#if error}
      <div class="notice error">{error}</div>
    {/if}

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Numero</th>
            <th>Tipo</th>
            <th>Status</th>
            <th>Acoes</th>
          </tr>
        </thead>
        <tbody>
          {#if loading}
            <tr><td colspan="4">Carregando...</td></tr>
          {:else if items.length === 0}
            <tr><td colspan="4">Nenhum contrato encontrado.</td></tr>
          {:else}
            {#each items as item}
              <tr>
                <td>{item.numero}</td>
                <td>{item.tipo}</td>
                <td>{item.status}</td>
                <td>
                  <button class="btn ghost" on:click={() => push(`/contracts/${item.id}`)}>
                    Abrir
                  </button>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>

  <div class="panel">
    <h2>Novo contrato</h2>
    <div class="grid">
      <div class="field">
        <label for="contract_numero">Numero</label>
        <input id="contract_numero" bind:value={numero} placeholder="Ex.: 1981" />
      </div>
      <div class="field span-all">
        <label for="contract_tipo">Tipo</label>
        <select id="contract_tipo" bind:value={tipo}>
          <option>Compromisso de Venda e Compra de Imovel</option>
          <option>Cessao de Posse e Direitos sobre Imovel</option>
        </select>
      </div>
    </div>

    <div class="actions">
      <button class="btn primary" disabled={creating} on:click={createContract}>
        {creating ? "Criando..." : "Criar contrato"}
      </button>
    </div>
  </div>
</section>
