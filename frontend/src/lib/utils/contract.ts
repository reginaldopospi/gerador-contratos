export function safeParseJSON<T>(value: string, fallback: T): T {
  try {
    return JSON.parse(value) as T;
  } catch {
    return fallback;
  }
}

// Converte texto em numero BR aceitando formatos com/sem "R$" e separadores de milhar.
export function parseMoneyBR(value: string): number {
  const compact = value.replace(/\s/g, "").replace(/^R\$/i, "");
  if (compact === "") {
    return 0;
  }

  const cleaned = compact.replace(/[^\d,.-]/g, "");
  if (cleaned === "" || cleaned === "-" || cleaned === "," || cleaned === ".") {
    return 0;
  }

  if (cleaned.includes(",")) {
    const normalized = cleaned.replace(/\./g, "").replace(",", ".");
    const parsed = Number(normalized);
    return Number.isFinite(parsed) ? parsed : 0;
  }

  // Quando houver somente pontos, trata grupos de milhar como inteiro.
  const thousandGrouped = /^-?\d{1,3}(\.\d{3})+$/.test(cleaned);
  const normalized = thousandGrouped ? cleaned.replace(/\./g, "") : cleaned;
  const parsed = Number(normalized);
  return Number.isFinite(parsed) ? parsed : 0;
}

// Padroniza exibicao de moeda para BRL com duas casas decimais.
export function formatMoneyBR(value: number): string {
  const safe = Number.isFinite(value) ? value : 0;
  return safe.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

// Aplica mascara de moeda sem forcar "R$ 0,00" quando o campo esta vazio.
export function maskMoneyInputBR(value: string): string {
  const compact = value.replace(/\s/g, "").replace(/^R\$/i, "");
  if (compact === "") {
    return "";
  }
  return formatMoneyBR(parseMoneyBR(value));
}

export function normalizeMoneyBR(value: string): string {
  return formatMoneyBR(parseMoneyBR(value));
}
