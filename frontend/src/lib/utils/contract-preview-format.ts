export type ContractPreviewBlockKind =
  | "title"
  | "heading"
  | "centered_heading"
  | "section_label"
  | "paragraph"
  | "blank";

export interface ContractPreviewBlock {
  kind: ContractPreviewBlockKind;
  text: string;
}

export interface ContractPreviewTitleLayout {
  mainTitle: string;
  subtitle: string;
}

const CENTERED_HEADING_TOKENS = ["quadro resumo", "das clausulas e condicoes"] as const;
const SECTION_LABEL_TOKENS = [
  "parte vendedora",
  "parte compradora",
  "imovel",
  "intermediadora"
] as const;

// Normaliza headings para comparacao resiliente (caixa, acento e pontuacao).
function normalizeHeadingToken(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-z0-9]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

// Detecta linhas de cabecalho juridico que devem ficar centralizadas na previa.
function isCenteredHeadingLine(value: string): boolean {
  const normalized = normalizeHeadingToken(value);
  return CENTERED_HEADING_TOKENS.some((token) => normalized.startsWith(token));
}

// Detecta labels que devem aparecer em caixa alta e negrito na previa.
function isSectionLabelLine(value: string): boolean {
  const normalized = normalizeHeadingToken(value);
  // Usa correspondencia exata para nao transformar em label linhas longas como "IMOVEL: descricao...".
  return SECTION_LABEL_TOKENS.some((token) => normalized === token);
}

// Mantem labels juridicos em caixa alta para leitura rapida na previa.
function formatSectionLabelText(value: string): string {
  const upper = value.trim().toLocaleUpperCase("pt-BR");
  return upper.replace(/\bIMOVEL\b/g, "IM\u00D3VEL");
}

// Resolve o titulo principal/subtitulo para manter o layout padrao do modelo juridico.
export function resolveContractPreviewTitle(rawTitle: string): ContractPreviewTitleLayout {
  const trimmed = rawTitle.trim();
  if (trimmed === "") {
    return { mainTitle: "", subtitle: "" };
  }

  const normalized = normalizeHeadingToken(trimmed);
  const hasCompromisso = normalized.includes("compromisso");
  const hasCompra = normalized.includes("compra");
  const hasVenda = normalized.includes("venda");
  const hasImovel = normalized.includes("imovel");

  if (hasCompromisso && hasCompra && hasVenda) {
    return {
      mainTitle: "COMPROMISSO DE COMPRA E VENDA",
      subtitle: hasImovel ? "DE IM\u00D3VEL RESIDENCIAL" : ""
    };
  }

  return {
    mainTitle: trimmed.toLocaleUpperCase("pt-BR"),
    subtitle: ""
  };
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

    if (isCenteredHeadingLine(trimmed)) {
      blocks.push({ kind: "centered_heading", text: trimmed.toLocaleUpperCase("pt-BR") });
      meaningfulCount += 1;
      continue;
    }

    if (isSectionLabelLine(trimmed)) {
      blocks.push({ kind: "section_label", text: formatSectionLabelText(trimmed) });
      meaningfulCount += 1;
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
