import { describe, expect, it } from "vitest";
import { readInputValue, readSelectValue, readTextareaValue } from "../lib/utils/dom-events";

describe("dom-events utils", () => {
  it("deve ler valor de input via target", () => {
    const input = document.createElement("input");
    input.value = "abc";
    const event = { target: input, currentTarget: null } as unknown as Event;

    expect(readInputValue(event)).toBe("abc");
  });

  it("deve usar currentTarget como fallback quando target nao existir", () => {
    const select = document.createElement("select");
    const option = document.createElement("option");
    option.value = "Pessoa Juridica";
    option.text = "Pessoa Juridica";
    select.appendChild(option);
    select.value = "Pessoa Juridica";
    const event = { target: null, currentTarget: select } as unknown as Event;

    expect(readSelectValue(event)).toBe("Pessoa Juridica");
  });

  it("deve ler select via currentTarget quando target for option", () => {
    // Reproduz o caso em que o browser entrega target como OPTION no evento change.
    const select = document.createElement("select");
    const option = document.createElement("option");
    option.value = "Pessoa Juridica";
    option.text = "Pessoa Juridica";
    select.appendChild(option);
    select.value = "Pessoa Juridica";
    const event = { target: option, currentTarget: select } as unknown as Event;

    expect(readSelectValue(event)).toBe("Pessoa Juridica");
  });

  it("deve ler select em evento delegado quando currentTarget nao for o select", () => {
    // Reproduz delegacao de evento: target=OPTION e currentTarget diferente do SELECT.
    const select = document.createElement("select");
    const option = document.createElement("option");
    option.value = "Pessoa Juridica";
    option.text = "Pessoa Juridica";
    select.appendChild(option);
    select.value = "Pessoa Juridica";

    const delegatedEvent = { target: option, currentTarget: document.body } as unknown as Event;
    expect(readSelectValue(delegatedEvent)).toBe("Pessoa Juridica");
  });

  it("deve ler valor de textarea e retornar vazio para tipo inesperado", () => {
    const textarea = document.createElement("textarea");
    textarea.value = "texto";
    const textareaEvent = { target: textarea, currentTarget: null } as unknown as Event;
    expect(readTextareaValue(textareaEvent)).toBe("texto");

    // Evita estourar erro quando o evento vier de outro tipo de elemento.
    const div = document.createElement("div");
    const wrongEvent = { target: div, currentTarget: null } as unknown as Event;
    expect(readInputValue(wrongEvent)).toBe("");
  });
});
