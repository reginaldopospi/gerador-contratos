import { fireEvent, render, screen, waitFor, within } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ContractEditorPage from "../routes/ContractEditorPage.svelte";

const { apiMock, requireAuthMock, lookupAddressByCepMock } = vi.hoisted(() => ({
  apiMock: {
    getContract: vi.fn(),
    listClauses: vi.fn(),
    previewContractFromData: vi.fn(),
    addContractVersion: vi.fn(),
    lookupCompanyByCnpj: vi.fn()
  },
  requireAuthMock: vi.fn(),
  lookupAddressByCepMock: vi.fn()
}));

vi.mock("../lib/api", () => {
  class APIError extends Error {}
  return { api: apiMock, APIError };
});

vi.mock("../lib/utils/guards", () => ({
  requireAuth: requireAuthMock
}));

vi.mock("../lib/utils/cep", async () => {
  const actual = await vi.importActual<typeof import("../lib/utils/cep")>("../lib/utils/cep");
  return {
    ...actual,
    lookupAddressByCep: lookupAddressByCepMock
  };
});

describe("ContractEditorPage party input behavior", () => {
  beforeEach(() => {
    // Mantem dados minimos para exercitar apenas o comportamento do formulario.
    vi.clearAllMocks();
    requireAuthMock.mockReturnValue(true);
    apiMock.getContract.mockResolvedValue({
      contract: {
        id: "ct-1",
        tenant_id: "tenant-1",
        numero: "123",
        tipo: "Compromisso de Venda e Compra de Imovel",
        status: "rascunho",
        created_at: "2026-03-01T00:00:00Z",
        updated_at: "2026-03-01T00:00:00Z"
      },
      latest_version: {
        id: "v-1",
        contract_id: "ct-1",
        version_number: 1,
        data: {},
        created_at: "2026-03-01T00:00:00Z"
      },
      versions: [
        {
          id: "v-1",
          contract_id: "ct-1",
          version_number: 1,
          data: {},
          created_at: "2026-03-01T00:00:00Z"
        }
      ]
    });
    apiMock.listClauses.mockResolvedValue([]);
    apiMock.previewContractFromData.mockResolvedValue({
      title: "Previa",
      sections: []
    });
    lookupAddressByCepMock.mockResolvedValue({
      cep: "08663-040",
      logradouro: "Rua das Flores",
      complemento: "",
      bairro: "Centro",
      cidade: "Guarulhos",
      uf: "SP"
    });
  });

  it("deve aplicar defaults de nacionalidade, mascaras e preencher endereco por CEP", async () => {
    render(ContractEditorPage, { params: { id: "ct-1" } });

    const typeSelects = await screen.findAllByLabelText("Tipo da parte");
    const sellerTypeSelect = typeSelects[0] as HTMLSelectElement;
    const sellerCard = sellerTypeSelect.closest(".party-card") as HTMLElement;

    await fireEvent.change(sellerTypeSelect, {
      target: { value: "Pessoa Fisica" }
    });

    // Ao abrir PF, nacionalidade deve iniciar em brasileiro(a).
    const nacionalidadeSelect = (await within(sellerCard).findByLabelText(
      "Nacionalidade"
    )) as HTMLSelectElement;
    expect(nacionalidadeSelect.value).toBe("brasileiro(a)");

    // CPF deve aplicar mascara progressiva durante digitacao.
    const cpfInput = (await within(sellerCard).findByLabelText("CPF n.o")) as HTMLInputElement;
    await fireEvent.input(cpfInput, { target: { value: "11122233344" } });
    expect(cpfInput.value).toBe("111.222.333-44");

    // CEP deve disparar consulta e completar endereco da parte.
    const cepInput = (await within(sellerCard).findByLabelText("CEP")) as HTMLInputElement;
    await fireEvent.input(cepInput, { target: { value: "08663040" } });
    expect(cepInput.value).toBe("08663-040");

    await waitFor(() => {
      expect(within(sellerCard).getByLabelText("Logradouro")).toHaveValue("Rua das Flores");
    });
    expect(within(sellerCard).getByLabelText("Bairro")).toHaveValue("Centro");
    expect(within(sellerCard).getByLabelText("Cidade")).toHaveValue("Guarulhos");
    expect(within(sellerCard).getByLabelText("UF")).toHaveValue("SP");
    expect(lookupAddressByCepMock).toHaveBeenCalledWith("08663040");

    await fireEvent.change(sellerTypeSelect, {
      target: { value: "Pessoa Juridica" }
    });

    // CNPJ e CPF de representante devem manter mascara ao digitar.
    const cnpjInput = (await within(sellerCard).findByLabelText("CNPJ")) as HTMLInputElement;
    await fireEvent.input(cnpjInput, { target: { value: "12345678000199" } });
    expect(cnpjInput.value).toBe("12.345.678/0001-99");

    const repCpfInput = (await within(sellerCard).findByLabelText(
      "CPF do representante"
    )) as HTMLInputElement;
    await fireEvent.input(repCpfInput, { target: { value: "99988877766" } });
    expect(repCpfInput.value).toBe("999.888.777-66");
  });
});
