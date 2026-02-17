<script lang="ts">
  import { onMount } from "svelte";
  import { api, APIError } from "../lib/api";
  import type { Broker } from "../lib/types";
  import { requireAuth } from "../lib/utils/guards";

  let items: Broker[] = [];
  let loading = false;
  let error = "";
  let success = "";

  let nome = "";
  let cpf = "";
  let creci = "";
  let banco = "";
  let agencia = "";
  let conta = "";
  let pix = "";

  let editingId = "";
  let editNome = "";
  let editCpf = "";
  let editCreci = "";
  let editBanco = "";
  let editAgencia = "";
  let editConta = "";
  let editPix = "";

  async function load(): Promise<void> {
    loading = true;
    error = "";
    success = "";
    try {
      items = await api.listBrokers();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar corretores";
    } finally {
      loading = false;
    }
  }

  async function create(): Promise<void> {
    error = "";
    success = "";
    try {
      await api.createBroker({ nome, cpf, creci, banco, agencia, conta, pix });
      nome = "";
      cpf = "";
      creci = "";
      banco = "";
      agencia = "";
      conta = "";
      pix = "";
      success = "Corretor cadastrado com sucesso.";
      await load();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao criar corretor";
    }
  }

  function startEdit(item: Broker): void {
    editingId = item.id;
    editNome = item.nome ?? "";
    editCpf = item.cpf ?? "";
    editCreci = item.creci ?? "";
    editBanco = item.banco ?? "";
    editAgencia = item.agencia ?? "";
    editConta = item.conta ?? "";
    editPix = item.pix ?? "";
    error = "";
    success = "";
  }

  function cancelEdit(): void {
    editingId = "";
    editNome = "";
    editCpf = "";
    editCreci = "";
    editBanco = "";
    editAgencia = "";
    editConta = "";
    editPix = "";
  }

  async function saveEdit(id: string): Promise<void> {
    error = "";
    success = "";
    try {
      await api.updateBroker(id, {
        nome: editNome,
        cpf: editCpf,
        creci: editCreci,
        banco: editBanco,
        agencia: editAgencia,
        conta: editConta,
        pix: editPix
      });
      success = "Corretor atualizado com sucesso.";
      cancelEdit();
      await load();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao atualizar corretor";
    }
  }

  async function remove(id: string): Promise<void> {
    error = "";
    success = "";
    try {
      await api.deleteBroker(id);
      success = "Corretor removido com sucesso.";
      if (editingId === id) {
        cancelEdit();
      }
      await load();
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao remover corretor";
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
      <h1>Corretores</h1>
      <p>Cadastro e manutencao da base de corretores.</p>
    </div>

    <div class="table-wrap">
      <table class="brokers-table">
        <thead>
          <tr>
            <th>Nome</th>
            <th>CPF</th>
            <th>CRECI</th>
            <th>Banco</th>
            <th>Agencia</th>
            <th>Conta</th>
            <th>PIX</th>
            <th>Acoes</th>
          </tr>
        </thead>
        <tbody>
          {#if loading}
            <tr><td colspan="8">Carregando...</td></tr>
          {:else if items.length === 0}
            <tr><td colspan="8">Nenhum corretor cadastrado.</td></tr>
          {:else}
            {#each items as item}
              <tr>
                <td>
                  {#if editingId === item.id}
                    <input bind:value={editNome} />
                  {:else}
                    {item.nome}
                  {/if}
                </td>
                <td>
                  {#if editingId === item.id}
                    <input bind:value={editCpf} />
                  {:else}
                    {item.cpf}
                  {/if}
                </td>
                <td>
                  {#if editingId === item.id}
                    <input bind:value={editCreci} />
                  {:else}
                    {item.creci}
                  {/if}
                </td>
                <td>
                  {#if editingId === item.id}
                    <input bind:value={editBanco} />
                  {:else}
                    {item.banco}
                  {/if}
                </td>
                <td>
                  {#if editingId === item.id}
                    <input bind:value={editAgencia} />
                  {:else}
                    {item.agencia}
                  {/if}
                </td>
                <td>
                  {#if editingId === item.id}
                    <input bind:value={editConta} />
                  {:else}
                    {item.conta}
                  {/if}
                </td>
                <td>
                  {#if editingId === item.id}
                    <input bind:value={editPix} />
                  {:else}
                    {item.pix}
                  {/if}
                </td>
                <td>
                  <div class="brokers-actions">
                    {#if editingId === item.id}
                      <button class="btn primary" on:click={() => saveEdit(item.id)}>Salvar</button>
                      <button class="btn ghost" on:click={cancelEdit}>Cancelar</button>
                    {:else}
                      <button class="btn ghost" on:click={() => startEdit(item)}>Editar</button>
                    {/if}
                    <button class="btn danger" on:click={() => remove(item.id)}>Excluir</button>
                  </div>
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
    <h2>Novo corretor</h2>
    <p>Todos os campos abaixo sao opcionais, exceto nome.</p>
    <div class="grid cols-2">
      <div class="field">
        <label for="broker_nome">Nome</label>
        <input id="broker_nome" bind:value={nome} />
      </div>
      <div class="field">
        <label for="broker_cpf">CPF</label>
        <input id="broker_cpf" bind:value={cpf} />
      </div>
      <div class="field">
        <label for="broker_creci">CRECI (opcional)</label>
        <input id="broker_creci" bind:value={creci} />
      </div>
      <div class="field">
        <label for="broker_banco">Banco</label>
        <input id="broker_banco" bind:value={banco} />
      </div>
      <div class="field">
        <label for="broker_agencia">Agencia</label>
        <input id="broker_agencia" bind:value={agencia} />
      </div>
      <div class="field">
        <label for="broker_conta">Conta</label>
        <input id="broker_conta" bind:value={conta} />
      </div>
      <div class="field">
        <label for="broker_pix">PIX</label>
        <input id="broker_pix" bind:value={pix} />
      </div>
    </div>

    <div class="actions">
      <button class="btn primary" on:click={create}>Cadastrar corretor</button>
    </div>
  </div>
</section>

<style>
  .brokers-table {
    table-layout: fixed;
    min-width: 980px;
  }

  .brokers-table th,
  .brokers-table td {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .brokers-table input {
    width: 100%;
    min-width: 0;
    max-width: 100%;
  }

  .brokers-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .brokers-actions .btn {
    padding: 7px 10px;
    font-size: 0.85rem;
  }

  @media (max-width: 900px) {
    .brokers-table {
      min-width: 860px;
    }
  }

  @media (max-width: 680px) {
    .brokers-table {
      min-width: 760px;
    }
  }
</style>
