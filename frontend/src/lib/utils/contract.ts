export function safeParseJSON<T>(value: string, fallback: T): T {
  try {
    return JSON.parse(value) as T;
  } catch {
    return fallback;
  }
}

export function normalizeMoneyBR(value: string): string {
  const cleaned = value.replace(/[^\d,.-]/g, "").replace(/\./g, "").replace(",", ".");
  const num = Number(cleaned);
  if (Number.isNaN(num)) {
    return "R$ 0,00";
  }
  return num.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}
