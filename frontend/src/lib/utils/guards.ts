import { push } from "svelte-spa-router";
import { getAuthState, isAuthenticated } from "../stores/auth";

export function requireAuth(): boolean {
  if (!isAuthenticated()) {
    push("/login");
    return false;
  }
  return true;
}

export function requirePlatformAdmin(): boolean {
  if (!isAuthenticated()) {
    push("/login");
    return false;
  }

  // Garante que apenas o admin da plataforma acesse a area administrativa.
  if (!getAuthState().user?.is_platform_admin) {
    push("/dashboard");
    return false;
  }

  return true;
}
