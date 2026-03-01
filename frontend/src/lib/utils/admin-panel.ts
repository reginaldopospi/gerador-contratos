import type { PendingRegistration, TenantSummary } from "../types";

function normalizeQuery(query: string): string {
  return query.trim().toLowerCase();
}

export function filterPendingRegistrations(
  items: PendingRegistration[],
  query: string
): PendingRegistration[] {
  // Mantem o filtro da grade de pendencias em uma funcao pura para facilitar teste.
  const normalizedQuery = normalizeQuery(query);
  if (!normalizedQuery) {
    return items;
  }

  return items.filter((item) => {
    return (
      item.tenant_name.toLowerCase().includes(normalizedQuery) ||
      item.name.toLowerCase().includes(normalizedQuery) ||
      item.email.toLowerCase().includes(normalizedQuery)
    );
  });
}

export function filterTenantSummaries(items: TenantSummary[], query: string): TenantSummary[] {
  // Reaproveita uma busca textual unica para localizar dados da imobiliaria e do admin.
  const normalizedQuery = normalizeQuery(query);
  if (!normalizedQuery) {
    return items;
  }

  return items.filter((item) => {
    return (
      item.tenant_name.toLowerCase().includes(normalizedQuery) ||
      (item.tenant_cnpj ?? "").toLowerCase().includes(normalizedQuery) ||
      (item.admin_email ?? "").toLowerCase().includes(normalizedQuery)
    );
  });
}
