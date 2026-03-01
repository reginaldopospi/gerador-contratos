import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
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

describe("ContractEditorPage party type toggle", () => {
  beforeEach(() => {
    // Garante estado inicial consistente para validar reacao imediata no select de tipo.
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

  it("deve trocar imediatamente os campos entre PF e PJ ao alterar o tipo", async () => {
    render(ContractEditorPage, { params: { id: "ct-1" } });

    const typeSelects = await screen.findAllByLabelText("Tipo da parte");
    const sellerTypeSelect = typeSelects[0] as HTMLSelectElement;

    expect(sellerTypeSelect.value).toBe("");
    expect(screen.getAllByText("Selecione o tipo da parte para exibir os campos do formulario.").length).toBeGreaterThan(0);

    await fireEvent.change(sellerTypeSelect, {
      target: { value: "Pessoa Juridica" }
    });
    expect(sellerTypeSelect.value).toBe("Pessoa Juridica");

    await waitFor(() => {
      expect(screen.getByLabelText("CNPJ")).toBeInTheDocument();
    });
    expect(screen.queryByLabelText("CPF n.o")).not.toBeInTheDocument();

    await fireEvent.change(sellerTypeSelect, {
      target: { value: "Pessoa Fisica" }
    });

    await waitFor(() => {
      expect(screen.getByLabelText("CPF n.o")).toBeInTheDocument();
    });
    expect(screen.queryByLabelText("CNPJ")).not.toBeInTheDocument();
  });
});
