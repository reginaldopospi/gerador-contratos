import { describe, expect, it } from "vitest";
import { buildContractPreviewBlocks, isContractHeadingLine } from "../lib/utils/contract-preview-format";

describe("isContractHeadingLine", () => {
  it("deve identificar cabecalho juridico em caixa alta", () => {
    // Garante que titulos de clausula sejam tratados como heading.
    expect(isContractHeadingLine("CLAUSULA PRIMEIRA - OBJETO")).toBe(true);
  });

  it("deve ignorar linha com texto comum", () => {
    expect(isContractHeadingLine("A parte compradora pagara o preco ajustado.")).toBe(false);
  });
});

describe("buildContractPreviewBlocks", () => {
  it("deve marcar as duas primeiras linhas validas como titulo", () => {
    const blocks = buildContractPreviewBlocks("COMPROMISSO DE COMPRA E VENDA\nDAS PARTES\nParagrafo comum");

    expect(blocks[0]).toEqual({ kind: "title", text: "COMPROMISSO DE COMPRA E VENDA" });
    expect(blocks[1]).toEqual({ kind: "title", text: "DAS PARTES" });
    expect(blocks[2]).toEqual({ kind: "paragraph", text: "Paragrafo comum" });
  });

  it("deve preservar linha em branco e detectar heading apos o titulo", () => {
    const blocks = buildContractPreviewBlocks(
      "TITULO\nSUBTITULO\n\nCLAUSULA SEGUNDA - DO PRECO\nTexto normal"
    );

    expect(blocks[2]).toEqual({ kind: "blank", text: "" });
    expect(blocks[3]).toEqual({ kind: "heading", text: "CLAUSULA SEGUNDA - DO PRECO" });
    expect(blocks[4]).toEqual({ kind: "paragraph", text: "Texto normal" });
  });
});
