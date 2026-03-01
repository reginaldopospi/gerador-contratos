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

describe("ContractEditorPage payment section", () => {
  beforeEach(() => {
    // Mantem o teste focado em comportamento de UI sem depender de backend real.
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

  it("deve aplicar mascara e indicar falta/sobra em relacao ao preco total", async () => {
    render(ContractEditorPage, { params: { id: "ct-1" } });

    const precoTotalInput = (await screen.findByLabelText("Preco total")) as HTMLInputElement;
    const financiamentoInput = screen.getByLabelText("Financiamento") as HTMLInputElement;
    const fgtsInput = screen.getByLabelText("FGTS") as HTMLInputElement;
    const entradaInput = screen.getByLabelText("Entrada") as HTMLInputElement;
    const sinalInput = screen.getByLabelText("Sinal") as HTMLInputElement;
    const recursoInput = screen.getByLabelText("Recurso proprio") as HTMLInputElement;
    const somaInput = screen.getByLabelText("Soma dos valores preenchidos") as HTMLInputElement;

    await fireEvent.focus(precoTotalInput);
    await fireEvent.input(precoTotalInput, { target: { value: "450000" } });
    expect(precoTotalInput.value).toBe("450000");
    await fireEvent.blur(precoTotalInput);
    expect(precoTotalInput.value).toMatch(/^R\$\s*450\.000,00$/);

    await fireEvent.focus(financiamentoInput);
    await fireEvent.input(financiamentoInput, { target: { value: "200000" } });
    await fireEvent.blur(financiamentoInput);

    await fireEvent.focus(fgtsInput);
    await fireEvent.input(fgtsInput, { target: { value: "100000" } });
    await fireEvent.blur(fgtsInput);

    await fireEvent.focus(entradaInput);
    await fireEvent.input(entradaInput, { target: { value: "100000" } });
    await fireEvent.blur(entradaInput);

    expect(somaInput.value).toMatch(/^R\$\s*400\.000,00$/);
    expect(screen.getByText(/Faltam R\$\s*50\.000,00 para atingir o Preco total\./)).toBeInTheDocument();

    await fireEvent.focus(sinalInput);
    await fireEvent.input(sinalInput, { target: { value: "50000" } });
    await fireEvent.blur(sinalInput);
    expect(somaInput.value).toMatch(/^R\$\s*450\.000,00$/);
    expect(screen.getByText("Soma dos pagamentos confere com o Preco total.")).toBeInTheDocument();

    await fireEvent.focus(recursoInput);
    await fireEvent.input(recursoInput, { target: { value: "10000" } });
    await fireEvent.blur(recursoInput);
    expect(somaInput.value).toMatch(/^R\$\s*460\.000,00$/);
    expect(screen.getByText(/Sobram R\$\s*10\.000,00 em relacao ao Preco total\./)).toBeInTheDocument();
  });
});
