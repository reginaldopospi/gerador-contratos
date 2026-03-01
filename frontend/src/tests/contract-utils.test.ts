import { describe, expect, it } from "vitest";
import {
  formatMoneyBR,
  maskMoneyInputBR,
  normalizeMoneyBR,
  parseMoneyBR,
  safeParseJSON
} from "../lib/utils/contract";

describe("safeParseJSON", () => {
  it("deve fazer parse quando json valido", () => {
    const result = safeParseJSON<{ ok: boolean }>("{\"ok\":true}", { ok: false });
    expect(result.ok).toBe(true);
  });

  it("deve retornar fallback quando json invalido", () => {
    const result = safeParseJSON<{ ok: boolean }>("{", { ok: false });
    expect(result.ok).toBe(false);
  });
});

describe("normalizeMoneyBR", () => {
  it("deve formatar moeda BR", () => {
    expect(normalizeMoneyBR("R$ 1234,50")).toContain("1.234,50");
  });

  it("deve retornar zero para valor invalido", () => {
    expect(normalizeMoneyBR("abc")).toMatch(/^R\$\s*0,00$/);
  });
});

describe("money helpers", () => {
  it("deve converter texto com milhar para numero", () => {
    // Garante leitura de formatos digitados sem quebrar o valor real.
    expect(parseMoneyBR("R$ 450.000,00")).toBe(450000);
  });

  it("deve aplicar mascara para entrada numerica simples", () => {
    // Mantem consistencia visual ao digitar valor sem pontuacao.
    expect(maskMoneyInputBR("50000")).toMatch(/^R\$\s*50\.000,00$/);
  });

  it("deve formatar numero em BRL", () => {
    expect(formatMoneyBR(1234.5)).toMatch(/^R\$\s*1\.234,50$/);
  });
});
