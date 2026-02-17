export type PartyRole = "vendedores" | "compradores";
export type ExtraFieldType = "text" | "number" | "boolean" | "json";

export interface PartyDraft {
  ref: string;
  nome: string;
  razaoSocial: string;
}

export interface DeliveryClauseDraft {
  key: string;
  text: string;
}

export interface ExtraFieldDraft {
  key: string;
  type: ExtraFieldType;
  value: string;
}

export interface ContractEditorDraft {
  vendedores: PartyDraft[];
  compradores: PartyDraft[];
  clausulasSelecionadas: string[];
  imovelTipo: string;
  imovelEndereco: string;
  imovelMatricula: string;
  imovelCartorio: string;
  precoTotal: string;
  precoFinanciamento: string;
  precoFgts: string;
  precoEntrada: string;
  precoSinal: string;
  precoRecursoProprio: string;
  precoCartaCredito: string;
  precoSubsidio: string;
  precoParcelamentoTotal: string;
  precoOutros: string;
  quemPagaComissao: string;
  valorComissao: string;
  momentoPagto: string;
  entregaChaves: string;
  entregaChavesTexto: string;
  clausulasEntregaChaves: DeliveryClauseDraft[];
  extras: ExtraFieldDraft[];
}

export type DraftStringField = {
  [K in keyof ContractEditorDraft]: ContractEditorDraft[K] extends string ? K : never;
}[keyof ContractEditorDraft];

const CLAUSES_KEY = "clausulas_entrega_chaves";
const SELECTED_CLAUSES_KEY = "clausulas_selecionadas";
const SELECTED_CLAUSES_ALIASES = ["clause_keys", "clausulas_keys"] as const;

const FIELD_TO_DATA_KEY = {
  imovelTipo: "imovel__tipo",
  imovelEndereco: "imovel__end__texto",
  imovelMatricula: "imovel__matricula",
  imovelCartorio: "imovel__cartorio",
  precoTotal: "preco_total",
  precoFinanciamento: "preco_financiamento",
  precoFgts: "preco_fgts",
  precoEntrada: "preco_entrada",
  precoSinal: "preco_sinal",
  precoRecursoProprio: "preco_recurso_proprio",
  precoCartaCredito: "preco_carta_credito",
  precoSubsidio: "preco_subsidio",
  precoParcelamentoTotal: "preco_parcelamento_total",
  precoOutros: "preco_outros",
  quemPagaComissao: "quem_paga_comissao",
  valorComissao: "valor_comissao",
  momentoPagto: "momento_pagto",
  entregaChaves: "entrega_chaves",
  entregaChavesTexto: "entrega_chaves_texto"
} as const satisfies Record<DraftStringField, string>;

const BASE_RESERVED_KEYS = new Set<string>([
  ...Object.values(FIELD_TO_DATA_KEY),
  "vendedores",
  "compradores",
  CLAUSES_KEY,
  SELECTED_CLAUSES_KEY,
  ...SELECTED_CLAUSES_ALIASES
]);

const SELLER_TOKENS = ["vendedor", "cedente"];
const BUYER_TOKENS = ["comprador", "cessionario", "cessionaria"];

export const DELIVERY_OPTIONS = [
  "30 dias apos credito em conta",
  "30 dias apos assinatura no Banco",
  "30 dias apos assinatura do CCV",
  "No ato da assinatura no Banco",
  "No ato da assinatura do CCV",
  "24 horas do credito em conta",
  "Escrever no contrato"
] as const;

export function defaultPartyRef(role: PartyRole, index: number): string {
  return role === "vendedores" ? `vendedor_${index}` : `comprador_${index}`;
}

export function emptyContractDraft(): ContractEditorDraft {
  return {
    vendedores: [{ ref: defaultPartyRef("vendedores", 1), nome: "", razaoSocial: "" }],
    compradores: [{ ref: defaultPartyRef("compradores", 1), nome: "", razaoSocial: "" }],
    clausulasSelecionadas: [],
    imovelTipo: "",
    imovelEndereco: "",
    imovelMatricula: "",
    imovelCartorio: "",
    precoTotal: "",
    precoFinanciamento: "",
    precoFgts: "",
    precoEntrada: "",
    precoSinal: "",
    precoRecursoProprio: "",
    precoCartaCredito: "",
    precoSubsidio: "",
    precoParcelamentoTotal: "",
    precoOutros: "",
    quemPagaComissao: "",
    valorComissao: "",
    momentoPagto: "",
    entregaChaves: "",
    entregaChavesTexto: "",
    clausulasEntregaChaves: [],
    extras: []
  };
}

export function draftFromContractData(rawData: Record<string, unknown> | null | undefined): ContractEditorDraft {
  const data = asRecord(rawData) ?? {};
  const draft = emptyContractDraft();

  (Object.keys(FIELD_TO_DATA_KEY) as Array<keyof typeof FIELD_TO_DATA_KEY>).forEach((field) => {
    draft[field] = getString(data, FIELD_TO_DATA_KEY[field]);
  });

  draft.vendedores = buildPartyRows(data, "vendedores", SELLER_TOKENS);
  draft.compradores = buildPartyRows(data, "compradores", BUYER_TOKENS);
  draft.clausulasSelecionadas = uniqueStrings([
    ...asStringArray(data[SELECTED_CLAUSES_KEY]),
    ...SELECTED_CLAUSES_ALIASES.flatMap((alias) => asStringArray(data[alias]))
  ]);
  draft.clausulasEntregaChaves = buildDeliveryClauseRows(data[CLAUSES_KEY]);

  const knownKeys = new Set<string>(BASE_RESERVED_KEYS);
  for (const party of draft.vendedores) {
    knownKeys.add(`${party.ref}__nome`);
    knownKeys.add(`${party.ref}__razao_social`);
  }
  for (const party of draft.compradores) {
    knownKeys.add(`${party.ref}__nome`);
    knownKeys.add(`${party.ref}__razao_social`);
  }

  draft.extras = Object.keys(data)
    .filter((key) => !knownKeys.has(key))
    .sort((a, b) => a.localeCompare(b))
    .map((key) => mapToExtraField(key, data[key]));

  return draft;
}

export function buildContractData(draft: ContractEditorDraft): Record<string, unknown> {
  const data: Record<string, unknown> = {};

  for (const [field, dataKey] of Object.entries(FIELD_TO_DATA_KEY) as Array<
    [keyof typeof FIELD_TO_DATA_KEY, string]
  >) {
    const value = cleanText(draft[field]);
    if (value !== "") {
      data[dataKey] = value;
    }
  }

  const vendedores = writePartyData(data, draft.vendedores, "vendedores");
  const compradores = writePartyData(data, draft.compradores, "compradores");

  if (vendedores.length > 0) {
    data.vendedores = vendedores;
  }
  if (compradores.length > 0) {
    data.compradores = compradores;
  }
  if (draft.clausulasSelecionadas.length > 0) {
    data[SELECTED_CLAUSES_KEY] = uniqueStrings(draft.clausulasSelecionadas);
  }

  const clausulasEntrega: Record<string, string> = {};
  for (const clause of draft.clausulasEntregaChaves) {
    const key = cleanText(clause.key);
    const text = cleanText(clause.text);
    if (key === "" || text === "") {
      continue;
    }
    clausulasEntrega[key] = text;
  }
  if (Object.keys(clausulasEntrega).length > 0) {
    data[CLAUSES_KEY] = clausulasEntrega;
  }

  const reservedExtraKeys = new Set<string>([...BASE_RESERVED_KEYS, ...Object.keys(data)]);
  for (let index = 0; index < draft.extras.length; index += 1) {
    const extra = draft.extras[index];
    const parsed = parseExtraField(extra, reservedExtraKeys);
    if (!parsed) {
      continue;
    }
    data[parsed.key] = parsed.value;
  }

  return data;
}

function writePartyData(
  data: Record<string, unknown>,
  parties: PartyDraft[],
  role: PartyRole
): string[] {
  const refs: string[] = [];

  for (let index = 0; index < parties.length; index += 1) {
    const party = parties[index];
    const nome = cleanText(party.nome);
    const razao = cleanText(party.razaoSocial);
    if (nome === "" && razao === "") {
      continue;
    }

    let ref = cleanText(party.ref);
    if (ref === "") {
      ref = defaultPartyRef(role, index + 1);
    }
    ref = uniqueRef(ref, refs);
    refs.push(ref);

    if (nome !== "") {
      data[`${ref}__nome`] = nome;
    }
    if (razao !== "") {
      data[`${ref}__razao_social`] = razao;
    }
  }

  return refs;
}

function uniqueRef(initialRef: string, taken: string[]): string {
  if (!taken.includes(initialRef)) {
    return initialRef;
  }

  let count = 2;
  while (taken.includes(`${initialRef}_${count}`)) {
    count += 1;
  }
  return `${initialRef}_${count}`;
}

function buildPartyRows(
  data: Record<string, unknown>,
  role: PartyRole,
  inferTokens: readonly string[]
): PartyDraft[] {
  const listedRefs = asStringArray(data[role]);
  const inferredRefs = inferPartyRefs(data, inferTokens);

  const refs = uniqueStrings([...listedRefs, ...inferredRefs]);
  if (refs.length === 0) {
    return [{ ref: defaultPartyRef(role, 1), nome: "", razaoSocial: "" }];
  }

  return refs.map((ref) => ({
    ref,
    nome: getString(data, `${ref}__nome`),
    razaoSocial: getString(data, `${ref}__razao_social`)
  }));
}

function inferPartyRefs(data: Record<string, unknown>, tokens: readonly string[]): string[] {
  const refs: string[] = [];
  const keys = Object.keys(data);

  for (const key of keys) {
    const match = /^(.+)__(nome|razao_social)$/.exec(key);
    if (!match) {
      continue;
    }

    const ref = match[1];
    const lowered = ref.toLowerCase();
    if (!tokens.some((token) => lowered.includes(token))) {
      continue;
    }
    refs.push(ref);
  }

  return uniqueStrings(refs);
}

function uniqueStrings(values: string[]): string[] {
  const seen = new Set<string>();
  const out: string[] = [];

  for (const value of values) {
    const clean = cleanText(value);
    if (clean === "" || seen.has(clean)) {
      continue;
    }
    seen.add(clean);
    out.push(clean);
  }

  return out;
}

function buildDeliveryClauseRows(raw: unknown): DeliveryClauseDraft[] {
  const map = asRecord(raw);
  if (!map) {
    return [];
  }

  return Object.keys(map)
    .sort((a, b) => a.localeCompare(b))
    .map((key) => ({ key, text: getString(map, key) }))
    .filter((item) => cleanText(item.key) !== "");
}

function mapToExtraField(key: string, value: unknown): ExtraFieldDraft {
  if (typeof value === "string") {
    return { key, type: "text", value };
  }
  if (typeof value === "number") {
    return { key, type: "number", value: String(value) };
  }
  if (typeof value === "boolean") {
    return { key, type: "boolean", value: value ? "true" : "false" };
  }
  return {
    key,
    type: "json",
    value: JSON.stringify(value ?? null, null, 2)
  };
}

function parseExtraField(
  field: ExtraFieldDraft,
  reservedKeys: Set<string>
): { key: string; value: unknown } | null {
  const key = cleanText(field.key);
  if (key === "") {
    return null;
  }
  if (reservedKeys.has(key)) {
    throw new Error(`A chave '${key}' ja e usada no formulario principal.`);
  }

  let value: unknown;
  switch (field.type) {
    case "text":
      value = field.value;
      break;
    case "number": {
      const raw = cleanText(field.value);
      if (raw === "") {
        throw new Error(`O campo adicional '${key}' precisa de um numero.`);
      }
      const normalized = raw.includes(",") && !raw.includes(".") ? raw.replace(",", ".") : raw;
      const parsed = Number(normalized);
      if (!Number.isFinite(parsed)) {
        throw new Error(`O campo adicional '${key}' tem numero invalido.`);
      }
      value = parsed;
      break;
    }
    case "boolean":
      if (field.value !== "true" && field.value !== "false") {
        throw new Error(`O campo adicional '${key}' precisa ser true ou false.`);
      }
      value = field.value === "true";
      break;
    case "json": {
      const raw = cleanText(field.value);
      if (raw === "") {
        throw new Error(`O campo adicional '${key}' precisa de JSON valido.`);
      }
      try {
        value = JSON.parse(raw) as unknown;
      } catch {
        throw new Error(`O campo adicional '${key}' tem JSON invalido.`);
      }
      break;
    }
    default:
      throw new Error(`Tipo invalido para campo adicional '${key}'.`);
  }

  reservedKeys.add(key);
  return { key, value };
}

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  return value as Record<string, unknown>;
}

function asStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }

  const out: string[] = [];
  for (const item of value) {
    const text = cleanText(String(item ?? ""));
    if (text !== "") {
      out.push(text);
    }
  }
  return uniqueStrings(out);
}

function getString(data: Record<string, unknown>, key: string): string {
  const value = data[key];
  if (value === null || value === undefined) {
    return "";
  }
  return String(value);
}

function cleanText(value: string): string {
  return value.trim();
}
