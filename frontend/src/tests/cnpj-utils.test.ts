import { describe, expect, it, vi } from "vitest";
import {
  formatCnpj,
  isCompleteCnpj,
  lookupCompanyByCnpj,
  mapBrasilApiCnpjResponse,
  onlyCnpjDigits
} from "../lib/utils/cnpj";

describe("cnpj utils", () => {
  it("deve normalizar e mascarar CNPJ", () => {
    expect(onlyCnpjDigits("12.345.678/0001-99")).toBe("12345678000199");
    expect(formatCnpj("12345678000199")).toBe("12.345.678/0001-99");
    expect(formatCnpj("12.345.678/0001-99")).toBe("12.345.678/0001-99");
  });

  it("deve identificar quando o CNPJ esta completo", () => {
    expect(isCompleteCnpj("12.345.678/0001-99")).toBe(true);
    expect(isCompleteCnpj("12.345.678/0001-9")).toBe(false);
  });

  it("deve mapear payload valido da BrasilAPI", () => {
    const mapped = mapBrasilApiCnpjResponse({
      cnpj: "12345678000199",
      razao_social: "Empresa Teste LTDA",
      descricao_tipo_de_logradouro: "Rua",
      logradouro: "das Flores",
      numero: "123",
      complemento: "Sala 2",
      bairro: "Centro",
      cep: "08663040",
      municipio: "Guarulhos",
      uf: "sp"
    });

    expect(mapped).toEqual({
      cnpj: "12.345.678/0001-99",
      razaoSocial: "Empresa Teste LTDA",
      endereco: {
        cep: "08663040",
        logradouro: "Rua das Flores",
        numero: "123",
        complemento: "Sala 2",
        bairro: "Centro",
        cidade: "Guarulhos",
        uf: "SP"
      }
    });
  });
});

describe("lookupCompanyByCnpj", () => {
  it("deve consultar o endpoint da BrasilAPI e mapear os dados da empresa", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          cnpj: "12345678000199",
          razao_social: "Empresa Teste LTDA",
          descricao_tipo_de_logradouro: "Avenida",
          logradouro: "Brasil",
          numero: "1000",
          complemento: "",
          bairro: "Centro",
          cep: "08663040",
          municipio: "Guarulhos",
          uf: "SP"
        }),
        { status: 200 }
      )
    );

    const result = await lookupCompanyByCnpj("12.345.678/0001-99");

    expect(fetchMock).toHaveBeenCalledWith("https://brasilapi.com.br/api/cnpj/v1/12345678000199", {
      method: "GET",
      headers: {
        Accept: "application/json"
      }
    });
    expect(result?.razaoSocial).toBe("Empresa Teste LTDA");
    expect(result?.endereco.logradouro).toBe("Avenida Brasil");
    fetchMock.mockRestore();
  });
});
