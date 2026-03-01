import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "../lib/api";
import { clearAuth } from "../lib/stores/auth";

describe("api.lookupCompanyByCnpj", () => {
  beforeEach(() => {
    clearAuth();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("deve mapear os dados da empresa retornados pela API interna", async () => {
    // A API interna retorna snake_case e o editor trabalha com camelCase.
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          company: {
            cnpj: "58.132.597/0001-08",
            razao_social: "EMPRESA EXEMPLO LTDA",
            endereco: {
              cep: "08663040",
              logradouro: "Rua das Flores",
              numero: "123",
              complemento: "Sala 2",
              bairro: "Centro",
              cidade: "Suzano",
              uf: "SP"
            }
          }
        }),
        { status: 200 }
      )
    );

    const result = await api.lookupCompanyByCnpj("58132597000108");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/cnpj/58132597000108",
      expect.objectContaining({
        method: "GET"
      })
    );
    expect(result).toEqual({
      cnpj: "58.132.597/0001-08",
      razaoSocial: "EMPRESA EXEMPLO LTDA",
      endereco: {
        cep: "08663040",
        logradouro: "Rua das Flores",
        numero: "123",
        complemento: "Sala 2",
        bairro: "Centro",
        cidade: "Suzano",
        uf: "SP"
      }
    });
  });

  it("deve retornar null quando o CNPJ nao existir", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          error: {
            code: "cnpj_not_found",
            message: "cnpj nao encontrado na Receita Federal"
          }
        }),
        { status: 404 }
      )
    );

    const result = await api.lookupCompanyByCnpj("58132597000108");
    expect(result).toBeNull();
  });
});
