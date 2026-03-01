import { describe, expect, it } from "vitest";
import {
  filterPendingRegistrations,
  filterTenantSummaries
} from "../lib/utils/admin-panel";
import type { PendingRegistration, TenantSummary } from "../lib/types";

const pendingItems: PendingRegistration[] = [
  {
    user_id: "u-1",
    tenant_id: "t-1",
    tenant_name: "Imobiliaria Alpha",
    name: "Joao Silva",
    email: "joao@alpha.com",
    role: "admin",
    is_active: false,
    created_at: "2026-02-28T10:00:00Z"
  },
  {
    user_id: "u-2",
    tenant_id: "t-2",
    tenant_name: "Imobiliaria Beta",
    name: "Maria Souza",
    email: "maria@beta.com",
    role: "admin",
    is_active: false,
    created_at: "2026-02-28T11:00:00Z"
  }
];

const tenantItems: TenantSummary[] = [
  {
    tenant_id: "t-1",
    tenant_name: "Imobiliaria Alpha",
    tenant_cnpj: "12.345.678/0001-90",
    admin_email: "admin@alpha.com",
    total_users: 3,
    active_users: 2,
    created_at: "2026-02-27T08:00:00Z"
  },
  {
    tenant_id: "t-2",
    tenant_name: "Litoral Negocios",
    tenant_cnpj: "98.765.432/0001-11",
    admin_email: "admin@litoral.com",
    total_users: 2,
    active_users: 2,
    created_at: "2026-02-26T08:00:00Z"
  }
];

describe("filterPendingRegistrations", () => {
  it("retorna todos os itens quando a busca esta vazia", () => {
    // Busca vazia nao deve remover registros do painel.
    expect(filterPendingRegistrations(pendingItems, "   ")).toEqual(pendingItems);
  });

  it("filtra por nome da imobiliaria, usuario ou email", () => {
    expect(filterPendingRegistrations(pendingItems, "beta")).toEqual([pendingItems[1]]);
    expect(filterPendingRegistrations(pendingItems, "joao")).toEqual([pendingItems[0]]);
    expect(filterPendingRegistrations(pendingItems, "maria@beta")).toEqual([pendingItems[1]]);
  });
});

describe("filterTenantSummaries", () => {
  it("retorna todos os itens quando a busca esta vazia", () => {
    // Mantem comportamento consistente com a grade de pendencias.
    expect(filterTenantSummaries(tenantItems, "")).toEqual(tenantItems);
  });

  it("filtra por nome, cnpj ou email do admin", () => {
    expect(filterTenantSummaries(tenantItems, "litoral")).toEqual([tenantItems[1]]);
    expect(filterTenantSummaries(tenantItems, "12.345")).toEqual([tenantItems[0]]);
    expect(filterTenantSummaries(tenantItems, "admin@alpha")).toEqual([tenantItems[0]]);
  });
});
