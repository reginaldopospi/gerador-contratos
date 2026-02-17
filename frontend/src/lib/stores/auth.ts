import { get, writable } from "svelte/store";
import type { AuthUser } from "../types";

const STORAGE_KEY = "contract_app_auth";

export interface AuthState {
  accessToken: string;
  refreshToken: string;
  user?: AuthUser;
}

function loadState(): AuthState {
  if (typeof window === "undefined") {
    return { accessToken: "", refreshToken: "" };
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return { accessToken: "", refreshToken: "" };
  }

  try {
    const parsed = JSON.parse(raw) as AuthState;
    return {
      accessToken: parsed.accessToken ?? "",
      refreshToken: parsed.refreshToken ?? "",
      user: parsed.user
    };
  } catch {
    return { accessToken: "", refreshToken: "" };
  }
}

export const authStore = writable<AuthState>(loadState());

authStore.subscribe((value) => {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
});

export function setAuth(state: AuthState): void {
  authStore.set(state);
}

export function clearAuth(): void {
  authStore.set({ accessToken: "", refreshToken: "" });
}

export function getAuthState(): AuthState {
  return get(authStore);
}

export function isAuthenticated(): boolean {
  return get(authStore).accessToken.length > 0;
}
