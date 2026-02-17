import { describe, expect, it } from "vitest";
import {
  buildContractData,
  draftFromContractData,
  emptyContractDraft
} from "../lib/utils/contract-editor";

describe("draftFromContractData", () => {
  it("deve mapear campos estruturados e extras", () => {
    const draft = draftFromContractData({
      vendedores: ["vend_1"],
      compradores: ["comp_1"],
      clausulas_selecionadas: ["multa_atraso", "foro_eleicao"],
      "vend_1__nome": "Ana Maria",
      "comp_1__razao_social": "Empresa XP LTDA",
      imovel__tipo: "Apartamento",
      clausulas_entrega_chaves: {
        "Na vistoria final": "Entrega apos vistoria assinada."
      },
      prioridade: 10,
      aceita_permuta: true,
      metadados: { canal: "digital" }
    });

    expect(draft.vendedores[0].nome).toBe("Ana Maria");
    expect(draft.compradores[0].razaoSocial).toBe("Empresa XP LTDA");
    expect(draft.clausulasSelecionadas).toEqual(["multa_atraso", "foro_eleicao"]);
    expect(draft.imovelTipo).toBe("Apartamento");
    expect(draft.clausulasEntregaChaves[0].key).toBe("Na vistoria final");

    const extraNumero = draft.extras.find((item) => item.key === "prioridade");
    const extraBool = draft.extras.find((item) => item.key === "aceita_permuta");
    const extraJson = draft.extras.find((item) => item.key === "metadados");

    expect(extraNumero?.type).toBe("number");
    expect(extraBool?.type).toBe("boolean");
    expect(extraJson?.type).toBe("json");
  });
});

describe("buildContractData", () => {
  it("deve converter o rascunho visual em payload de versao", () => {
    const draft = emptyContractDraft();
    draft.vendedores = [{ ref: "vend_1", nome: "Ana", razaoSocial: "" }];
    draft.compradores = [{ ref: "comp_1", nome: "", razaoSocial: "Comprador SA" }];
    draft.clausulasSelecionadas = ["multa_atraso", "foro_eleicao"];
    draft.imovelTipo = "Casa";
    draft.precoTotal = "R$ 350.000,00";
    draft.entregaChaves = "Escrever no contrato";
    draft.entregaChavesTexto = "Entrega em ate 5 dias apos quitacao.";
    draft.clausulasEntregaChaves = [{ key: "Na assinatura", text: "Entrega imediata no ato." }];
    draft.extras = [
      { key: "aceita_permuta", type: "boolean", value: "true" },
      { key: "prioridade", type: "number", value: "7" },
      { key: "metadados", type: "json", value: "{\"origem\":\"crm\"}" }
    ];

    const data = buildContractData(draft);

    expect(data.vendedores).toEqual(["vend_1"]);
    expect(data.compradores).toEqual(["comp_1"]);
    expect(data["vend_1__nome"]).toBe("Ana");
    expect(data["comp_1__razao_social"]).toBe("Comprador SA");
    expect(data.clausulas_selecionadas).toEqual(["multa_atraso", "foro_eleicao"]);
    expect(data.imovel__tipo).toBe("Casa");
    expect(data.preco_total).toBe("R$ 350.000,00");
    expect(data.entrega_chaves_texto).toContain("5 dias");
    expect(data.aceita_permuta).toBe(true);
    expect(data.prioridade).toBe(7);
    expect(data.metadados).toEqual({ origem: "crm" });
  });

  it("deve falhar para JSON invalido em campo adicional", () => {
    const draft = emptyContractDraft();
    draft.extras = [{ key: "dados_livres", type: "json", value: "{" }];

    expect(() => buildContractData(draft)).toThrow(/JSON invalido/i);
  });

  it("deve falhar se chave adicional colidir com formulario principal", () => {
    const draft = emptyContractDraft();
    draft.extras = [{ key: "imovel__tipo", type: "text", value: "Apartamento" }];

    expect(() => buildContractData(draft)).toThrow(/formulario principal/i);
  });
});
