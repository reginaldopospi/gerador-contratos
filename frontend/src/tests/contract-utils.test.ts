import { describe, expect, it } from "vitest";
import { normalizeMoneyBR, safeParseJSON } from "../lib/utils/contract";

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
