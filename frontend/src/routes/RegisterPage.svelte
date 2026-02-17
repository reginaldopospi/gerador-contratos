<script lang="ts">
  import { api, APIError } from "../lib/api";
  import { setAuth } from "../lib/stores/auth";
  import { push } from "svelte-spa-router";

  let tenant_name = "";
  let tenant_cnpj = "";
  let name = "";
  let email = "";
  let password = "";
  let loading = false;
  let error = "";

  async function submit(): Promise<void> {
    loading = true;
    error = "";
    try {
      const result = await api.registerTenant({
        tenant_name,
        tenant_cnpj,
        name,
        email,
        password
      });

      setAuth({
        accessToken: result.tokens.access_token,
        refreshToken: result.tokens.refresh_token,
        user: result.user
      });

      push("/dashboard");
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha no cadastro";
    } finally {
      loading = false;
    }
  }
</script>

<section class="panel auth-card">
  <div class="page-head">
    <h1>Criar Conta da Imobiliaria</h1>
    <p>Primeiro acesso: cria o tenant e o usuario administrador.</p>
  </div>

  <div class="grid cols-2">
    <div class="field">
      <label for="tenant_name">Nome da imobiliaria</label>
      <input id="tenant_name" bind:value={tenant_name} placeholder="Imobiliaria Exemplo" />
    </div>

    <div class="field">
      <label for="tenant_cnpj">CNPJ</label>
      <input id="tenant_cnpj" bind:value={tenant_cnpj} placeholder="00.000.000/0000-00" />
    </div>

    <div class="field">
      <label for="admin_name">Nome do administrador</label>
      <input id="admin_name" bind:value={name} placeholder="Nome completo" />
    </div>

    <div class="field">
      <label for="admin_email">Email</label>
      <input id="admin_email" type="email" bind:value={email} placeholder="admin@empresa.com" />
    </div>

    <div class="field">
      <label for="admin_password">Senha</label>
      <input id="admin_password" type="password" bind:value={password} placeholder="Minimo 8 caracteres" />
    </div>
  </div>

  <div class="actions">
    <button class="btn primary" disabled={loading} on:click={submit}>
      {loading ? "Criando..." : "Criar conta"}
    </button>
    <a href="#/login" class="auth-links">Ja tenho acesso</a>
  </div>

  {#if error}
    <div class="notice error">{error}</div>
  {/if}
</section>
