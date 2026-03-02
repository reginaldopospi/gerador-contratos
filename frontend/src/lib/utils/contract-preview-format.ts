export type ContractPreviewBlockKind = "title" | "heading" | "paragraph" | "blank";

export interface ContractPreviewBlock {
  kind: ContractPreviewBlockKind;
  text: string;
}

// Identifica linhas que devem aparecer como cabecalho juridico em caixa alta.
export function isContractHeadingLine(value: string): boolean {
  const trimmed = value.trim();
  if (trimmed === "") {
    return false;
  }

  const withoutNumericPrefix = trimmed.replace(/^[\d.\-()\s]+/, "");
  const lettersOnly = withoutNumericPrefix.replace(/[^A-Za-z\u00C0-\u00FF]/g, "");
  if (lettersOnly.length < 3) {
    return false;
  }

  const isUppercase = lettersOnly === lettersOnly.toUpperCase();
  const words = withoutNumericPrefix.split(/\s+/).filter((item) => item !== "");
  return isUppercase && words.length <= 10;
}

// Divide o texto integral do contrato em blocos para aplicar estilo na previa.
export function buildContractPreviewBlocks(text: string): ContractPreviewBlock[] {
  const lines = text.split(/\r?\n/);
  const blocks: ContractPreviewBlock[] = [];
  let meaningfulCount = 0;

  for (const line of lines) {
    const trimmed = line.trim();
    if (trimmed === "") {
      blocks.push({ kind: "blank", text: "" });
      continue;
    }

    if (meaningfulCount < 2) {
      blocks.push({ kind: "title", text: trimmed });
      meaningfulCount += 1;
      continue;
    }

    if (isContractHeadingLine(trimmed)) {
      blocks.push({ kind: "heading", text: trimmed });
      meaningfulCount += 1;
      continue;
    }

    blocks.push({ kind: "paragraph", text: trimmed });
    meaningfulCount += 1;
  }

  return blocks;
}
