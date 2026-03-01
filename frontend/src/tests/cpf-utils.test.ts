import { describe, expect, it } from "vitest";
import { formatCpf, isCompleteCpf, onlyCpfDigits } from "../lib/utils/cpf";

describe("cpf utils", () => {
  it("deve normalizar e mascarar CPF", () => {
    expect(onlyCpfDigits("111.222.333-44")).toBe("11122233344");
    expect(formatCpf("11122233344")).toBe("111.222.333-44");
    expect(formatCpf("111.222.333-44")).toBe("111.222.333-44");
  });

  it("deve identificar quando o CPF esta completo", () => {
    expect(isCompleteCpf("111.222.333-44")).toBe(true);
    expect(isCompleteCpf("111.222.333-4")).toBe(false);
  });
});
