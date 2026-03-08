import { fireEvent, render, screen } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ContractEditorPage from "../routes/ContractEditorPage.svelte";

const { apiMock, requireAuthMock } = vi.hoisted(() => ({
  apiMock: {
    getContract: vi.fn(),
    listClauses: vi.fn(),
    previewContractFromData: vi.fn(),
    addContractVersion: vi.fn(),
    lookupCompanyByCnpj: vi.fn()
  },
  requireAuthMock: vi.fn()
}));

vi.mock("../lib/api", () => {
  class APIError extends Error {}
  return { api: apiMock, APIError };
});

vi.mock("../lib/utils/guards", () => ({
  requireAuth: requireAuthMock
}));

describe("ContractEditorPage commission fields", () => {
  beforeEach(() => {
    // Isola o teste de UI sem depender de chamadas reais.
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
  });

  it("deve usar selects fixos para quem paga e momento do pagamento", async () => {
    render(ContractEditorPage, { params: { id: "ct-1" } });

    const payerSelect = (await screen.findByRole("combobox", {
      name: "Quem paga comissao"
    })) as HTMLSelectElement;
    const commissionInput = screen.getByRole("textbox", {
      name: "Valor da comissao"
    }) as HTMLInputElement;
    const momentSelect = screen.getByRole("combobox", {
      name: "Momento do pagamento"
    }) as HTMLSelectElement;

    expect(screen.queryByRole("textbox", { name: "Quem paga comissao" })).not.toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Momento do pagamento" })).not.toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Parte Vendedora/Cedente" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /Parte Compradora/ })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Ambas as Partes" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "NA ESCRITURA" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "NA ASSINATURA DO CONTRATO" })).toBeInTheDocument();
    expect(
      screen.getByRole("option", { name: "NA LIBERACAO DE VALORES NA CONTA DO VENDEDOR" })
    ).toBeInTheDocument();

    await fireEvent.change(payerSelect, { target: { value: "Ambas as Partes" } });
    expect(payerSelect.value).toBe("Ambas as Partes");

    // Garante mascara de percentual no valor da comissao.
    await fireEvent.input(commissionInput, { target: { value: "6" } });
    expect(commissionInput.value).toBe("6%");

    await fireEvent.input(commissionInput, { target: { value: "12,345" } });
    expect(commissionInput.value).toBe("12,34%");

    await fireEvent.change(momentSelect, { target: { value: "NA ESCRITURA" } });
    expect(momentSelect.value).toBe("NA ESCRITURA");
  });
});
