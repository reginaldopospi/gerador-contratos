<script lang="ts">
  import { api, APIError } from "../lib/api";

  let email = "";
  let loading = false;
  let message = "";
  let devToken = "";
  let error = "";

  async function submit(): Promise<void> {
    loading = true;
    error = "";
    message = "";
    devToken = "";

    try {
      const result = await api.forgotPassword(email);
      message = result.message;
      devToken = result.dev_reset_token ?? "";
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao solicitar recuperacao";
    } finally {
      loading = false;
    }
  }
</script>

<section class="panel auth-card">
  <div class="page-head">
    <h1>Recuperar Senha</h1>
    <p>Informe seu email para receber o token de redefinicao.</p>
  </div>

  <div class="field">
    <label for="forgot_email">Email</label>
    <input id="forgot_email" type="email" bind:value={email} placeholder="admin@empresa.com" />
  </div>

  <div class="actions">
    <button class="btn primary" disabled={loading} on:click={submit}>
      {loading ? "Enviando..." : "Solicitar recuperacao"}
    </button>
    <a href="#/reset-password" class="auth-links">Ja tenho token</a>
  </div>

  {#if message}
    <div class="notice success">{message}</div>
  {/if}
  {#if devToken}
    <div class="notice info">
      Ambiente dev - token: <code>{devToken}</code>
    </div>
  {/if}
  {#if error}
    <div class="notice error">{error}</div>
  {/if}
</section>
