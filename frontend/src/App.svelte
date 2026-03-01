<script lang="ts">
  import { onMount } from "svelte";
  import Router from "svelte-spa-router";
  import { authStore, clearAuth, getAuthState, setAuth } from "./lib/stores/auth";
  import { push } from "svelte-spa-router";
  import { api } from "./lib/api";

  import LoginPage from "./routes/LoginPage.svelte";
  import RegisterPage from "./routes/RegisterPage.svelte";
  import ForgotPasswordPage from "./routes/ForgotPasswordPage.svelte";
  import ResetPasswordPage from "./routes/ResetPasswordPage.svelte";
  import DashboardPage from "./routes/DashboardPage.svelte";
  import ContractsPage from "./routes/ContractsPage.svelte";
  import ContractEditorPage from "./routes/ContractEditorPage.svelte";
  import BrokersPage from "./routes/BrokersPage.svelte";
  import ClausesPage from "./routes/ClausesPage.svelte";
  import PendingApprovalsPage from "./routes/PendingApprovalsPage.svelte";
  import NotFoundPage from "./routes/NotFoundPage.svelte";

  const routes = {
    "/": DashboardPage,
    "/login": LoginPage,
    "/register": RegisterPage,
    "/forgot-password": ForgotPasswordPage,
    "/reset-password": ResetPasswordPage,
    "/dashboard": DashboardPage,
    "/contracts": ContractsPage,
    "/contracts/:id": ContractEditorPage,
    "/brokers": BrokersPage,
    "/clauses": ClausesPage,
    "/admin": PendingApprovalsPage,
    "/approvals": PendingApprovalsPage,
    "*": NotFoundPage
  };

  $: loggedIn = $authStore.accessToken.length > 0;
  $: isPlatformAdmin = $authStore.user?.is_platform_admin === true;

  onMount(async () => {
    const state = getAuthState();
    if (!state.accessToken) {
      return;
    }

    try {
      // Atualiza dados do usuario para manter os controles de permissao sincronizados.
      const result = await api.me();
      setAuth({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: result.user
      });
    } catch {
      // Mantem o comportamento existente de logout quando o token nao for mais valido.
      clearAuth();
      push("/login");
    }
  });

  function logout(): void {
    clearAuth();
    push("/login");
  }
</script>

<div class="app-shell">
  <header class="topbar">
    <div class="inner">
      <div class="brand-block">
        <small class="brand-kicker">Plataforma Imobiliaria</small>
        <strong class="brand">Gerador de Contratos</strong>
      </div>

      {#if loggedIn}
        <nav>
          <a href="#/dashboard">Dashboard</a>
          <a href="#/contracts">Contratos</a>
          <a href="#/brokers">Corretores</a>
          <a href="#/clauses">Clausulas</a>
          {#if isPlatformAdmin}
            <a href="#/admin">Administracao</a>
          {/if}
        </nav>

        <div class="topbar-user">
          <small class="user-email" title={$authStore.user?.email}>{$authStore.user?.email}</small>
          <button class="btn ghost" on:click={logout}>Sair</button>
        </div>
      {:else}
        <nav>
          <a href="#/login">Login</a>
          <a href="#/register">Cadastro</a>
        </nav>
      {/if}
    </div>
  </header>

  <main>
    <Router {routes} />
  </main>
</div>
