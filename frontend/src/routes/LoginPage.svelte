<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { api, APIError } from "../lib/api";
  import { isAuthenticated, setAuth } from "../lib/stores/auth";

  let email = "";
  let password = "";
  let loading = false;
  let error = "";

  onMount(() => {
    if (isAuthenticated()) {
      push("/dashboard");
    }
  });

  async function submit(): Promise<void> {
    loading = true;
    error = "";
    try {
      const result = await api.login({ email, password });
      setAuth({
        accessToken: result.tokens.access_token,
        refreshToken: result.tokens.refresh_token,
        user: result.user
      });
      push("/dashboard");
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao autenticar";
    } finally {
      loading = false;
    }
  }
</script>

<section class="panel auth-card">
  <div class="page-head">
    <h1>Login</h1>
    <p>Entre com seu usuario para acessar o sistema.</p>
  </div>

  <div class="grid">
    <div class="field">
      <label for="email">Email</label>
      <input id="email" type="email" bind:value={email} placeholder="admin@empresa.com" />
    </div>

    <div class="field">
      <label for="password">Senha</label>
      <input id="password" type="password" bind:value={password} placeholder="********" />
    </div>
  </div>

  <div class="actions">
    <button class="btn primary" disabled={loading} on:click={submit}>
      {loading ? "Entrando..." : "Entrar"}
    </button>
    <a href="#/forgot-password" class="auth-links">Esqueci minha senha</a>
  </div>

  {#if error}
    <div class="notice error">{error}</div>
  {/if}
</section>
