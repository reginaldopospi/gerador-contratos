import { describe, expect, it } from "vitest";
import { resolveApiBase } from "../lib/utils/api-base";

describe("resolveApiBase", () => {
  it("usa o fallback relativo quando nao existe variavel de ambiente", () => {
    // O fallback relativo permite proxy local e deploy sob o mesmo host.
    expect(resolveApiBase(undefined)).toBe("/api/v1");
  });

  it("mantem URLs absolutas e remove barra final", () => {
    expect(resolveApiBase("http://localhost:8080/api/v1/")).toBe(
      "http://localhost:8080/api/v1"
    );
  });

  it("aceita caminho relativo absoluto", () => {
    expect(resolveApiBase("/api/v1/")).toBe("/api/v1");
  });

  it("normaliza host sem protocolo", () => {
    expect(resolveApiBase("localhost:8080/api/v1")).toBe(
      "http://localhost:8080/api/v1"
    );
  });
});
