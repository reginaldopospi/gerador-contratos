import { push } from "svelte-spa-router";
import { isAuthenticated } from "../stores/auth";

export function requireAuth(): boolean {
  if (!isAuthenticated()) {
    push("/login");
    return false;
  }
  return true;
}
