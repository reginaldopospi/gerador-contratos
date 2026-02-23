export type PartyRole = "vendedores" | "compradores";
export type ExtraFieldType = "text" | "number" | "boolean" | "json";
export type PartyType = "Pessoa Fisica" | "Pessoa Juridica";

export interface PartyDraft {
  ref: string;
  tipo: PartyType;
  nome: string;
  razaoSocial: string;
  nacionalidade: string;
  nacionalidadeOutra: string;
  rg: string;
  cpf: string;
  profissao: string;
  estadoCivil: string;
  regimeBens: string;
  regimeBensOutro: string;
  conjNome: string;
  conjNacionalidade: string;
  conjNacionalidadeOutra: string;
  conjProfissao: string;
  conjRg: string;
  conjCpf: string;
  cnpj: string;
  repNome: string;
  repCpf: string;
  endCep: string;
  endLogradouro: string;
  endNumero: string;
  endComplemento: string;
  endBairro: string;
  endCidade: string;
  endUf: string;
  endTexto: string;
}

export interface DeliveryClauseDraft {
  key: string;
  text: string;
}

export interface SelectedClauseDraft {
  clauseKey: string;
  title: string;
  content: string;
  index: string;
}

export interface CustomClauseDraft {
  title: string;
  content: string;
  index: string;
}

export interface ExtraFieldDraft {
  key: string;
  type: ExtraFieldType;
  value: string;
}

export interface ContractEditorDraft {
  vendedores: PartyDraft[];
  compradores: PartyDraft[];
  clausulasSelecionadas: SelectedClauseDraft[];
  clausulasCustomizadas: CustomClauseDraft[];
  imovelTipo: string;
  imovelCep: string;
  imovelLogradouro: string;
  imovelNumero: string;
  imovelComplemento: string;
  imovelBairro: string;
  imovelCidade: string;
  imovelUf: string;
  imovelEndereco: string;
  imovelMatricula: string;
  imovelCartorio: string;
  imovelCidadeCartorio: string;
  imovelContribuinte: string;
  imovelParFar: string;
  imovelAlienado: string;
  imovelAlugado: string;
  imovelLocacao: string;
  imovelFicaraBens: string;
  imovelBens: string;
  imovelDescricaoMatricula: string;
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
const SELECTED_CLAUSES_LINKED_KEY = "clausulas_selecionadas_vinculos";
const CUSTOM_CLAUSES_KEY = "clausulas_customizadas";
const SELECTED_CLAUSES_ALIASES = ["clause_keys", "clausulas_keys"] as const;

const FIELD_TO_DATA_KEY = {
  imovelTipo: "imovel__tipo",
  imovelCep: "imovel__end__cep",
  imovelLogradouro: "imovel__end__logradouro",
  imovelNumero: "imovel__end__numero",
  imovelComplemento: "imovel__end__complemento",
  imovelBairro: "imovel__end__bairro",
  imovelCidade: "imovel__end__cidade",
  imovelUf: "imovel__end__uf",
  imovelEndereco: "imovel__end__texto",
  imovelMatricula: "imovel__matricula",
  imovelCartorio: "imovel__cartorio",
  imovelCidadeCartorio: "imovel__cidade_cartorio",
  imovelContribuinte: "imovel__contribuinte",
  imovelParFar: "imovel__par_far",
  imovelAlienado: "imovel__alienado",
  imovelAlugado: "imovel__alugado",
  imovelLocacao: "imovel__locacao",
  imovelFicaraBens: "imovel__ficara_bens",
  imovelBens: "imovel__bens",
  imovelDescricaoMatricula: "imovel__descricao_matricula",
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
  SELECTED_CLAUSES_LINKED_KEY,
  CUSTOM_CLAUSES_KEY,
  ...SELECTED_CLAUSES_ALIASES
]);

const PARTY_FIELD_SUFFIXES = [
  "tipo",
  "nome",
  "razao_social",
  "nacionalidade",
  "nacionalidade_outra",
  "rg",
  "cpf",
  "profissao",
  "estado_civil",
  "regime_bens",
  "regime_bens_outro",
  "conj_nome",
  "conj_nacionalidade",
  "conj_nacionalidade_outra",
  "conj_profissao",
  "conj_rg",
  "conj_cpf",
  "cnpj",
  "rep_nome",
  "rep_cpf",
  "end__cep",
  "end__logradouro",
  "end__numero",
  "end__complemento",
  "end__bairro",
  "end__cidade",
  "end__uf",
  "end__texto"
] as const;

const SELLER_TOKENS = ["vendedor", "cedente", "vend"];
const BUYER_TOKENS = ["comprador", "cessionario", "cessionaria", "comp"];
const PROPERTY_YES_NO_FIELDS = new Set<DraftStringField>([
  "imovelParFar",
  "imovelAlienado",
  "imovelAlugado",
  "imovelFicaraBens"
]);
const PARTY_TYPE_FISICA: PartyType = "Pessoa Fisica";
const PARTY_TYPE_JURIDICA: PartyType = "Pessoa Juridica";
const PARTY_ESTADO_CIVIL_OPTIONS = [
  "solteiro(a)",
  "casado(a)",
  "uniao estavel",
  "divorciado(a)",
  "viuvo(a)"
] as const;

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

export function emptyPartyDraft(role: PartyRole, index: number): PartyDraft {
  return {
    ref: defaultPartyRef(role, index),
    tipo: PARTY_TYPE_FISICA,
    nome: "",
    razaoSocial: "",
    nacionalidade: "brasileiro",
    nacionalidadeOutra: "",
    rg: "",
    cpf: "",
    profissao: "",
    estadoCivil: "solteiro(a)",
    regimeBens: "",
    regimeBensOutro: "",
    conjNome: "",
    conjNacionalidade: "brasileiro",
    conjNacionalidadeOutra: "",
    conjProfissao: "",
    conjRg: "",
    conjCpf: "",
    cnpj: "",
    repNome: "",
    repCpf: "",
    endCep: "",
    endLogradouro: "",
    endNumero: "",
    endComplemento: "",
    endBairro: "",
    endCidade: "",
    endUf: "",
    endTexto: ""
  };
}

export function normalizePartyType(value: string): PartyType {
  const normalized = cleanText(value)
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "");
  if (normalized.includes("juridica")) {
    return PARTY_TYPE_JURIDICA;
  }
  return PARTY_TYPE_FISICA;
}

export function isPartyEstadoCivilComConjuge(estadoCivil: string): boolean {
  const normalized = cleanText(estadoCivil)
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "");
  return normalized === "casado(a)" || normalized === "uniao estavel";
}

// Reutiliza o mesmo formato de endereco completo usado no app Python.
export function buildAddressText(input: {
  cep: string;
  logradouro: string;
  numero: string;
  complemento: string;
  bairro: string;
  cidade: string;
  uf: string;
}): string {
  const logradouro = cleanText(input.logradouro);
  const numero = cleanText(input.numero);
  const complemento = cleanText(input.complemento);
  const bairro = cleanText(input.bairro);
  const cidade = cleanText(input.cidade);
  const uf = cleanText(input.uf);
  const cep = cleanText(input.cep);

  const parts: string[] = [];
  if (logradouro !== "") {
    parts.push(logradouro);
  }
  if (numero !== "") {
    parts.push(`n.o ${numero}`);
  }
  if (complemento !== "") {
    parts.push(complemento);
  }
  if (bairro !== "") {
    parts.push(bairro);
  }
  if (cidade !== "" && uf !== "") {
    parts.push(`${cidade}/${uf}`);
  } else if (cidade !== "") {
    parts.push(cidade);
  } else if (uf !== "") {
    parts.push(uf);
  }

  let text = parts.join(", ");
  if (cep !== "") {
    text = text === "" ? `CEP: ${cep}` : `${text} - CEP: ${cep}`;
  }
  return text.trim();
}

// Mantem a mesma regra do app Python para montar "Endereco completo (gerado)".
export function buildPropertyAddressText(
  draft: Pick<
    ContractEditorDraft,
    | "imovelCep"
    | "imovelLogradouro"
    | "imovelNumero"
    | "imovelComplemento"
    | "imovelBairro"
    | "imovelCidade"
    | "imovelUf"
  >
): string {
  return buildAddressText({
    cep: draft.imovelCep,
    logradouro: draft.imovelLogradouro,
    numero: draft.imovelNumero,
    complemento: draft.imovelComplemento,
    bairro: draft.imovelBairro,
    cidade: draft.imovelCidade,
    uf: draft.imovelUf
  });
}

export function isMatriculaAreaMaior(tipoImovel: string): boolean {
  return cleanText(tipoImovel).toLowerCase().includes("matricula em area maior");
}

export function emptyContractDraft(): ContractEditorDraft {
  return {
    vendedores: [emptyPartyDraft("vendedores", 1)],
    compradores: [emptyPartyDraft("compradores", 1)],
    clausulasSelecionadas: [],
    clausulasCustomizadas: [],
    imovelTipo: "",
    imovelCep: "",
    imovelLogradouro: "",
    imovelNumero: "",
    imovelComplemento: "",
    imovelBairro: "",
    imovelCidade: "",
    imovelUf: "",
    imovelEndereco: "",
    imovelMatricula: "",
    imovelCartorio: "",
    imovelCidadeCartorio: "",
    imovelContribuinte: "",
    imovelParFar: "NAO",
    imovelAlienado: "NAO",
    imovelAlugado: "NAO",
    imovelLocacao: "",
    imovelFicaraBens: "NAO",
    imovelBens: "",
    imovelDescricaoMatricula: "",
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
    const rawValue = getString(data, FIELD_TO_DATA_KEY[field]);
    if (PROPERTY_YES_NO_FIELDS.has(field)) {
      const normalized = normalizeYesNo(rawValue);
      if (normalized !== "") {
        draft[field] = normalized;
      }
      return;
    }
    draft[field] = rawValue;
  });

  draft.vendedores = buildPartyRows(data, "vendedores", SELLER_TOKENS);
  draft.compradores = buildPartyRows(data, "compradores", BUYER_TOKENS);
  draft.clausulasSelecionadas = buildSelectedClauseRows(data);
  draft.clausulasCustomizadas = buildCustomClauseRows(data);
  draft.clausulasEntregaChaves = buildDeliveryClauseRows(data[CLAUSES_KEY]);

  const knownKeys = new Set<string>(BASE_RESERVED_KEYS);
  for (const party of draft.vendedores) {
    addPartyKnownKeys(knownKeys, party.ref);
  }
  for (const party of draft.compradores) {
    addPartyKnownKeys(knownKeys, party.ref);
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

  // Se houver endereco estruturado, ele prevalece sobre o texto manual.
  const generatedPropertyAddress = buildPropertyAddressText(draft);
  if (generatedPropertyAddress !== "") {
    data[FIELD_TO_DATA_KEY.imovelEndereco] = generatedPropertyAddress;
  }

  // Replica a regra do Python: tipo "matricula em area maior" nao envia descricao.
  if (isMatriculaAreaMaior(draft.imovelTipo)) {
    delete data[FIELD_TO_DATA_KEY.imovelDescricaoMatricula];
  }

  const vendedores = writePartyData(data, draft.vendedores, "vendedores");
  const compradores = writePartyData(data, draft.compradores, "compradores");

  if (vendedores.length > 0) {
    data.vendedores = vendedores;
  }
  if (compradores.length > 0) {
    data.compradores = compradores;
  }

  const selectedClauseKeys = uniqueStrings(
    draft.clausulasSelecionadas.map((item) => cleanText(item.clauseKey))
  );
  if (selectedClauseKeys.length > 0) {
    data[SELECTED_CLAUSES_KEY] = selectedClauseKeys;
  }

  const linkedSelectedClauses = draft.clausulasSelecionadas
    .map((item) => {
      const clauseKey = cleanText(item.clauseKey);
      if (clauseKey === "") {
        return null;
      }
      const index = cleanText(item.index);
      if (!isValidClauseIndex(index)) {
        throw new Error(`A clausula '${clauseKey}' precisa de um indice valido (ex.: 1.1.2).`);
      }
      return {
        clause_key: clauseKey,
        title: cleanText(item.title),
        content: cleanText(item.content),
        indice: index
      };
    })
    .filter((item): item is { clause_key: string; title: string; content: string; indice: string } =>
      item !== null
    );
  if (linkedSelectedClauses.length > 0) {
    data[SELECTED_CLAUSES_LINKED_KEY] = linkedSelectedClauses;
  }

  const customClauses = draft.clausulasCustomizadas
    .map((item) => {
      const titulo = cleanText(item.title);
      const conteudo = cleanText(item.content);
      const indice = cleanText(item.index);
      if (titulo === "" && conteudo === "") {
        return null;
      }
      if (!isValidClauseIndex(indice)) {
        throw new Error(`A clausula customizada '${titulo || "sem titulo"}' precisa de um indice valido (ex.: 1.1.2).`);
      }
      return { titulo, conteudo, indice };
    })
    .filter((item): item is { titulo: string; conteudo: string; indice: string } => item !== null);
  if (customClauses.length > 0) {
    data[CUSTOM_CLAUSES_KEY] = customClauses;
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
    if (!partyHasAnyData(party)) {
      continue;
    }

    let ref = cleanText(party.ref);
    if (ref === "") {
      ref = defaultPartyRef(role, index + 1);
    }
    ref = uniqueRef(ref, refs);
    refs.push(ref);

    const tipo = normalizePartyType(party.tipo);
    const nome = cleanText(party.nome);
    const razao = cleanText(party.razaoSocial);
    const endereco = buildAddressText({
      cep: party.endCep,
      logradouro: party.endLogradouro,
      numero: party.endNumero,
      complemento: party.endComplemento,
      bairro: party.endBairro,
      cidade: party.endCidade,
      uf: party.endUf
    });
    const enderecoFinal = endereco !== "" ? endereco : cleanText(party.endTexto);

    data[`${ref}__tipo`] = tipo;
    if (nome !== "") {
      data[`${ref}__nome`] = nome;
    }
    if (razao !== "") {
      data[`${ref}__razao_social`] = razao;
    }

    if (tipo === PARTY_TYPE_JURIDICA) {
      writeIfText(data, `${ref}__cnpj`, party.cnpj);
      writeIfText(data, `${ref}__rep_nome`, party.repNome);
      writeIfText(data, `${ref}__rep_cpf`, party.repCpf);
    } else {
      writeIfText(data, `${ref}__nacionalidade`, party.nacionalidade);
      writeIfText(data, `${ref}__nacionalidade_outra`, party.nacionalidadeOutra);
      writeIfText(data, `${ref}__rg`, party.rg);
      writeIfText(data, `${ref}__cpf`, party.cpf);
      writeIfText(data, `${ref}__profissao`, party.profissao);

      const estadoCivil = sanitizePartyEstadoCivil(party.estadoCivil);
      writeIfText(data, `${ref}__estado_civil`, estadoCivil);
      if (isPartyEstadoCivilComConjuge(estadoCivil)) {
        writeIfText(data, `${ref}__regime_bens`, party.regimeBens);
        writeIfText(data, `${ref}__regime_bens_outro`, party.regimeBensOutro);
        writeIfText(data, `${ref}__conj_nome`, party.conjNome);
        writeIfText(data, `${ref}__conj_nacionalidade`, party.conjNacionalidade);
        writeIfText(data, `${ref}__conj_nacionalidade_outra`, party.conjNacionalidadeOutra);
        writeIfText(data, `${ref}__conj_profissao`, party.conjProfissao);
        writeIfText(data, `${ref}__conj_rg`, party.conjRg);
        writeIfText(data, `${ref}__conj_cpf`, party.conjCpf);
      }
    }

    writeIfText(data, `${ref}__end__cep`, party.endCep);
    writeIfText(data, `${ref}__end__logradouro`, party.endLogradouro);
    writeIfText(data, `${ref}__end__numero`, party.endNumero);
    writeIfText(data, `${ref}__end__complemento`, party.endComplemento);
    writeIfText(data, `${ref}__end__bairro`, party.endBairro);
    writeIfText(data, `${ref}__end__cidade`, party.endCidade);
    writeIfText(data, `${ref}__end__uf`, party.endUf);
    writeIfText(data, `${ref}__end__texto`, enderecoFinal);
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
    return [emptyPartyDraft(role, 1)];
  }

  return refs.map((ref, index) => buildPartyDraftFromData(data, role, ref, index + 1));
}

function inferPartyRefs(data: Record<string, unknown>, tokens: readonly string[]): string[] {
  const refs: string[] = [];
  const keys = Object.keys(data);

  for (const key of keys) {
    const match = /^(.+?)__(.+)$/.exec(key);
    if (!match) {
      continue;
    }

    const ref = match[1];
    const suffix = match[2];
    if (!isPartySuffix(suffix)) {
      continue;
    }
    const lowered = ref.toLowerCase();
    if (!tokens.some((token) => lowered.includes(token))) {
      continue;
    }
    refs.push(ref);
  }

  return uniqueStrings(refs);
}

function buildPartyDraftFromData(
  data: Record<string, unknown>,
  role: PartyRole,
  ref: string,
  index: number
): PartyDraft {
  const party = emptyPartyDraft(role, index);
  party.ref = cleanText(ref) || defaultPartyRef(role, index);
  party.tipo = normalizePartyType(getString(data, `${ref}__tipo`));
  party.nome = getString(data, `${ref}__nome`);
  party.razaoSocial = getString(data, `${ref}__razao_social`);
  party.nacionalidade = getString(data, `${ref}__nacionalidade`) || party.nacionalidade;
  party.nacionalidadeOutra = getString(data, `${ref}__nacionalidade_outra`);
  party.rg = getString(data, `${ref}__rg`);
  party.cpf = getString(data, `${ref}__cpf`);
  party.profissao = getString(data, `${ref}__profissao`);
  party.estadoCivil =
    sanitizePartyEstadoCivil(getString(data, `${ref}__estado_civil`)) || party.estadoCivil;
  party.regimeBens = getString(data, `${ref}__regime_bens`);
  party.regimeBensOutro = getString(data, `${ref}__regime_bens_outro`);
  party.conjNome = getString(data, `${ref}__conj_nome`);
  party.conjNacionalidade =
    getString(data, `${ref}__conj_nacionalidade`) || party.conjNacionalidade;
  party.conjNacionalidadeOutra = getString(data, `${ref}__conj_nacionalidade_outra`);
  party.conjProfissao = getString(data, `${ref}__conj_profissao`);
  party.conjRg = getString(data, `${ref}__conj_rg`);
  party.conjCpf = getString(data, `${ref}__conj_cpf`);
  party.cnpj = getString(data, `${ref}__cnpj`);
  party.repNome = getString(data, `${ref}__rep_nome`);
  party.repCpf = getString(data, `${ref}__rep_cpf`);
  party.endCep = getString(data, `${ref}__end__cep`);
  party.endLogradouro = getString(data, `${ref}__end__logradouro`);
  party.endNumero = getString(data, `${ref}__end__numero`);
  party.endComplemento = getString(data, `${ref}__end__complemento`);
  party.endBairro = getString(data, `${ref}__end__bairro`);
  party.endCidade = getString(data, `${ref}__end__cidade`);
  party.endUf = getString(data, `${ref}__end__uf`);
  party.endTexto = getString(data, `${ref}__end__texto`);
  return party;
}

function addPartyKnownKeys(knownKeys: Set<string>, ref: string): void {
  const cleanRef = cleanText(ref);
  if (cleanRef === "") {
    return;
  }
  for (const suffix of PARTY_FIELD_SUFFIXES) {
    knownKeys.add(`${cleanRef}__${suffix}`);
  }
}

function isPartySuffix(suffix: string): boolean {
  return PARTY_FIELD_SUFFIXES.includes(suffix as (typeof PARTY_FIELD_SUFFIXES)[number]);
}

function sanitizePartyEstadoCivil(value: string): string {
  const normalized = cleanText(value)
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "");
  const matched = PARTY_ESTADO_CIVIL_OPTIONS.find((item) => item === normalized);
  return matched ?? cleanText(value);
}

function partyHasAnyData(party: PartyDraft): boolean {
  const values = [
    party.nome,
    party.razaoSocial,
    party.nacionalidadeOutra,
    party.rg,
    party.cpf,
    party.profissao,
    party.regimeBens,
    party.regimeBensOutro,
    party.conjNome,
    party.conjNacionalidadeOutra,
    party.conjProfissao,
    party.conjRg,
    party.conjCpf,
    party.cnpj,
    party.repNome,
    party.repCpf,
    party.endCep,
    party.endLogradouro,
    party.endNumero,
    party.endComplemento,
    party.endBairro,
    party.endCidade,
    party.endUf,
    party.endTexto
  ];
  return values.some((value) => cleanText(value) !== "");
}

function writeIfText(data: Record<string, unknown>, key: string, value: string): void {
  const clean = cleanText(value);
  if (clean === "") {
    return;
  }
  data[key] = clean;
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

function buildSelectedClauseRows(data: Record<string, unknown>): SelectedClauseDraft[] {
  const linked = asRecordArray(data[SELECTED_CLAUSES_LINKED_KEY]).map((item) => ({
    clauseKey: cleanText(getString(item, "clause_key")),
    title: cleanText(getString(item, "title") || getString(item, "titulo")),
    content: cleanText(getString(item, "content") || getString(item, "conteudo")),
    index: cleanText(getString(item, "indice") || getString(item, "index"))
  }));

  const selectedFromList = uniqueStrings([
    ...asStringArray(data[SELECTED_CLAUSES_KEY]),
    ...SELECTED_CLAUSES_ALIASES.flatMap((alias) => asStringArray(data[alias]))
  ]);

  const keyed = new Map<string, SelectedClauseDraft>();
  for (const row of linked) {
    if (row.clauseKey === "") {
      continue;
    }
    keyed.set(row.clauseKey, row);
  }

  for (const clauseKey of selectedFromList) {
    if (!keyed.has(clauseKey)) {
      keyed.set(clauseKey, {
        clauseKey,
        title: "",
        content: "",
        index: ""
      });
    }
  }

  return [...keyed.values()].sort((a, b) => a.clauseKey.localeCompare(b.clauseKey));
}

function buildCustomClauseRows(data: Record<string, unknown>): CustomClauseDraft[] {
  return asRecordArray(data[CUSTOM_CLAUSES_KEY])
    .map((item) => ({
      title: cleanText(getString(item, "titulo") || getString(item, "title")),
      content: cleanText(getString(item, "conteudo") || getString(item, "content")),
      index: cleanText(getString(item, "indice") || getString(item, "index"))
    }))
    .filter((item) => item.title !== "" || item.content !== "");
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

function asRecordArray(value: unknown): Array<Record<string, unknown>> {
  if (!Array.isArray(value)) {
    return [];
  }

  const out: Array<Record<string, unknown>> = [];
  for (const item of value) {
    const mapped = asRecord(item);
    if (mapped) {
      out.push(mapped);
    }
  }
  return out;
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

function normalizeYesNo(value: string): string {
  // Remove acentos para aceitar entradas como "NÃO" e persistir em formato unico.
  const normalized = cleanText(value)
    .toUpperCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "");
  if (normalized === "SIM") {
    return "SIM";
  }
  if (normalized === "NAO") {
    return "NAO";
  }
  return "";
}

function isValidClauseIndex(value: string): boolean {
  return /^\d+(\.\d+)+$/.test(value.trim());
}
