import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ContractsPage from "../routes/ContractsPage.svelte";

const { apiMock, requireAuthMock, pushMock } = vi.hoisted(() => ({
  apiMock: {
    listContracts: vi.fn(),
    createContract: vi.fn()
  },
  requireAuthMock: vi.fn(),
  pushMock: vi.fn()
}));

vi.mock("../lib/api", () => {
  class APIError extends Error {}
  return { api: apiMock, APIError };
});

vi.mock("../lib/utils/guards", () => ({
  requireAuth: requireAuthMock
}));

vi.mock("svelte-spa-router", () => ({
  push: pushMock
}));

describe("ContractsPage new contract form", () => {
  beforeEach(() => {
    // Mantem os testes isolados sem trafego real de rede/roteamento.
    vi.clearAllMocks();
    requireAuthMock.mockReturnValue(true);
    apiMock.listContracts.mockResolvedValue([]);
    apiMock.createContract.mockResolvedValue({
      contract: {
        id: "ct-1",
        tenant_id: "tenant-1",
        numero: "1981",
        tipo: "Compromisso de Venda e Compra de Imovel",
        status: "rascunho",
        created_at: "2026-03-01T00:00:00Z",
        updated_at: "2026-03-01T00:00:00Z"
      },
      versions: []
    });
  });

  it("nao exibe campo de status e cria contrato com status rascunho", async () => {
    render(ContractsPage);

    expect(await screen.findByText("Nenhum contrato encontrado.")).toBeInTheDocument();
    expect(screen.queryByLabelText("Status")).not.toBeInTheDocument();

    const numeroInput = screen.getByLabelText("Numero");
    await fireEvent.input(numeroInput, { target: { value: "1981" } });
    await fireEvent.click(screen.getByRole("button", { name: "Criar contrato" }));

    await waitFor(() => {
      // Garante que o backend receba status controlado pelo sistema.
      expect(apiMock.createContract).toHaveBeenCalledWith(
        expect.objectContaining({
          numero: "1981",
          status: "rascunho"
        })
      );
    });
    expect(pushMock).toHaveBeenCalledWith("/contracts/ct-1");
  });
});
