<script lang="ts">
  import { api, APIError } from "../lib/api";
  import { push } from "svelte-spa-router";

  let token = "";
  let new_password = "";
  let loading = false;
  let message = "";
  let error = "";

  async function submit(): Promise<void> {
    loading = true;
    error = "";
    message = "";
    try {
      const result = await api.resetPassword({ token, new_password });
      message = result.message;
      setTimeout(() => push("/login"), 1200);
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao redefinir senha";
    } finally {
      loading = false;
    }
  }
</script>

<section class="panel auth-card">
  <div class="page-head">
    <h1>Redefinir Senha</h1>
    <p>Use o token recebido e defina a nova senha.</p>
  </div>

  <div class="grid">
    <div class="field">
      <label for="reset_token">Token</label>
      <input id="reset_token" bind:value={token} placeholder="Cole o token aqui" />
    </div>

    <div class="field">
      <label for="new_password">Nova senha</label>
      <input id="new_password" type="password" bind:value={new_password} placeholder="Minimo 8 caracteres" />
    </div>
  </div>

  <div class="actions">
    <button class="btn primary" disabled={loading} on:click={submit}>
      {loading ? "Redefinindo..." : "Redefinir senha"}
    </button>
    <a href="#/login" class="auth-links">Voltar para login</a>
  </div>

  {#if message}
    <div class="notice success">{message}</div>
  {/if}
  {#if error}
    <div class="notice error">{error}</div>
  {/if}
</section>
