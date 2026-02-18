import { describe, expect, it, vi } from "vitest";
import {
  formatCep,
  isCompleteCep,
  lookupAddressByCep,
  mapViaCepResponse,
  onlyCepDigits
} from "../lib/utils/cep";

describe("cep utils", () => {
  it("deve normalizar e mascarar CEP", () => {
    expect(onlyCepDigits("08.663-040")).toBe("08663040");
    expect(formatCep("08663040")).toBe("08663-040");
    expect(formatCep("08663-040")).toBe("08663-040");
  });

  it("deve identificar quando o CEP esta completo", () => {
    expect(isCompleteCep("08663-040")).toBe(true);
    expect(isCompleteCep("08663-04")).toBe(false);
  });

  it("deve mapear payload valido do ViaCEP", () => {
    const mapped = mapViaCepResponse({
      cep: "08663-040",
      logradouro: "Rua Exemplo",
      complemento: "Apto 32",
      bairro: "Centro",
      localidade: "Guarulhos",
      uf: "SP"
    });

    expect(mapped).toEqual({
      cep: "08663-040",
      logradouro: "Rua Exemplo",
      complemento: "Apto 32",
      bairro: "Centro",
      cidade: "Guarulhos",
      uf: "SP"
    });
  });

  it("deve retornar null quando o ViaCEP indicar erro", () => {
    expect(mapViaCepResponse({ erro: true })).toBeNull();
  });
});

describe("lookupAddressByCep", () => {
  it("deve consultar o endpoint do ViaCEP e mapear endereco", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          cep: "08663-040",
          logradouro: "Rua Exemplo",
          complemento: "",
          bairro: "Centro",
          localidade: "Guarulhos",
          uf: "SP"
        }),
        { status: 200 }
      )
    );

    const result = await lookupAddressByCep("08663-040");

    expect(fetchMock).toHaveBeenCalledWith("https://viacep.com.br/ws/08663040/json/", {
      method: "GET",
      headers: {
        Accept: "application/json"
      }
    });
    expect(result?.cidade).toBe("Guarulhos");
    expect(result?.uf).toBe("SP");
    fetchMock.mockRestore();
  });
});
