import { describe, expect, it } from "vitest";
import { resolveAPIError } from "../lib/utils/api-error";

describe("resolveAPIError", () => {
  it("prioriza mensagem enviada pela API", () => {
    // Quando o backend devolve payload de erro, a UI deve respeitar essa mensagem.
    const result = resolveAPIError(
      {
        error: {
          code: "invalid_token",
          message: "token invalido"
        }
      },
      401,
      "Unauthorized"
    );

    expect(result).toEqual({
      code: "invalid_token",
      message: "token invalido"
    });
  });

  it("retorna mensagem de conectividade quando nao existe resposta HTTP", () => {
    const result = resolveAPIError(undefined, 0, "");

    expect(result).toEqual({
      code: "network_error",
      message: "Nao foi possivel conectar ao servidor. Verifique se a API esta ativa."
    });
  });

  it("retorna mensagem de sessao para 401 sem payload", () => {
    const result = resolveAPIError({}, 401, "Unauthorized");

    expect(result).toEqual({
      code: "request_error",
      message: "Sessao expirada ou invalida. Faca login novamente."
    });
  });

  it("retorna mensagem de indisponibilidade para erro 5xx sem payload", () => {
    const result = resolveAPIError({}, 500, "Internal Server Error");

    expect(result).toEqual({
      code: "request_error",
      message: "Servidor indisponivel no momento. Verifique se a API esta ativa."
    });
  });
});
