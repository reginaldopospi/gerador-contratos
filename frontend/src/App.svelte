<script lang="ts">
  import Router from "svelte-spa-router";
  import { authStore, clearAuth } from "./lib/stores/auth";
  import { push } from "svelte-spa-router";

  import LoginPage from "./routes/LoginPage.svelte";
  import RegisterPage from "./routes/RegisterPage.svelte";
  import ForgotPasswordPage from "./routes/ForgotPasswordPage.svelte";
  import ResetPasswordPage from "./routes/ResetPasswordPage.svelte";
  import DashboardPage from "./routes/DashboardPage.svelte";
  import ContractsPage from "./routes/ContractsPage.svelte";
  import ContractEditorPage from "./routes/ContractEditorPage.svelte";
  import BrokersPage from "./routes/BrokersPage.svelte";
  import ClausesPage from "./routes/ClausesPage.svelte";
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
    "*": NotFoundPage
  };

  $: loggedIn = $authStore.accessToken.length > 0;

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
