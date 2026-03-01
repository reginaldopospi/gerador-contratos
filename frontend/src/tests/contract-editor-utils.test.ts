import { describe, expect, it } from "vitest";
import {
  buildContractData,
  draftFromContractData,
  emptyPartyDraft,
  emptyContractDraft
} from "../lib/utils/contract-editor";

describe("draftFromContractData", () => {
  it("deve iniciar partes novas sem tipo selecionado", () => {
    const draft = emptyContractDraft();

    // O editor deve exigir selecao explicita antes de abrir os campos de PF/PJ.
    expect(draft.vendedores[0].tipo).toBe("");
    expect(draft.compradores[0].tipo).toBe("");
    // Mantem nacionalidade inicial padrao para reduzir digitacao manual.
    expect(draft.vendedores[0].nacionalidade).toBe("brasileiro(a)");
    expect(draft.vendedores[0].conjNacionalidade).toBe("brasileiro(a)");
  });

  it("deve mapear campos estruturados e extras", () => {
    const draft = draftFromContractData({
      vendedores: ["vend_1"],
      compradores: ["comp_1"],
      clausulas_selecionadas: ["multa_atraso", "foro_eleicao"],
      clausulas_selecionadas_vinculos: [
        {
          clause_key: "multa_atraso",
          title: "Multa por atraso",
          content: "Aplica multa de 2% sobre o saldo.",
          indice: "1.1.2"
        }
      ],
      clausulas_customizadas: [
        {
          titulo: "Clausula X",
          conteudo: "Texto da clausula X",
          indice: "2.1.1"
        }
      ],
      "vend_1__nome": "Ana Maria",
      "vend_1__tipo": "Pessoa Fisica",
      "vend_1__cpf": "111.222.333-44",
      "vend_1__estado_civil": "uniao estavel",
      "vend_1__end__cidade": "Suzano",
      "comp_1__razao_social": "Empresa XP LTDA",
      imovel__tipo: "Apartamento",
      imovel__end__cidade: "Guarulhos",
      imovel__end__uf: "SP",
      imovel__par_far: "NÃO",
      imovel__alienado: "SIM",
      clausulas_entrega_chaves: {
        "Na vistoria final": "Entrega apos vistoria assinada."
      },
      prioridade: 10,
      aceita_permuta: true,
      metadados: { canal: "digital" }
    });

    expect(draft.vendedores[0].nome).toBe("Ana Maria");
    expect(draft.compradores[0].razaoSocial).toBe("Empresa XP LTDA");
    expect(draft.vendedores[0].tipo).toBe("Pessoa Fisica");
    expect(draft.vendedores[0].cpf).toBe("111.222.333-44");
    expect(draft.vendedores[0].estadoCivil).toBe("uniao estavel");
    expect(draft.vendedores[0].endCidade).toBe("Suzano");
    expect(draft.clausulasSelecionadas.map((item) => item.clauseKey)).toEqual([
      "foro_eleicao",
      "multa_atraso"
    ]);
    expect(draft.clausulasSelecionadas.find((item) => item.clauseKey === "multa_atraso")?.index).toBe(
      "1.1.2"
    );
    expect(draft.clausulasCustomizadas[0].title).toBe("Clausula X");
    expect(draft.clausulasCustomizadas[0].index).toBe("2.1.1");
    expect(draft.imovelTipo).toBe("Apartamento");
    expect(draft.imovelCidade).toBe("Guarulhos");
    expect(draft.imovelUf).toBe("SP");
    expect(draft.imovelParFar).toBe("NAO");
    expect(draft.imovelAlienado).toBe("SIM");
    expect(draft.clausulasEntregaChaves[0].key).toBe("Na vistoria final");

    const extraNumero = draft.extras.find((item) => item.key === "prioridade");
    const extraBool = draft.extras.find((item) => item.key === "aceita_permuta");
    const extraJson = draft.extras.find((item) => item.key === "metadados");

    expect(extraNumero?.type).toBe("number");
    expect(extraBool?.type).toBe("boolean");
    expect(extraJson?.type).toBe("json");
  });

  it("deve aceitar aliases titulo/conteudo nas clausulas selecionadas vinculadas", () => {
    const draft = draftFromContractData({
      clausulas_selecionadas_vinculos: [
        {
          clause_key: "clausula_extra",
          titulo: "Clausula extra",
          conteudo: "Texto completo da clausula extra.",
          indice: "1.2.1"
        }
      ]
    });

    expect(draft.clausulasSelecionadas).toEqual([
      {
        clauseKey: "clausula_extra",
        title: "Clausula extra",
        content: "Texto completo da clausula extra.",
        index: "1.2.1"
      }
    ]);
  });
});

describe("buildContractData", () => {
  it("deve converter o rascunho visual em payload de versao", () => {
    const draft = emptyContractDraft();
    draft.vendedores = [{ ...emptyPartyDraft("vendedores", 1), ref: "vend_1", nome: "Ana" }];
    draft.compradores = [{ ...emptyPartyDraft("compradores", 1), ref: "comp_1", razaoSocial: "Comprador SA" }];
    draft.vendedores[0].cpf = "111.222.333-44";
    draft.vendedores[0].estadoCivil = "casado(a)";
    draft.vendedores[0].conjNome = "Maria";
    draft.vendedores[0].conjCpf = "555.666.777-88";
    draft.compradores[0].tipo = "Pessoa Juridica";
    draft.compradores[0].cnpj = "12.345.678/0001-99";
    draft.compradores[0].repNome = "Joao";
    draft.clausulasSelecionadas = [
      {
        clauseKey: "multa_atraso",
        title: "Multa por atraso",
        content: "Aplica multa de 2% sobre o saldo.",
        index: "1.1.2"
      },
      {
        clauseKey: "foro_eleicao",
        title: "Eleicao de foro",
        content: "Foro da situacao do imovel.",
        index: "15.1.1"
      }
    ];
    draft.clausulasCustomizadas = [
      {
        title: "Clausula X",
        content: "Texto da clausula X",
        index: "2.1.1"
      }
    ];
    draft.imovelTipo = "Casa em condominio (matricula em area maior)";
    draft.imovelCep = "08663-040";
    draft.imovelLogradouro = "Rua Exemplo";
    draft.imovelNumero = "245";
    draft.imovelBairro = "Vila Teste";
    draft.imovelCidade = "Guarulhos";
    draft.imovelUf = "SP";
    draft.imovelDescricaoMatricula = "Nao deve ser enviada para area maior";
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
    expect(data["vend_1__tipo"]).toBe("Pessoa Fisica");
    expect(data["vend_1__cpf"]).toBe("111.222.333-44");
    expect(data["vend_1__conj_nome"]).toBe("Maria");
    expect(data["comp_1__tipo"]).toBe("Pessoa Juridica");
    expect(data["comp_1__cnpj"]).toBe("12.345.678/0001-99");
    expect(data["comp_1__rep_nome"]).toBe("Joao");
    expect(data.clausulas_selecionadas).toEqual(["multa_atraso", "foro_eleicao"]);
    expect(data.clausulas_selecionadas_vinculos).toEqual([
      {
        clause_key: "multa_atraso",
        title: "Multa por atraso",
        content: "Aplica multa de 2% sobre o saldo.",
        indice: "1.1.2"
      },
      {
        clause_key: "foro_eleicao",
        title: "Eleicao de foro",
        content: "Foro da situacao do imovel.",
        indice: "15.1.1"
      }
    ]);
    expect(data.clausulas_customizadas).toEqual([
      {
        titulo: "Clausula X",
        conteudo: "Texto da clausula X",
        indice: "2.1.1"
      }
    ]);
    expect(data.imovel__tipo).toBe("Casa em condominio (matricula em area maior)");
    // Endereco completo deve ser gerado a partir dos campos detalhados.
    expect(data.imovel__end__texto).toBe("Rua Exemplo, n.o 245, Vila Teste, Guarulhos/SP - CEP: 08663-040");
    expect(data.imovel__descricao_matricula).toBeUndefined();
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

  it("deve falhar quando clausula selecionada nao tiver indice", () => {
    const draft = emptyContractDraft();
    draft.clausulasSelecionadas = [
      {
        clauseKey: "multa_atraso",
        title: "Multa por atraso",
        content: "Texto da clausula",
        index: ""
      }
    ];

    expect(() => buildContractData(draft)).toThrow(/indice valido/i);
  });

  it("deve falhar se chave adicional colidir com formulario principal", () => {
    const draft = emptyContractDraft();
    draft.extras = [{ key: "imovel__tipo", type: "text", value: "Apartamento" }];

    expect(() => buildContractData(draft)).toThrow(/formulario principal/i);
  });
});
