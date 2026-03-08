import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { APIError, api } from "../lib/api";
import { clearAuth } from "../lib/stores/auth";

function createLoginPayload() {
  // Payload minimo para validar o fluxo de sucesso do endpoint de login.
  return {
    tokens: {
      access_token: "access-token",
      refresh_token: "refresh-token",
      access_expires_at: "2026-03-08T19:30:00Z",
      refresh_expires_at: "2026-03-15T19:30:00Z"
    },
    user: {
      id: "user-id",
      email: "admin@plataforma.local",
      role: "admin",
      tenant_id: "tenant-id",
      is_platform_admin: true
    }
  };
}

describe("api.login retry", () => {
  beforeEach(() => {
    clearAuth();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("reexecuta login quando a API responde 5xx temporariamente", async () => {
    const fetchMock = vi.fn<() => Promise<Response>>();
    fetchMock
      // Primeira tentativa: backend indisponivel.
      .mockResolvedValueOnce(new Response("{}", { status: 503, statusText: "Service Unavailable" }))
      // Segunda tentativa: backend ainda iniciando.
      .mockResolvedValueOnce(new Response("{}", { status: 503, statusText: "Service Unavailable" }))
      // Terceira tentativa: backend voltou e login funciona.
      .mockResolvedValueOnce(new Response(JSON.stringify(createLoginPayload()), { status: 200 }));

    vi.stubGlobal("fetch", fetchMock);

    const result = await api.login({ email: "admin", password: "Admin12345" });

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(result.tokens.access_token).toBe("access-token");
  });

  it("mantem erro de rede quando todas as tentativas falham", async () => {
    const fetchMock = vi.fn<() => Promise<Response>>();
    fetchMock.mockRejectedValue(new Error("network down"));
    vi.stubGlobal("fetch", fetchMock);

    const thrown = await api
      .login({ email: "admin", password: "Admin12345" })
      .catch((err) => err as APIError);
    expect(thrown).toBeInstanceOf(APIError);
    expect(thrown.code).toBe("network_error");
    expect(fetchMock).toHaveBeenCalledTimes(4);
  });
});
