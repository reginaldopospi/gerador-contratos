import { describe, expect, it } from "vitest";
import {
  buildContractPreviewBlocks,
  isContractHeadingLine,
  resolveContractPreviewTitle
} from "../lib/utils/contract-preview-format";

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

  it("deve centralizar QUADRO RESUMO como cabecalho dedicado", () => {
    const blocks = buildContractPreviewBlocks(
      "COMPROMISSO DE COMPRA E VENDA\nNUMERO DO CONTRATO: 001\nQUADRO RESUMO\nTexto normal"
    );

    expect(blocks[2]).toEqual({ kind: "centered_heading", text: "QUADRO RESUMO" });
  });

  it("deve destacar labels de partes/imovel/intermediadora em caixa alta", () => {
    const blocks = buildContractPreviewBlocks(
      "TITULO\nSUBTITULO\nparte vendedora\nparte compradora\nimovel\nintermediadora"
    );

    expect(blocks[2]).toEqual({ kind: "section_label", text: "PARTE VENDEDORA" });
    expect(blocks[3]).toEqual({ kind: "section_label", text: "PARTE COMPRADORA" });
    expect(blocks[4]).toEqual({ kind: "section_label", text: "IM\u00D3VEL" });
    expect(blocks[5]).toEqual({ kind: "section_label", text: "INTERMEDIADORA" });
  });
});

describe("resolveContractPreviewTitle", () => {
  it("deve normalizar compromisso de compra e venda para titulo padrao", () => {
    const title = resolveContractPreviewTitle("Compromisso de Venda e Compra de Imovel com Financiamento");

    expect(title).toEqual({
      mainTitle: "COMPROMISSO DE COMPRA E VENDA",
      subtitle: "DE IM\u00D3VEL RESIDENCIAL"
    });
  });

  it("deve manter fallback em caixa alta para titulos genericos", () => {
    const title = resolveContractPreviewTitle("Contrato particular");

    expect(title).toEqual({
      mainTitle: "CONTRATO PARTICULAR",
      subtitle: ""
    });
  });
});
