<script lang="ts">
  import { onMount } from "svelte";
  import { Document, HeadingLevel, Packer, Paragraph } from "docx";
  import { api, APIError } from "../lib/api";
  import {
    formatCep,
    isCompleteCep,
    lookupAddressByCep,
    onlyCepDigits
  } from "../lib/utils/cep";
  import {
    formatCnpj,
    isCompleteCnpj,
    onlyCnpjDigits
  } from "../lib/utils/cnpj";
  import { formatCpf } from "../lib/utils/cpf";
  import { formatMoneyBR, maskMoneyInputBR, parseMoneyBR } from "../lib/utils/contract";
  import { readInputValue, readSelectValue, readTextareaValue } from "../lib/utils/dom-events";
  import type { ClauseTemplate, ContractDetails, ContractPreview, ContractVersion } from "../lib/types";
  import {
    buildPropertyAddressText,
    buildAddressText,
    buildContractData,
    DELIVERY_OPTIONS,
    draftFromContractData,
    emptyContractDraft,
    emptyPartyDraft,
    isPartyEstadoCivilComConjuge,
    isMatriculaAreaMaior,
    normalizePartyType,
    type ContractEditorDraft,
    type DraftStringField,
    type PartyDraft,
    type PartyRole
  } from "../lib/utils/contract-editor";
  import { requireAuth } from "../lib/utils/guards";

  export let params: { id: string };

  let details: ContractDetails | null = null;
  let preview: ContractPreview | null = null;
  let draft: ContractEditorDraft = emptyContractDraft();
  let availableClauses: ClauseTemplate[] = [];
  let clauseSearch = "";
  let clauseLoadError = "";
  let selectedVersion = 0;
  let loading = false;
  let saving = false;
  let previewing = false;
  let downloadingDocx = false;
  let error = "";
  let success = "";
  let generatedPropertyAddress = "";
  let propertyAddressPreview = "";
  let hideMatriculaDescription = false;
  let cepLookupStatus: "idle" | "loading" | "error" | "success" = "idle";
  let cepLookupMessage = "";
  let cepLookupRequestId = 0;
  let lastFetchedCep = "";
  let partyConditionalErrors: string[] = [];
  let hasPartyConditionalBlockers = false;
  type LookupStatus = "idle" | "loading" | "error" | "success";
  type LookupState = { status: LookupStatus; message: string };
  let partyCnpjLookup = new Map<string, LookupState>();
  const partyCnpjLookupRequestIds = new Map<string, number>();
  const partyLastFetchedCnpj = new Map<string, string>();
  let partyCepLookup = new Map<string, LookupState>();
  const partyCepLookupRequestIds = new Map<string, number>();
  const partyLastFetchedCep = new Map<string, string>();

  type PropertyToggleField = "imovelParFar" | "imovelAlienado" | "imovelAlugado" | "imovelFicaraBens";

  // Replica as opcoes de tipo de imovel usadas no app Python.
  const PROPERTY_TYPE_OPTIONS = [
    "imovel",
    "apartamento",
    "apartamento (matricula em area maior)",
    "sobrado",
    "sobrado em condominio",
    "sobrado em condominio (matricula em area maior)",
    "casa",
    "casa em condominio",
    "casa em condominio (matricula em area maior)",
    "terreno",
    "outro"
  ] as const;

  const YES_NO_OPTIONS = ["NAO", "SIM"] as const;

  const propertyAddressFields: Array<{ key: DraftStringField; label: string; placeholder: string }> = [
    { key: "imovelCep", label: "CEP", placeholder: "Ex.: 08663-040" },
    { key: "imovelLogradouro", label: "Logradouro", placeholder: "Rua, avenida..." },
    { key: "imovelNumero", label: "Numero", placeholder: "Ex.: 123" },
    { key: "imovelComplemento", label: "Complemento", placeholder: "Ex.: Apto 42" },
    { key: "imovelBairro", label: "Bairro", placeholder: "Ex.: Centro" },
    { key: "imovelCidade", label: "Cidade", placeholder: "Ex.: Guarulhos" },
    { key: "imovelUf", label: "UF", placeholder: "Ex.: SP" }
  ];

  const propertyIdentificationFields: Array<{ key: DraftStringField; label: string; placeholder: string }> = [
    { key: "imovelMatricula", label: "N.o matricula", placeholder: "Ex.: 12345" },
    { key: "imovelCartorio", label: "N.o do cartorio", placeholder: "Ex.: 2" },
    { key: "imovelCidadeCartorio", label: "Cidade do cartorio", placeholder: "Ex.: Guarulhos" },
    { key: "imovelContribuinte", label: "N.o do contribuinte", placeholder: "Ex.: 123.456.789.000" }
  ];

  const propertyToggleFields: Array<{ key: PropertyToggleField; label: string }> = [
    { key: "imovelParFar", label: "Imovel do PAR ou FAR?" },
    { key: "imovelAlienado", label: "Alienado fiduciariamente?" },
    { key: "imovelAlugado", label: "O imovel esta locado a terceiros?" },
    { key: "imovelFicaraBens", label: "Ficara bens no imovel?" }
  ];

  const PARTY_TYPE_OPTIONS = ["Pessoa Fisica", "Pessoa Juridica"] as const;
  const PARTY_NACIONALIDADE_OPTIONS = [
    "brasileiro(a)",
    "brasileiro",
    "brasileira",
    "portuguesa",
    "portugues",
    "italiana",
    "italiano",
    "espanhola",
    "espanhol",
    "argentina",
    "argentino",
    "americana",
    "americano",
    "alema",
    "alemao",
    "francesa",
    "frances",
    "japonesa",
    "japones",
    "chinesa",
    "chines",
    "outra (escrever)"
  ] as const;
  const PARTY_ESTADO_CIVIL_OPTIONS = [
    "solteiro(a)",
    "casado(a)",
    "uniao estavel",
    "divorciado(a)",
    "viuvo(a)"
  ] as const;
  const PARTY_REGIME_BENS_OPTIONS = [
    "comunhao parcial de bens",
    "comunhao universal de bens",
    "separacao total de bens",
    "participacao final nos aquestos",
    "outro (escrever)"
  ] as const;
  const PARTY_OTHER_OPTION = "outra (escrever)";
  const PARTY_REGIME_OTHER_OPTION = "outro (escrever)";
  const PARTY_SECTIONS: Array<{ role: PartyRole; title: string; idPrefix: string }> = [
    { role: "vendedores", title: "Parte vendedora / cedente", idPrefix: "seller" },
    { role: "compradores", title: "Parte compradora / cessionaria", idPrefix: "buyer" }
  ];

  type PaymentMoneyField =
    | "precoTotal"
    | "precoFinanciamento"
    | "precoFgts"
    | "precoEntrada"
    | "precoSinal"
    | "precoRecursoProprio"
    | "precoCartaCredito"
    | "precoSubsidio"
    | "precoParcelamentoTotal"
    | "precoOutros";

  // Mantem uma lista unica para aplicar mascara e calcular fechamento financeiro.
  const PAYMENT_MONEY_FIELDS = [
    "precoTotal",
    "precoFinanciamento",
    "precoFgts",
    "precoEntrada",
    "precoSinal",
    "precoRecursoProprio",
    "precoCartaCredito",
    "precoSubsidio",
    "precoParcelamentoTotal",
    "precoOutros"
  ] as const satisfies ReadonlyArray<PaymentMoneyField>;

  // Soma de comparacao contra o preco total (sem contar o proprio total).
  const PAYMENT_BREAKDOWN_FIELDS = PAYMENT_MONEY_FIELDS.filter(
    (field) => field !== "precoTotal"
  ) as Exclude<PaymentMoneyField, "precoTotal">[];

  const paymentFields: Array<{ key: PaymentMoneyField; label: string; placeholder: string }> = [
    { key: "precoTotal", label: "Preco total", placeholder: "R$ 450.000,00" },
    { key: "precoFinanciamento", label: "Financiamento", placeholder: "R$ 300.000,00" },
    { key: "precoFgts", label: "FGTS", placeholder: "R$ 20.000,00" },
    { key: "precoEntrada", label: "Entrada", placeholder: "R$ 80.000,00" },
    { key: "precoSinal", label: "Sinal", placeholder: "R$ 10.000,00" },
    { key: "precoRecursoProprio", label: "Recurso proprio", placeholder: "R$ 30.000,00" },
    { key: "precoCartaCredito", label: "Carta de credito", placeholder: "R$ 0,00" },
    { key: "precoSubsidio", label: "Subsidio", placeholder: "R$ 0,00" },
    { key: "precoParcelamentoTotal", label: "Parcelamento", placeholder: "R$ 15.000,00" },
    { key: "precoOutros", label: "Outros valores", placeholder: "R$ 0,00" }
  ];

  const COMMISSION_PAYER_OPTIONS = [
    "Parte Vendedora/Cedente",
    "Parte Compradora/Cessionária",
    "Ambas as Partes"
  ] as const;
  // Mantem as opcoes de momento conforme o sistema original.
  const COMMISSION_PAYMENT_MOMENT_OPTIONS = [
    "NA ESCRITURA",
    "NA ASSINATURA DO CONTRATO",
    "NA LIBERACAO DE VALORES NA CONTA DO VENDEDOR"
  ] as const;

  $: {
    const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
    if (hydratedDraft !== draft) {
      draft = hydratedDraft;
    }
  }
  $: selectedClauseKeys = draft.clausulasSelecionadas.map((item) => item.clauseKey);
  $: clauseSuggestions = getClauseSuggestions(clauseSearch, availableClauses, selectedClauseKeys);
  $: generatedPropertyAddress = buildPropertyAddressText(draft);
  $: propertyAddressPreview = generatedPropertyAddress !== "" ? generatedPropertyAddress : draft.imovelEndereco;
  $: hideMatriculaDescription = isMatriculaAreaMaior(draft.imovelTipo);
  $: partyConditionalErrors = collectPartyConditionalErrors(draft);
  $: hasPartyConditionalBlockers = partyConditionalErrors.length > 0;
  let paymentBreakdownTotalFormatted = "R$ 0,00";
  let paymentBalanceNoticeType: "info" | "success" | "error" = "info";
  let paymentBalanceMessage = "Informe o Preco total para validar se falta ou sobra valor.";

  $: {
    // Calcula o fechamento financeiro sempre que qualquer campo de pagamento mudar.
    const precoTotal = roundMoneyValue(parseMoneyBR(draft.precoTotal));
    const somaPagamentos = roundMoneyValue(
      PAYMENT_BREAKDOWN_FIELDS.reduce((total, field) => total + parseMoneyBR(draft[field]), 0)
    );
    const diferenca = roundMoneyValue(precoTotal - somaPagamentos);

    paymentBreakdownTotalFormatted = formatMoneyBR(somaPagamentos);

    if (!hasMoneyDigits(draft.precoTotal)) {
      paymentBalanceNoticeType = "info";
      paymentBalanceMessage = "Informe o Preco total para validar se falta ou sobra valor.";
    } else if (Math.abs(diferenca) < 0.005) {
      paymentBalanceNoticeType = "success";
      paymentBalanceMessage = "Soma dos pagamentos confere com o Preco total.";
    } else if (diferenca > 0) {
      paymentBalanceNoticeType = "error";
      paymentBalanceMessage = `Faltam ${formatMoneyBR(diferenca)} para atingir o Preco total.`;
    } else {
      paymentBalanceNoticeType = "error";
      paymentBalanceMessage = `Sobram ${formatMoneyBR(Math.abs(diferenca))} em relacao ao Preco total.`;
    }
  }

  async function load(): Promise<void> {
    loading = true;
    error = "";
    success = "";
    try {
      details = await api.getContract(params.id);
      try {
        availableClauses = await api.listClauses();
        clauseLoadError = "";
      } catch (clauseErr) {
        availableClauses = [];
        clauseLoadError =
          clauseErr instanceof APIError
            ? clauseErr.message
            : "Falha ao carregar clausulas para o seletor";
      }
      applyVersion(details.latest_version);
      await refreshPreview(false);
    } catch (err) {
      error = err instanceof APIError ? err.message : "Falha ao carregar contrato";
    } finally {
      loading = false;
    }
  }

  function applyVersion(version?: ContractVersion): void {
    draft = hydrateDraftClauseMetadata(draftFromContractData(version?.data ?? {}), availableClauses);
    selectedVersion = version?.version_number ?? 0;
  }

  async function refreshPreview(showNotice: boolean): Promise<void> {
    if (!details) {
      return;
    }
    if (!ensureNoPartyConditionalBlockers()) {
      return;
    }

    previewing = true;
    error = "";
    if (showNotice) {
      success = "";
    }

    try {
      const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
      if (hydratedDraft !== draft) {
        draft = hydratedDraft;
      }
      const data = buildContractData(hydratedDraft);
      preview = await api.previewContractFromData({
        numero: details.contract.numero,
        tipo: details.contract.tipo,
        data
      });
      if (showNotice) {
        success = "Previa atualizada com os dados atuais do formulario.";
      }
    } catch (err) {
      if (err instanceof APIError) {
        error = err.message;
      } else if (err instanceof Error) {
        error = err.message;
      } else {
        error = "Falha ao gerar previa";
      }
    } finally {
      previewing = false;
    }
  }

  async function saveNewVersion(): Promise<void> {
    if (!ensureNoPartyConditionalBlockers()) {
      return;
    }

    saving = true;
    error = "";
    success = "";

    try {
      const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
      if (hydratedDraft !== draft) {
        draft = hydratedDraft;
      }
      const data = buildContractData(hydratedDraft);
      await api.addContractVersion(params.id, data);
      await load();
      success = "Nova versao salva com sucesso.";
    } catch (err) {
      if (err instanceof APIError) {
        error = err.message;
      } else if (err instanceof Error) {
        error = err.message;
      } else {
        error = "Falha ao salvar versao";
      }
    } finally {
      saving = false;
    }
  }

  async function downloadContractDocx(): Promise<void> {
    if (!details) {
      return;
    }
    if (!ensureNoPartyConditionalBlockers()) {
      return;
    }

    downloadingDocx = true;
    error = "";
    success = "";

    try {
      const hydratedDraft = hydrateDraftClauseMetadata(draft, availableClauses);
      if (hydratedDraft !== draft) {
        draft = hydratedDraft;
      }

      const data = buildContractData(hydratedDraft);
      const latestPreview = await api.previewContractFromData({
        numero: details.contract.numero,
        tipo: details.contract.tipo,
        data
      });
      preview = latestPreview;

      const contractText = buildContractText(latestPreview);
      if (contractText.trim() === "") {
        throw new Error("Nao foi possivel montar o contrato para exportacao.");
      }

      const documentBody = new Document({
        sections: [
          {
            children: buildContractParagraphs(latestPreview.title, contractText)
          }
        ]
      });

      const blob = await Packer.toBlob(documentBody);
      triggerBlobDownload(blob, buildDocxFilename(details.contract.numero, details.contract.tipo));
      success = "Contrato DOCX baixado com sucesso.";
    } catch (err) {
      if (err instanceof APIError) {
        error = err.message;
      } else if (err instanceof Error) {
        error = err.message;
      } else {
        error = "Falha ao baixar contrato em DOCX";
      }
    } finally {
      downloadingDocx = false;
    }
  }

  onMount(async () => {
    if (!requireAuth()) {
      return;
    }
    await load();
  });

  function restoreLatestVersion(): void {
    if (!details?.latest_version) {
      draft = emptyContractDraft();
      selectedVersion = 0;
      preview = null;
      success = "Sem versao salva ainda. Formulario limpo.";
      error = "";
      return;
    }

    applyVersion(details.latest_version);
    success = `Versao mais recente v${details.latest_version.version_number} restaurada.`;
    error = "";
    void refreshPreview(false);
  }

  function loadVersion(version: ContractVersion): void {
    applyVersion(version);
    success = `Versao v${version.version_number} carregada para edicao.`;
    error = "";
    void refreshPreview(false);
  }

  function updateField(key: DraftStringField, value: string): void {
    const nextDraft = { ...draft, [key]: value } as ContractEditorDraft;

    // Mantem a regra do Python para autocompletar cidade do cartorio em SP.
    if (
      (key === "imovelCidade" || key === "imovelUf") &&
      nextDraft.imovelUf.trim().toUpperCase() === "SP" &&
      nextDraft.imovelCidade.trim() !== "" &&
      nextDraft.imovelCidadeCartorio.trim() === ""
    ) {
      nextDraft.imovelCidadeCartorio = nextDraft.imovelCidade;
    }

    // Mantem a regra de "matricula em area maior" sem descricao.
    if (key === "imovelTipo" && isMatriculaAreaMaior(value)) {
      nextDraft.imovelDescricaoMatricula = "";
    }

    draft = nextDraft;
  }

  // Mantem digitacao livre (sem distorcer o valor) durante o input.
  function updatePaymentField(key: PaymentMoneyField, value: string): void {
    updateField(key, sanitizePaymentInput(value));
  }

  // Remove prefixo/simbolos e preserva somente o necessario para parse de moeda BR.
  function sanitizePaymentInput(value: string): string {
    const compact = value.replace(/\s/g, "").replace(/^R\$/i, "");
    if (compact === "") {
      return "";
    }
    return compact.replace(/[^\d,.-]/g, "");
  }

  // Ao focar, mostra valor sem mascara para facilitar edicao manual.
  function preparePaymentFieldForEdit(key: PaymentMoneyField): void {
    const current = draft[key];
    if (!current.includes("R$")) {
      return;
    }
    const parsed = roundMoneyValue(parseMoneyBR(current));
    if (Math.abs(parsed - Math.trunc(parsed)) < 0.005) {
      updateField(key, String(Math.trunc(parsed)));
      return;
    }
    updateField(key, parsed.toFixed(2).replace(".", ","));
  }

  // Ao sair do campo, aplica a mascara de moeda no valor digitado.
  function formatPaymentFieldOnBlur(key: PaymentMoneyField): void {
    const current = draft[key];
    if (!hasMoneyDigits(current)) {
      updateField(key, "");
      return;
    }
    updateField(key, maskMoneyInputBR(current));
  }

  function roundMoneyValue(value: number): number {
    return Math.round(value * 100) / 100;
  }

  function hasMoneyDigits(value: string): boolean {
    return /\d/.test(value);
  }

  function listByRole(role: PartyRole): PartyDraft[] {
    return role === "vendedores" ? draft.vendedores : draft.compradores;
  }

  // Mantem dependencia explicita de draft no template para re-render imediato dos blocos PF/PJ.
  function listByRoleForRender(currentDraft: ContractEditorDraft, role: PartyRole): PartyDraft[] {
    return role === "vendedores" ? currentDraft.vendedores : currentDraft.compradores;
  }

  function patchParty(role: PartyRole, index: number, patch: Partial<PartyDraft>): void {
    const list = listByRole(role);
    const next = list.map((item, i) => (i === index ? { ...item, ...patch } : item));
    draft = { ...draft, [role]: next } as ContractEditorDraft;
  }

  function updateParty(role: PartyRole, index: number, key: keyof PartyDraft, value: string): void {
    patchParty(role, index, { [key]: value } as Partial<PartyDraft>);
  }

  function partyTypeOption(value: string): (typeof PARTY_TYPE_OPTIONS)[number] {
    return normalizePartyType(value);
  }

  function selectedPartyTypeOption(value: string): ((typeof PARTY_TYPE_OPTIONS)[number] | "") {
    if (value.trim() === "") {
      return "";
    }
    return partyTypeOption(value);
  }

  function hasPartyTypeSelected(party: PartyDraft): boolean {
    return selectedPartyTypeOption(party.tipo) !== "";
  }

  function isPartyPF(party: PartyDraft): boolean {
    return selectedPartyTypeOption(party.tipo) === "Pessoa Fisica";
  }

  function shouldShowConjuge(party: PartyDraft): boolean {
    return isPartyPF(party) && isPartyEstadoCivilComConjuge(party.estadoCivil);
  }

  function normalizeForComparison(value: string): string {
    return value
      .trim()
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "");
  }

  function partyConjugeNomeLabel(party: PartyDraft): string {
    const estadoCivil = normalizeForComparison(party.estadoCivil);
    return estadoCivil === "uniao estavel" ? "Nome do companheiro(a)" : "Nome do conjuge";
  }

  function isPartyConjugeObrigatorioIncompleto(party: PartyDraft): boolean {
    return shouldShowConjuge(party) && party.conjNome.trim() === "";
  }

  function roleToLabel(role: PartyRole): string {
    return role === "vendedores" ? "vendedor/cedente" : "comprador/cessionario";
  }

  // Replica o bloqueio do app Python quando faltam dados obrigatorios do conjuge.
  function collectPartyConditionalErrors(currentDraft: ContractEditorDraft): string[] {
    const out: string[] = [];
    for (const role of ["vendedores", "compradores"] as const) {
      const list = role === "vendedores" ? currentDraft.vendedores : currentDraft.compradores;
      for (let index = 0; index < list.length; index += 1) {
        if (isPartyConjugeObrigatorioIncompleto(list[index])) {
          out.push(
            `${roleToLabel(role)} ${index + 1}: para CASADO(A) ou UNIAO ESTAVEL, informe o nome do conjuge/companheiro(a).`
          );
        }
      }
    }
    return out;
  }

  function ensureNoPartyConditionalBlockers(): boolean {
    if (!hasPartyConditionalBlockers) {
      return true;
    }
    error = partyConditionalErrors[0] ?? "Existem campos obrigatorios pendentes nas partes.";
    success = "";
    return false;
  }

  function partyLookupKey(role: PartyRole, index: number): string {
    const party = listByRole(role)[index];
    const ref = party?.ref?.trim() ?? "";
    return ref !== "" ? `${role}_${ref}` : `${role}_${index}`;
  }

  function getPartyCnpjLookupState(role: PartyRole, index: number): LookupState {
    return partyCnpjLookup.get(partyLookupKey(role, index)) ?? { status: "idle", message: "" };
  }

  function setPartyCnpjLookupState(
    role: PartyRole,
    index: number,
    status: LookupStatus,
    message: string
  ): void {
    const nextLookup = new Map(partyCnpjLookup);
    nextLookup.set(partyLookupKey(role, index), { status, message });
    partyCnpjLookup = nextLookup;
  }

  function getPartyCepLookupState(role: PartyRole, index: number): LookupState {
    return partyCepLookup.get(partyLookupKey(role, index)) ?? { status: "idle", message: "" };
  }

  function setPartyCepLookupState(
    role: PartyRole,
    index: number,
    status: LookupStatus,
    message: string
  ): void {
    const nextLookup = new Map(partyCepLookup);
    nextLookup.set(partyLookupKey(role, index), { status, message });
    partyCepLookup = nextLookup;
  }

  async function fillPartyFromCnpj(role: PartyRole, index: number, cnpjValue: string): Promise<void> {
    const digits = onlyCnpjDigits(cnpjValue);
    if (!isCompleteCnpj(digits)) {
      return;
    }

    const lookupKey = partyLookupKey(role, index);
    const requestId = (partyCnpjLookupRequestIds.get(lookupKey) ?? 0) + 1;
    partyCnpjLookupRequestIds.set(lookupKey, requestId);
    setPartyCnpjLookupState(role, index, "loading", "Consultando CNPJ...");

    try {
      const company = await api.lookupCompanyByCnpj(digits);
      if (partyCnpjLookupRequestIds.get(lookupKey) !== requestId) {
        return;
      }
      if (!company) {
        setPartyCnpjLookupState(role, index, "error", "CNPJ nao encontrado.");
        return;
      }

      const currentParty = listByRole(role)[index];
      if (!currentParty) {
        return;
      }

      const basePatch: Partial<PartyDraft> = {
        cnpj: company.cnpj !== "" ? company.cnpj : formatCnpj(digits),
        razaoSocial:
          company.razaoSocial !== "" ? company.razaoSocial : currentParty.razaoSocial,
        endCep: company.endereco.cep !== "" ? formatCep(company.endereco.cep) : currentParty.endCep,
        endLogradouro:
          company.endereco.logradouro !== "" ? company.endereco.logradouro : currentParty.endLogradouro,
        endNumero: company.endereco.numero !== "" ? company.endereco.numero : currentParty.endNumero,
        endComplemento:
          company.endereco.complemento !== ""
            ? company.endereco.complemento
            : currentParty.endComplemento,
        endBairro: company.endereco.bairro !== "" ? company.endereco.bairro : currentParty.endBairro,
        endCidade: company.endereco.cidade !== "" ? company.endereco.cidade : currentParty.endCidade,
        endUf: company.endereco.uf !== "" ? company.endereco.uf : currentParty.endUf
      };
      patchParty(role, index, basePatch);

      const nextPartyAfterPatch = {
        ...currentParty,
        ...basePatch
      };
      const cepFromCnpj = onlyCepDigits(nextPartyAfterPatch.endCep);
      if (isCompleteCep(cepFromCnpj)) {
        const viaCepAddress = await lookupAddressByCep(cepFromCnpj);
        if (partyCnpjLookupRequestIds.get(lookupKey) !== requestId || !viaCepAddress) {
          return;
        }

        // Complementa somente os campos que vieram vazios da consulta de CNPJ.
        const cepPatch: Partial<PartyDraft> = {
          endCep: viaCepAddress.cep !== "" ? viaCepAddress.cep : nextPartyAfterPatch.endCep,
          endLogradouro:
            nextPartyAfterPatch.endLogradouro !== ""
              ? nextPartyAfterPatch.endLogradouro
              : viaCepAddress.logradouro,
          endComplemento:
            nextPartyAfterPatch.endComplemento !== ""
              ? nextPartyAfterPatch.endComplemento
              : viaCepAddress.complemento,
          endBairro:
            nextPartyAfterPatch.endBairro !== "" ? nextPartyAfterPatch.endBairro : viaCepAddress.bairro,
          endCidade:
            nextPartyAfterPatch.endCidade !== "" ? nextPartyAfterPatch.endCidade : viaCepAddress.cidade,
          endUf: nextPartyAfterPatch.endUf !== "" ? nextPartyAfterPatch.endUf : viaCepAddress.uf
        };
        patchParty(role, index, cepPatch);
      }

      partyLastFetchedCnpj.set(lookupKey, digits);
      setPartyCnpjLookupState(role, index, "success", "Dados da empresa preenchidos pelo CNPJ.");
    } catch (lookupError) {
      if (partyCnpjLookupRequestIds.get(lookupKey) !== requestId) {
        return;
      }
      const message =
        lookupError instanceof Error
          ? lookupError.message
          : "Nao foi possivel consultar o CNPJ agora.";
      setPartyCnpjLookupState(role, index, "error", message);
    }
  }

  function onPartyCnpjInput(role: PartyRole, index: number, value: string): void {
    const formatted = formatCnpj(value);
    updateParty(role, index, "cnpj", formatted);

    const digits = onlyCnpjDigits(formatted);
    if (!isCompleteCnpj(digits)) {
      setPartyCnpjLookupState(role, index, "idle", "");
      return;
    }

    const lookupKey = partyLookupKey(role, index);
    if (partyLastFetchedCnpj.get(lookupKey) === digits) {
      return;
    }
    void fillPartyFromCnpj(role, index, formatted);
  }

  type PartyCpfField = "cpf" | "conjCpf" | "repCpf";

  function onPartyCpfInput(role: PartyRole, index: number, key: PartyCpfField, value: string): void {
    const formatted = formatCpf(value);
    patchParty(role, index, { [key]: formatted } as Partial<PartyDraft>);
  }

  function onPartyCepInput(role: PartyRole, index: number, value: string): void {
    const formatted = formatCep(value);
    updateParty(role, index, "endCep", formatted);

    const digits = onlyCepDigits(formatted);
    if (!isCompleteCep(digits)) {
      setPartyCepLookupState(role, index, "idle", "");
      return;
    }

    void fillPartyAddressFromCep(role, index, formatted);
  }

  async function fillPartyAddressFromCep(role: PartyRole, index: number, cepValue: string): Promise<void> {
    const digits = onlyCepDigits(cepValue);
    if (!isCompleteCep(digits)) {
      return;
    }

    const lookupKey = partyLookupKey(role, index);
    const currentParty = listByRole(role)[index];
    if (!currentParty) {
      return;
    }

    // Evita consultas repetidas quando o endereco principal ja foi preenchido para o mesmo CEP.
    if (
      partyLastFetchedCep.get(lookupKey) === digits &&
      currentParty.endLogradouro.trim() !== "" &&
      currentParty.endBairro.trim() !== "" &&
      currentParty.endCidade.trim() !== "" &&
      currentParty.endUf.trim() !== ""
    ) {
      return;
    }

    const requestId = (partyCepLookupRequestIds.get(lookupKey) ?? 0) + 1;
    partyCepLookupRequestIds.set(lookupKey, requestId);
    setPartyCepLookupState(role, index, "loading", "Buscando endereco pelo CEP...");

    try {
      const result = await lookupAddressByCep(digits);
      if (partyCepLookupRequestIds.get(lookupKey) !== requestId) {
        return;
      }

      if (!result) {
        setPartyCepLookupState(role, index, "error", "CEP nao encontrado no ViaCEP.");
        return;
      }

      const partyAfterLookup = listByRole(role)[index];
      if (!partyAfterLookup) {
        return;
      }

      patchParty(role, index, {
        endCep: result.cep !== "" ? result.cep : formatCep(digits),
        endLogradouro: result.logradouro !== "" ? result.logradouro : partyAfterLookup.endLogradouro,
        endComplemento: result.complemento !== "" ? result.complemento : partyAfterLookup.endComplemento,
        endBairro: result.bairro !== "" ? result.bairro : partyAfterLookup.endBairro,
        endCidade: result.cidade !== "" ? result.cidade : partyAfterLookup.endCidade,
        endUf: result.uf !== "" ? result.uf : partyAfterLookup.endUf
      });

      partyLastFetchedCep.set(lookupKey, digits);
      setPartyCepLookupState(role, index, "success", "Endereco preenchido automaticamente pelo CEP.");
    } catch (lookupErr) {
      if (partyCepLookupRequestIds.get(lookupKey) !== requestId) {
        return;
      }
      setPartyCepLookupState(
        role,
        index,
        "error",
        lookupErr instanceof Error ? lookupErr.message : "Nao foi possivel consultar o CEP agora."
      );
    }
  }

  function partyAddressSectionTitle(party: PartyDraft): string {
    return isPartyPF(party) ? "Endereco" : "Endereco da empresa";
  }

  function partyAddressPreview(party: PartyDraft): string {
    const generated = buildAddressText({
      cep: party.endCep,
      logradouro: party.endLogradouro,
      numero: party.endNumero,
      complemento: party.endComplemento,
      bairro: party.endBairro,
      cidade: party.endCidade,
      uf: party.endUf
    });
    return generated !== "" ? generated : party.endTexto;
  }

  function isKnownPartyOption(value: string, options: readonly string[]): boolean {
    return options.includes(value.trim().toLowerCase());
  }

  function selectedNacionalidadeOption(value: string): string {
    const normalized = value.trim().toLowerCase();
    if (isKnownPartyOption(normalized, PARTY_NACIONALIDADE_OPTIONS)) {
      return normalized;
    }
    return PARTY_OTHER_OPTION;
  }

  function selectedRegimeBensOption(value: string): string {
    const normalized = value.trim().toLowerCase();
    if (isKnownPartyOption(normalized, PARTY_REGIME_BENS_OPTIONS)) {
      return normalized;
    }
    return PARTY_REGIME_OTHER_OPTION;
  }

  // Evita valor livre no select de comissao ao carregar rascunhos antigos.
  function selectedCommissionPayerOption(value: string): ((typeof COMMISSION_PAYER_OPTIONS)[number] | "") {
    const matched = COMMISSION_PAYER_OPTIONS.find((option) => option === value.trim());
    return matched ?? "";
  }

  function selectedCommissionPaymentMomentOption(
    value: string
  ): ((typeof COMMISSION_PAYMENT_MOMENT_OPTIONS)[number] | "") {
    const matched = COMMISSION_PAYMENT_MOMENT_OPTIONS.find((option) => option === value.trim());
    return matched ?? "";
  }

  function isKnownCommissionPaymentMoment(value: string): boolean {
    return selectedCommissionPaymentMomentOption(value) !== "";
  }

  function onPartyTipoChange(role: PartyRole, index: number, value: string): void {
    const tipo = value.trim() === "" ? "" : partyTypeOption(value);
    const lookupKey = partyLookupKey(role, index);
    partyLastFetchedCnpj.delete(lookupKey);
    partyCnpjLookupRequestIds.delete(lookupKey);
    partyLastFetchedCep.delete(lookupKey);
    partyCepLookupRequestIds.delete(lookupKey);
    // Limpa feedback antigo ao trocar PF/PJ para evitar mensagens fora de contexto.
    setPartyCnpjLookupState(role, index, "idle", "");
    setPartyCepLookupState(role, index, "idle", "");
    patchParty(role, index, { tipo });
  }

  // Quando seleciona uma opcao padrao, a nacionalidade principal recebe o valor.
  function onPartyNacionalidadeOptionChange(role: PartyRole, index: number, value: string): void {
    if (value === PARTY_OTHER_OPTION) {
      patchParty(role, index, {
        nacionalidade: "",
        nacionalidadeOutra: ""
      });
      return;
    }
    patchParty(role, index, {
      nacionalidade: value,
      nacionalidadeOutra: ""
    });
  }

  function onPartyNacionalidadeOutraInput(role: PartyRole, index: number, value: string): void {
    patchParty(role, index, {
      nacionalidade: value,
      nacionalidadeOutra: value
    });
  }

  function onPartyConjNacionalidadeOptionChange(role: PartyRole, index: number, value: string): void {
    if (value === PARTY_OTHER_OPTION) {
      patchParty(role, index, {
        conjNacionalidade: "",
        conjNacionalidadeOutra: ""
      });
      return;
    }
    patchParty(role, index, {
      conjNacionalidade: value,
      conjNacionalidadeOutra: ""
    });
  }

  function onPartyConjNacionalidadeOutraInput(role: PartyRole, index: number, value: string): void {
    patchParty(role, index, {
      conjNacionalidade: value,
      conjNacionalidadeOutra: value
    });
  }

  function onPartyEstadoCivilChange(role: PartyRole, index: number, value: string): void {
    const nextState: Partial<PartyDraft> = {
      estadoCivil: value
    };
    if (!isPartyEstadoCivilComConjuge(value)) {
      // Segue a regra do python para limpar bloco de conjuge/regime quando nao se aplica.
      nextState.regimeBens = "";
      nextState.regimeBensOutro = "";
      nextState.conjNome = "";
      nextState.conjNacionalidade = "";
      nextState.conjNacionalidadeOutra = "";
      nextState.conjProfissao = "";
      nextState.conjRg = "";
      nextState.conjCpf = "";
    }
    patchParty(role, index, nextState);
  }

  function onPartyRegimeBensChange(role: PartyRole, index: number, value: string): void {
    if (value === PARTY_REGIME_OTHER_OPTION) {
      patchParty(role, index, { regimeBens: "", regimeBensOutro: "" });
      return;
    }
    patchParty(role, index, {
      regimeBens: value,
      regimeBensOutro: ""
    });
  }

  function onPartyRegimeBensOutroInput(role: PartyRole, index: number, value: string): void {
    patchParty(role, index, {
      regimeBens: value,
      regimeBensOutro: value
    });
  }

  function handleCepInput(event: Event): void {
    const formatted = formatCep(inputValue(event));
    updateField("imovelCep", formatted);

    if (!isCompleteCep(formatted)) {
      cepLookupStatus = "idle";
      cepLookupMessage = "";
      return;
    }

    void fillAddressFromCep(formatted);
  }

  async function fillAddressFromCep(cepValue: string): Promise<void> {
    const digits = onlyCepDigits(cepValue);
    if (!isCompleteCep(digits)) {
      return;
    }

    // Evita buscar repetidamente o mesmo CEP quando os campos principais ja foram preenchidos.
    if (
      digits === lastFetchedCep &&
      draft.imovelLogradouro.trim() !== "" &&
      draft.imovelBairro.trim() !== "" &&
      draft.imovelCidade.trim() !== "" &&
      draft.imovelUf.trim() !== ""
    ) {
      return;
    }

    const requestId = ++cepLookupRequestId;
    cepLookupStatus = "loading";
    cepLookupMessage = "Buscando endereco pelo CEP...";

    try {
      const result = await lookupAddressByCep(digits);
      if (requestId !== cepLookupRequestId) {
        return;
      }

      if (!result) {
        cepLookupStatus = "error";
        cepLookupMessage = "CEP nao encontrado no ViaCEP.";
        return;
      }

      const nextDraft: ContractEditorDraft = {
        ...draft,
        imovelCep: result.cep !== "" ? result.cep : formatCep(digits),
        imovelLogradouro: result.logradouro !== "" ? result.logradouro : draft.imovelLogradouro,
        imovelComplemento: result.complemento !== "" ? result.complemento : draft.imovelComplemento,
        imovelBairro: result.bairro !== "" ? result.bairro : draft.imovelBairro,
        imovelCidade: result.cidade !== "" ? result.cidade : draft.imovelCidade,
        imovelUf: result.uf !== "" ? result.uf : draft.imovelUf
      };

      if (
        nextDraft.imovelUf.trim().toUpperCase() === "SP" &&
        nextDraft.imovelCidade.trim() !== "" &&
        nextDraft.imovelCidadeCartorio.trim() === ""
      ) {
        nextDraft.imovelCidadeCartorio = nextDraft.imovelCidade;
      }

      draft = nextDraft;
      lastFetchedCep = digits;
      cepLookupStatus = "success";
      cepLookupMessage = "Endereco preenchido automaticamente pelo CEP.";
    } catch (lookupErr) {
      if (requestId !== cepLookupRequestId) {
        return;
      }

      cepLookupStatus = "error";
      cepLookupMessage =
        lookupErr instanceof Error ? lookupErr.message : "Nao foi possivel consultar o CEP agora.";
    }
  }

  function addParty(role: PartyRole): void {
    const list = listByRole(role);
    const next = [...list, emptyPartyDraft(role, list.length + 1)];
    draft = { ...draft, [role]: next } as ContractEditorDraft;

    // Garante que novas linhas iniciem sem mensagens residuais de consultas anteriores.
    const newIndex = next.length - 1;
    setPartyCnpjLookupState(role, newIndex, "idle", "");
    setPartyCepLookupState(role, newIndex, "idle", "");
  }

  function removeParty(role: PartyRole, index: number): void {
    const lookupKey = partyLookupKey(role, index);
    partyCnpjLookupRequestIds.delete(lookupKey);
    partyLastFetchedCnpj.delete(lookupKey);
    partyCepLookupRequestIds.delete(lookupKey);
    partyLastFetchedCep.delete(lookupKey);

    const nextCnpjLookup = new Map(partyCnpjLookup);
    nextCnpjLookup.delete(lookupKey);
    partyCnpjLookup = nextCnpjLookup;

    const nextCepLookup = new Map(partyCepLookup);
    nextCepLookup.delete(lookupKey);
    partyCepLookup = nextCepLookup;

    const list = listByRole(role);
    const filtered = list.filter((_, i) => i !== index);
    const next = filtered.length > 0 ? filtered : [emptyPartyDraft(role, 1)];
    draft = { ...draft, [role]: next } as ContractEditorDraft;
  }

  function addClauseTag(clauseKey: string): void {
    const key = clauseKey.trim();
    if (key === "") {
      return;
    }
    if (draft.clausulasSelecionadas.some((item) => item.clauseKey === key)) {
      clauseSearch = "";
      return;
    }

    const match = availableClauses.find((item) => item.clause_key === key);

    draft = {
      ...draft,
      clausulasSelecionadas: [
        ...draft.clausulasSelecionadas,
        {
          clauseKey: key,
          title: match?.title ?? key,
          content: match?.content ?? "",
          index: ""
        }
      ]
    };
    clauseSearch = "";
  }

  function removeClauseTag(clauseKey: string): void {
    draft = {
      ...draft,
      clausulasSelecionadas: draft.clausulasSelecionadas.filter((item) => item.clauseKey !== clauseKey)
    };
  }

  function updateClauseTagIndex(clauseKey: string, value: string): void {
    draft = {
      ...draft,
      clausulasSelecionadas: draft.clausulasSelecionadas.map((item) =>
        item.clauseKey === clauseKey ? { ...item, index: value } : item
      )
    };
  }

  function addCustomClause(): void {
    draft = {
      ...draft,
      clausulasCustomizadas: [
        ...draft.clausulasCustomizadas,
        { title: "", content: "", index: "" }
      ]
    };
  }

  function updateCustomClause(index: number, key: "title" | "content" | "index", value: string): void {
    draft = {
      ...draft,
      clausulasCustomizadas: draft.clausulasCustomizadas.map((item, itemIndex) =>
        itemIndex === index ? { ...item, [key]: value } : item
      )
    };
  }

  function removeCustomClause(index: number): void {
    draft = {
      ...draft,
      clausulasCustomizadas: draft.clausulasCustomizadas.filter((_, itemIndex) => itemIndex !== index)
    };
  }

  function handleClauseSearchInput(event: Event): void {
    clauseSearch = inputValue(event);
  }

  function handleClauseSearchKeydown(event: KeyboardEvent): void {
    if (event.key !== "Enter") {
      return;
    }
    event.preventDefault();
    if (clauseSuggestions.length > 0) {
      addClauseTag(clauseSuggestions[0].clause_key);
    }
  }

  function getClauseSuggestions(
    query: string,
    clauses: ClauseTemplate[],
    selectedKeys: string[]
  ): ClauseTemplate[] {
    const active = clauses.filter((item) => item.is_active);
    const normalized = query.trim().toLowerCase();

    return active
      .filter((item) => !selectedKeys.includes(item.clause_key))
      .filter((item) => {
        if (normalized === "") {
          return true;
        }
        return (
          item.clause_key.toLowerCase().includes(normalized) ||
          item.title.toLowerCase().includes(normalized) ||
          item.content.toLowerCase().includes(normalized)
        );
      })
      .sort((a, b) => a.title.localeCompare(b.title))
      .slice(0, 8);
  }

  function inputValue(event: Event): string {
    return readInputValue(event);
  }

  function selectValue(event: Event): string {
    return readSelectValue(event);
  }

  function textareaValue(event: Event): string {
    return readTextareaValue(event);
  }

  function partyTypeValue(event: Event): string {
    // Centraliza a leitura para cobrir target SELECT e OPTION com o valor mais recente.
    return selectValue(event);
  }

  function isKnownPropertyType(value: string): boolean {
    return PROPERTY_TYPE_OPTIONS.includes(value as (typeof PROPERTY_TYPE_OPTIONS)[number]);
  }

  function buildContractText(contractPreview: ContractPreview): string {
    const fullText = (contractPreview.full_text ?? "").trim();
    if (fullText !== "") {
      return fullText;
    }
    return contractPreview.sections.join("\n\n").trim();
  }

  function buildContractParagraphs(title: string, text: string): Paragraph[] {
    const paragraphs: Paragraph[] = [];
    const normalizedTitle = title.trim();
    if (normalizedTitle !== "") {
      paragraphs.push(
        new Paragraph({
          heading: HeadingLevel.HEADING_1,
          text: normalizedTitle
        })
      );
      paragraphs.push(new Paragraph({ text: " " }));
    }

    for (const line of text.split(/\r?\n/)) {
      paragraphs.push(
        new Paragraph({
          text: line === "" ? " " : line
        })
      );
    }

    return paragraphs;
  }

  function triggerBlobDownload(blob: Blob, fileName: string): void {
    const blobUrl = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = blobUrl;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(blobUrl);
  }

  function buildDocxFilename(numero: string, tipo: string): string {
    const numberPart = sanitizeFilenamePart(numero) || "sem-numero";
    const typePart = sanitizeFilenamePart(tipo) || "contrato";
    return `contrato-${numberPart}-${typePart}.docx`;
  }

  function sanitizeFilenamePart(value: string): string {
    return value
      .trim()
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "")
      .slice(0, 60);
  }

  function hydrateDraftClauseMetadata(
    source: ContractEditorDraft,
    clauses: ClauseTemplate[]
  ): ContractEditorDraft {
    if (source.clausulasSelecionadas.length === 0 || clauses.length === 0) {
      return source;
    }

    const clausesByKey = new Map(clauses.map((item) => [item.clause_key, item]));
    let hasChanges = false;
    const selected = source.clausulasSelecionadas.map((item) => {
      const template = clausesByKey.get(item.clauseKey);
      if (!template) {
        return item;
      }

      const nextTitle = item.title.trim() === "" ? template.title : item.title;
      const nextContent = item.content.trim() === "" ? template.content : item.content;
      if (nextTitle === item.title && nextContent === item.content) {
        return item;
      }

      hasChanges = true;
      return { ...item, title: nextTitle, content: nextContent };
    });

    if (!hasChanges) {
      return source;
    }
    return { ...source, clausulasSelecionadas: selected };
  }
</script>

<section class="section-stack">
  <div class="panel">
    <div class="page-head">
      <h1>Editor Visual do Contrato</h1>
      <p>Edite os dados em blocos simples, sem precisar manipular JSON.</p>
    </div>

    {#if loading}
      <p>Carregando...</p>
    {:else if details}
      <div class="editor-meta">
        <article class="meta-card">
          <span class="meta-label">Numero</span>
          <strong>{details.contract.numero}</strong>
        </article>
        <article class="meta-card">
          <span class="meta-label">Tipo</span>
          <strong>{details.contract.tipo}</strong>
        </article>
        <article class="meta-card">
          <span class="meta-label">Status</span>
          <strong>{details.contract.status}</strong>
        </article>
        <article class="meta-card">
          <span class="meta-label">Versao em edicao</span>
          <strong>{selectedVersion > 0 ? `v${selectedVersion}` : "Rascunho local"}</strong>
        </article>
      </div>

      <div class="editor-toolbar">
        <button
          class="btn ghost"
          disabled={previewing || hasPartyConditionalBlockers}
          on:click={() => refreshPreview(true)}
        >
          {previewing ? "Gerando previa..." : "Atualizar previa"}
        </button>
        <button class="btn ghost" on:click={restoreLatestVersion}>Restaurar ultima versao</button>
        <button class="btn ghost" on:click={load}>Recarregar contrato</button>
      </div>

      <div class="editor-form">

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Imovel</h3>
            <p>Dados essenciais do objeto contratual.</p>
          </div>

          <div class="property-layout">
            <div class="property-column">
              <h4>Endereco do imovel</h4>

              <div class="grid cols-2">
                {#each propertyAddressFields as field}
                  <div class="field">
                    <label for={`property_${field.key}`}>{field.label}</label>
                    {#if field.key === "imovelCep"}
                      <input
                        id={`property_${field.key}`}
                        value={draft[field.key]}
                        placeholder={field.placeholder}
                        on:input={handleCepInput}
                      />
                      {#if cepLookupStatus === "loading"}
                        <small class="field-feedback info">{cepLookupMessage}</small>
                      {:else if cepLookupStatus === "error"}
                        <small class="field-feedback error">{cepLookupMessage}</small>
                      {:else if cepLookupStatus === "success"}
                        <small class="field-feedback info">{cepLookupMessage}</small>
                      {/if}
                    {:else}
                      <input
                        id={`property_${field.key}`}
                        value={draft[field.key]}
                        placeholder={field.placeholder}
                        on:input={(event) => updateField(field.key, inputValue(event))}
                      />
                    {/if}
                  </div>
                {/each}
              </div>

              <div class="field">
                <label for="property_imovelEndereco">Endereco completo (gerado)</label>
                <textarea id="property_imovelEndereco" value={propertyAddressPreview} rows="3" disabled></textarea>
              </div>
            </div>

            <div class="property-column">
              <h4>Identificacao</h4>

              <div class="field">
                <label for="property_imovelTipo">Tipo do imovel</label>
                <select
                  id="property_imovelTipo"
                  value={draft.imovelTipo}
                  on:change={(event) => updateField("imovelTipo", selectValue(event))}
                >
                  <option value="">Selecione</option>
                  {#if draft.imovelTipo.trim() !== "" && !isKnownPropertyType(draft.imovelTipo)}
                    <option value={draft.imovelTipo}>{draft.imovelTipo}</option>
                  {/if}
                  {#each PROPERTY_TYPE_OPTIONS as option}
                    <option value={option}>{option}</option>
                  {/each}
                </select>
              </div>

              <div class="grid cols-2">
                {#each propertyIdentificationFields as field}
                  <div class="field">
                    <label for={`property_${field.key}`}>{field.label}</label>
                    <input
                      id={`property_${field.key}`}
                      value={draft[field.key]}
                      placeholder={field.placeholder}
                      on:input={(event) => updateField(field.key, inputValue(event))}
                    />
                  </div>
                {/each}
              </div>

              <div class="editor-subsection">
                <h4>Informacoes adicionais</h4>

                <div class="property-toggle-grid">
                  {#each propertyToggleFields as field}
                    <div class="field">
                      <span class="field-static-label">{field.label}</span>
                      <div class="radio-inline">
                        {#each YES_NO_OPTIONS as option}
                          <label>
                            <input
                              type="radio"
                              name={`property_${field.key}`}
                              value={option}
                              checked={draft[field.key] === option}
                              on:change={(event) => updateField(field.key, inputValue(event))}
                            />
                            <span>{option}</span>
                          </label>
                        {/each}
                      </div>
                    </div>
                  {/each}
                </div>

                {#if draft.imovelAlugado === "SIM"}
                  <div class="field">
                    <label for="property_imovelLocacao">
                      O inquilino vai desocupar o imovel ou a parte compradora vai assumir a locacao?
                    </label>
                    <textarea
                      id="property_imovelLocacao"
                      value={draft.imovelLocacao}
                      rows="4"
                      on:input={(event) => updateField("imovelLocacao", textareaValue(event))}
                    ></textarea>
                  </div>
                {/if}

                {#if draft.imovelFicaraBens === "SIM"}
                  <div class="field">
                    <label for="property_imovelBens">
                      O que ficara no imovel? (indicar somente os bens)
                    </label>
                    <textarea
                      id="property_imovelBens"
                      value={draft.imovelBens}
                      rows="4"
                      on:input={(event) => updateField("imovelBens", textareaValue(event))}
                    ></textarea>
                  </div>
                {/if}
              </div>
            </div>
          </div>

          <div class="editor-subsection">
            <h4>Descricao do imovel na matricula</h4>

            {#if hideMatriculaDescription}
              <div class="notice info">
                Regra aplicada: nao lancar descricao do imovel para tipo com matricula em area maior.
              </div>
            {:else}
              <div class="field">
                <textarea
                  id="property_imovelDescricaoMatricula"
                  value={draft.imovelDescricaoMatricula}
                  rows="6"
                  on:input={(event) => updateField("imovelDescricaoMatricula", textareaValue(event))}
                ></textarea>
              </div>
            {/if}
          </div>
        </section>
        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Partes do contrato</h3>
            <p>Adicione vendedores/cedentes e compradores/cessionarios.</p>
          </div>

          <div class="party-columns">
            {#each PARTY_SECTIONS as section}
              <div class="party-column">
                <div class="party-column-head">
                  <h4>{section.title}</h4>
                  <button class="btn ghost btn-sm" on:click={() => addParty(section.role)}>
                    Adicionar
                  </button>
                </div>

                {#each listByRoleForRender(draft, section.role) as party, index}
                  <div class="party-card">
                    <div class="field">
                      <label for={`${section.idPrefix}_tipo_${index}`}>Tipo da parte</label>
                      <!-- Atualiza no input e no change para alternar PF/PJ sem atraso visual. -->
                      <select
                        id={`${section.idPrefix}_tipo_${index}`}
                        value={selectedPartyTypeOption(party.tipo)}
                        on:input={(event) =>
                          onPartyTipoChange(section.role, index, partyTypeValue(event))}
                        on:change={(event) =>
                          onPartyTipoChange(section.role, index, partyTypeValue(event))}
                      >
                        <option value="">Selecione o tipo da parte</option>
                        {#each PARTY_TYPE_OPTIONS as option}
                          <option value={option}>{option}</option>
                        {/each}
                      </select>
                    </div>

                    {#if !hasPartyTypeSelected(party)}
                      <p class="hint">
                        Selecione o tipo da parte para exibir os campos do formulario.
                      </p>
                    {:else if isPartyPF(party)}
                      <div class="grid cols-2">
                        <div class="field">
                          <label for={`${section.idPrefix}_nome_${index}`}>Nome completo</label>
                          <input
                            id={`${section.idPrefix}_nome_${index}`}
                            value={party.nome}
                            on:input={(event) => updateParty(section.role, index, "nome", inputValue(event))}
                          />
                        </div>

                        <div class="field">
                          <label for={`${section.idPrefix}_nat_${index}`}>Nacionalidade</label>
                          <select
                            id={`${section.idPrefix}_nat_${index}`}
                            value={selectedNacionalidadeOption(party.nacionalidade)}
                            on:change={(event) =>
                              onPartyNacionalidadeOptionChange(section.role, index, selectValue(event))}
                          >
                            {#each PARTY_NACIONALIDADE_OPTIONS as option}
                              <option value={option}>{option}</option>
                            {/each}
                          </select>
                        </div>

                        {#if selectedNacionalidadeOption(party.nacionalidade) === PARTY_OTHER_OPTION}
                          <div class="field span-all">
                            <label for={`${section.idPrefix}_nat_outro_${index}`}>Escreva a nacionalidade</label>
                            <input
                              id={`${section.idPrefix}_nat_outro_${index}`}
                              value={party.nacionalidadeOutra || party.nacionalidade}
                              on:input={(event) =>
                                onPartyNacionalidadeOutraInput(section.role, index, inputValue(event))}
                            />
                          </div>
                        {/if}

                        <div class="field">
                          <label for={`${section.idPrefix}_rg_${index}`}>RG n.o</label>
                          <input
                            id={`${section.idPrefix}_rg_${index}`}
                            value={party.rg}
                            on:input={(event) => updateParty(section.role, index, "rg", inputValue(event))}
                          />
                        </div>

                        <div class="field">
                          <label for={`${section.idPrefix}_cpf_${index}`}>CPF n.o</label>
                          <input
                            id={`${section.idPrefix}_cpf_${index}`}
                            value={party.cpf}
                            on:input={(event) =>
                              onPartyCpfInput(section.role, index, "cpf", inputValue(event))}
                          />
                        </div>

                        <div class="field">
                          <label for={`${section.idPrefix}_prof_${index}`}>Profissao</label>
                          <input
                            id={`${section.idPrefix}_prof_${index}`}
                            value={party.profissao}
                            on:input={(event) => updateParty(section.role, index, "profissao", inputValue(event))}
                          />
                        </div>

                        <div class="field">
                          <label for={`${section.idPrefix}_civil_${index}`}>Estado civil</label>
                          <select
                            id={`${section.idPrefix}_civil_${index}`}
                            value={party.estadoCivil}
                            on:change={(event) =>
                              onPartyEstadoCivilChange(section.role, index, selectValue(event))}
                          >
                            {#each PARTY_ESTADO_CIVIL_OPTIONS as option}
                              <option value={option}>{option}</option>
                            {/each}
                          </select>
                        </div>
                      </div>

                      {#if shouldShowConjuge(party)}
                        <div class="editor-subsection">
                          <h5>Conjuge / companheiro(a)</h5>

                          <div class="grid cols-2">
                            <div class="field">
                              <label for={`${section.idPrefix}_regime_${index}`}>Regime de bens</label>
                              <select
                                id={`${section.idPrefix}_regime_${index}`}
                                value={selectedRegimeBensOption(party.regimeBens)}
                                on:change={(event) =>
                                  onPartyRegimeBensChange(section.role, index, selectValue(event))}
                              >
                                {#each PARTY_REGIME_BENS_OPTIONS as option}
                                  <option value={option}>{option}</option>
                                {/each}
                              </select>
                            </div>

                            {#if selectedRegimeBensOption(party.regimeBens) === PARTY_REGIME_OTHER_OPTION}
                              <div class="field">
                                <label for={`${section.idPrefix}_regime_outro_${index}`}>
                                  Escreva o regime de bens
                                </label>
                                <input
                                  id={`${section.idPrefix}_regime_outro_${index}`}
                                  value={party.regimeBensOutro || party.regimeBens}
                                  on:input={(event) =>
                                    onPartyRegimeBensOutroInput(section.role, index, inputValue(event))}
                                />
                              </div>
                            {/if}

                            <div class="field">
                              <label for={`${section.idPrefix}_conj_nome_${index}`}>
                                {partyConjugeNomeLabel(party)}
                              </label>
                              <input
                                id={`${section.idPrefix}_conj_nome_${index}`}
                                value={party.conjNome}
                                on:input={(event) => updateParty(section.role, index, "conjNome", inputValue(event))}
                              />
                              {#if isPartyConjugeObrigatorioIncompleto(party)}
                                <small class="field-feedback error">
                                  Para CASADO(A) ou UNIAO ESTAVEL, o nome do conjuge/companheiro(a) e obrigatorio.
                                </small>
                              {/if}
                            </div>

                            <div class="field">
                              <label for={`${section.idPrefix}_conj_nat_${index}`}>Nacionalidade do conjuge</label>
                              <select
                                id={`${section.idPrefix}_conj_nat_${index}`}
                                value={selectedNacionalidadeOption(party.conjNacionalidade)}
                                on:change={(event) =>
                                  onPartyConjNacionalidadeOptionChange(section.role, index, selectValue(event))}
                              >
                                {#each PARTY_NACIONALIDADE_OPTIONS as option}
                                  <option value={option}>{option}</option>
                                {/each}
                              </select>
                            </div>

                            {#if selectedNacionalidadeOption(party.conjNacionalidade) === PARTY_OTHER_OPTION}
                              <div class="field span-all">
                                <label for={`${section.idPrefix}_conj_nat_outro_${index}`}>
                                  Escreva a nacionalidade do conjuge
                                </label>
                                <input
                                  id={`${section.idPrefix}_conj_nat_outro_${index}`}
                                  value={party.conjNacionalidadeOutra || party.conjNacionalidade}
                                  on:input={(event) =>
                                    onPartyConjNacionalidadeOutraInput(section.role, index, inputValue(event))}
                                />
                              </div>
                            {/if}

                            <div class="field">
                              <label for={`${section.idPrefix}_conj_prof_${index}`}>Profissao do conjuge</label>
                              <input
                                id={`${section.idPrefix}_conj_prof_${index}`}
                                value={party.conjProfissao}
                                on:input={(event) =>
                                  updateParty(section.role, index, "conjProfissao", inputValue(event))}
                              />
                            </div>

                            <div class="field">
                              <label for={`${section.idPrefix}_conj_rg_${index}`}>RG do conjuge</label>
                              <input
                                id={`${section.idPrefix}_conj_rg_${index}`}
                                value={party.conjRg}
                                on:input={(event) => updateParty(section.role, index, "conjRg", inputValue(event))}
                              />
                            </div>

                            <div class="field">
                              <label for={`${section.idPrefix}_conj_cpf_${index}`}>CPF do conjuge</label>
                              <input
                                id={`${section.idPrefix}_conj_cpf_${index}`}
                                value={party.conjCpf}
                                on:input={(event) =>
                                  onPartyCpfInput(section.role, index, "conjCpf", inputValue(event))}
                              />
                            </div>
                          </div>
                        </div>
                      {/if}
                    {:else}
                      <div class="grid cols-2">
                        <div class="field">
                          <label for={`${section.idPrefix}_cnpj_${index}`}>CNPJ</label>
                          <input
                            id={`${section.idPrefix}_cnpj_${index}`}
                            value={party.cnpj}
                            placeholder="00.000.000/0000-00"
                            on:input={(event) => onPartyCnpjInput(section.role, index, inputValue(event))}
                          />
                          {#if getPartyCnpjLookupState(section.role, index).status === "loading"}
                            <small class="field-feedback info">
                              {getPartyCnpjLookupState(section.role, index).message}
                            </small>
                          {:else if getPartyCnpjLookupState(section.role, index).status === "error"}
                            <small class="field-feedback error">
                              {getPartyCnpjLookupState(section.role, index).message}
                            </small>
                          {:else if getPartyCnpjLookupState(section.role, index).status === "success"}
                            <small class="field-feedback info">
                              {getPartyCnpjLookupState(section.role, index).message}
                            </small>
                          {/if}
                        </div>

                        <div class="field">
                          <label for={`${section.idPrefix}_razao_${index}`}>Razao social</label>
                          <input
                            id={`${section.idPrefix}_razao_${index}`}
                            value={party.razaoSocial}
                            placeholder="Preenchida automaticamente pelo CNPJ"
                            disabled
                          />
                        </div>

                        <div class="field span-all">
                          <h5>Representante legal (quem assina)</h5>
                        </div>

                        <div class="field">
                          <label for={`${section.idPrefix}_rep_nome_${index}`}>Nome do representante</label>
                          <input
                            id={`${section.idPrefix}_rep_nome_${index}`}
                            value={party.repNome}
                            on:input={(event) => updateParty(section.role, index, "repNome", inputValue(event))}
                          />
                        </div>

                        <div class="field">
                          <label for={`${section.idPrefix}_rep_cpf_${index}`}>CPF do representante</label>
                          <input
                            id={`${section.idPrefix}_rep_cpf_${index}`}
                            value={party.repCpf}
                            on:input={(event) =>
                              onPartyCpfInput(section.role, index, "repCpf", inputValue(event))}
                          />
                        </div>
                      </div>
                    {/if}

                    {#if hasPartyTypeSelected(party)}
                      <div class="editor-subsection">
                        <h5>{partyAddressSectionTitle(party)}</h5>
                        <div class="grid cols-2">
                          <div class="field">
                            <label for={`${section.idPrefix}_end_cep_${index}`}>CEP</label>
                            <input
                              id={`${section.idPrefix}_end_cep_${index}`}
                              value={party.endCep}
                              on:input={(event) => onPartyCepInput(section.role, index, inputValue(event))}
                            />
                            {#if getPartyCepLookupState(section.role, index).status === "loading"}
                              <small class="field-feedback info">
                                {getPartyCepLookupState(section.role, index).message}
                              </small>
                            {:else if getPartyCepLookupState(section.role, index).status === "error"}
                              <small class="field-feedback error">
                                {getPartyCepLookupState(section.role, index).message}
                              </small>
                            {:else if getPartyCepLookupState(section.role, index).status === "success"}
                              <small class="field-feedback info">
                                {getPartyCepLookupState(section.role, index).message}
                              </small>
                            {/if}
                          </div>
                          <div class="field">
                            <label for={`${section.idPrefix}_end_logradouro_${index}`}>Logradouro</label>
                            <input
                              id={`${section.idPrefix}_end_logradouro_${index}`}
                              value={party.endLogradouro}
                              on:input={(event) =>
                                updateParty(section.role, index, "endLogradouro", inputValue(event))}
                            />
                          </div>
                          <div class="field">
                            <label for={`${section.idPrefix}_end_numero_${index}`}>Numero</label>
                            <input
                              id={`${section.idPrefix}_end_numero_${index}`}
                              value={party.endNumero}
                              on:input={(event) => updateParty(section.role, index, "endNumero", inputValue(event))}
                            />
                          </div>
                          <div class="field">
                            <label for={`${section.idPrefix}_end_complemento_${index}`}>Complemento</label>
                            <input
                              id={`${section.idPrefix}_end_complemento_${index}`}
                              value={party.endComplemento}
                              on:input={(event) =>
                                updateParty(section.role, index, "endComplemento", inputValue(event))}
                            />
                          </div>
                          <div class="field">
                            <label for={`${section.idPrefix}_end_bairro_${index}`}>Bairro</label>
                            <input
                              id={`${section.idPrefix}_end_bairro_${index}`}
                              value={party.endBairro}
                              on:input={(event) => updateParty(section.role, index, "endBairro", inputValue(event))}
                            />
                          </div>
                          <div class="field">
                            <label for={`${section.idPrefix}_end_cidade_${index}`}>Cidade</label>
                            <input
                              id={`${section.idPrefix}_end_cidade_${index}`}
                              value={party.endCidade}
                              on:input={(event) => updateParty(section.role, index, "endCidade", inputValue(event))}
                            />
                          </div>
                          <div class="field">
                            <label for={`${section.idPrefix}_end_uf_${index}`}>UF</label>
                            <input
                              id={`${section.idPrefix}_end_uf_${index}`}
                              value={party.endUf}
                              on:input={(event) => updateParty(section.role, index, "endUf", inputValue(event))}
                            />
                          </div>
                          <div class="field span-all">
                            <label for={`${section.idPrefix}_end_texto_${index}`}>Endereco completo (gerado)</label>
                            <textarea
                              id={`${section.idPrefix}_end_texto_${index}`}
                              value={partyAddressPreview(party)}
                              rows="3"
                              disabled
                            ></textarea>
                          </div>
                        </div>
                      </div>
                    {/if}

                    <div class="inline-actions">
                      <button
                        class="btn ghost btn-sm"
                        disabled={listByRoleForRender(draft, section.role).length === 1}
                        on:click={() => removeParty(section.role, index)}
                      >
                        Remover
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/each}
          </div>

          {#if hasPartyConditionalBlockers}
            <div class="notice error">
              {partyConditionalErrors[0]}
            </div>
          {/if}
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Pagamento</h3>
            <p>Preencha somente os valores que se aplicam ao contrato.</p>
          </div>

          <div class="grid cols-3">
            {#each paymentFields as field}
              <div class="field">
                <label for={`payment_${field.key}`}>{field.label}</label>
                <input
                  id={`payment_${field.key}`}
                  value={draft[field.key]}
                  placeholder={field.placeholder}
                  on:input={(event) => updatePaymentField(field.key, inputValue(event))}
                  on:focus={() => preparePaymentFieldForEdit(field.key)}
                  on:blur={() => formatPaymentFieldOnBlur(field.key)}
                />
              </div>
            {/each}
          </div>

          <div class="payment-summary">
            <div class="field">
              <label for="payment_breakdown_total">Soma dos valores preenchidos</label>
              <input id="payment_breakdown_total" value={paymentBreakdownTotalFormatted} readonly />
            </div>
            <div class={`notice ${paymentBalanceNoticeType}`}>{paymentBalanceMessage}</div>
          </div>
        </section>

        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Comissao e intermediacao</h3>
          </div>

          <div class="grid cols-3">
            <div class="field">
              <label for="commission_quemPagaComissao">Quem paga comissao</label>
              <select
                id="commission_quemPagaComissao"
                value={selectedCommissionPayerOption(draft.quemPagaComissao)}
                on:change={(event) => updateField("quemPagaComissao", selectValue(event))}
              >
                <option value="">Selecione quem paga a comissao</option>
                {#each COMMISSION_PAYER_OPTIONS as option}
                  <option value={option}>{option}</option>
                {/each}
              </select>
            </div>

            <div class="field">
              <label for="commission_valorComissao">Valor da comissao</label>
              <input
                id="commission_valorComissao"
                value={draft.valorComissao}
                placeholder="Ex.: 6% ou R$ 18.000,00"
                on:input={(event) => updateField("valorComissao", inputValue(event))}
              />
            </div>

            <div class="field">
              <label for="commission_momentoPagto">Momento do pagamento</label>
              <select
                id="commission_momentoPagto"
                value={selectedCommissionPaymentMomentOption(draft.momentoPagto)}
                on:change={(event) => updateField("momentoPagto", selectValue(event))}
              >
                <option value="">Selecione o momento do pagamento</option>
                <!-- Preserva valor legado caso venha diferente das opcoes atuais. -->
                {#if draft.momentoPagto.trim() !== "" && !isKnownCommissionPaymentMoment(draft.momentoPagto)}
                  <option value={draft.momentoPagto}>{draft.momentoPagto}</option>
                {/if}
                {#each COMMISSION_PAYMENT_MOMENT_OPTIONS as option}
                  <option value={option}>{option}</option>
                {/each}
              </select>
            </div>
          </div>
        </section>
        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Entrega de chaves</h3>
            <p>Selecione uma regra padrao ou escreva o texto personalizado.</p>
          </div>

          <div class="grid cols-2">
            <div class="field">
              <label for="entrega_chaves">Regra principal</label>
              <select
                id="entrega_chaves"
                value={draft.entregaChaves}
                on:change={(event) => updateField("entregaChaves", selectValue(event))}
              >
                <option value="">Selecione uma opcao</option>
                {#each DELIVERY_OPTIONS as option}
                  <option value={option}>{option}</option>
                {/each}
              </select>
            </div>
          </div>

          {#if draft.entregaChaves === "Escrever no contrato"}
            <div class="field">
              <label for="entrega_chaves_texto">Texto da clausula</label>
              <textarea
                id="entrega_chaves_texto"
                value={draft.entregaChavesTexto}
                placeholder="Descreva como e quando as chaves serao entregues."
                on:input={(event) => updateField("entregaChavesTexto", textareaValue(event))}
              ></textarea>
            </div>
          {/if}

        </section>


        <section class="editor-section">
          <div class="editor-section-head">
            <h3>Clausulas adicionais do contrato</h3>
            <p>Busque clausulas salvas e adicione em formato de tags.</p>
          </div>

          <div class="field">
            <label for="clause_search">Adicionar clausula</label>
            <input
              id="clause_search"
              value={clauseSearch}
              placeholder="Digite titulo, chave ou trecho da clausula"
              on:input={handleClauseSearchInput}
              on:keydown={handleClauseSearchKeydown}
            />
          </div>

          {#if clauseLoadError}
            <div class="notice info">{clauseLoadError}</div>
          {/if}

          {#if clauseSearch.trim() !== ""}
            <div class="clause-suggestion-list">
              {#if clauseSuggestions.length === 0}
                <p class="hint">Nenhuma clausula ativa encontrada para "{clauseSearch}".</p>
              {:else}
                {#each clauseSuggestions as clause}
                  <button
                    class="clause-suggestion-item"
                    type="button"
                    on:click={() => addClauseTag(clause.clause_key)}
                  >
                    <strong>{clause.title}</strong>
                    <small>{clause.clause_key}</small>
                  </button>
                {/each}
              {/if}
            </div>
          {/if}

          <div class="clause-tags-area">
            {#if draft.clausulasSelecionadas.length === 0}
              <p class="hint">Nenhuma clausula selecionada.</p>
            {:else}
              <div class="linked-clause-list">
                {#each draft.clausulasSelecionadas as clause}
                  <div class="linked-clause-card">
                    <div class="linked-clause-title">
                      <strong>{clause.title || clause.clauseKey}</strong>
                      <small>{clause.clauseKey}</small>
                    </div>
                    <div class="field">
                      <label for={`clause_index_${clause.clauseKey}`}>Indice no contrato</label>
                      <input
                        id={`clause_index_${clause.clauseKey}`}
                        value={clause.index}
                        placeholder="Ex.: 1.1.2"
                        on:input={(event) => updateClauseTagIndex(clause.clauseKey, inputValue(event))}
                      />
                    </div>
                    <div class="inline-actions">
                      <button
                        type="button"
                        class="btn ghost btn-sm"
                        aria-label={`Remover clausula ${clause.clauseKey}`}
                        on:click={() => removeClauseTag(clause.clauseKey)}
                      >
                        Remover
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

          <div class="editor-subsection">
            <div class="party-column-head">
              <h4>Clausulas customizadas vinculadas</h4>
              <button class="btn ghost btn-sm" on:click={addCustomClause}>Adicionar clausula customizada</button>
            </div>

            {#if draft.clausulasCustomizadas.length === 0}
              <p class="hint">Nenhuma clausula customizada vinculada.</p>
            {:else}
              <div class="delivery-clause-list">
                {#each draft.clausulasCustomizadas as customClause, customIndex}
                  <div class="delivery-clause-card">
                    <div class="grid cols-2">
                      <div class="field">
                        <label for={`custom_clause_index_${customIndex}`}>Indice no contrato</label>
                        <input
                          id={`custom_clause_index_${customIndex}`}
                          value={customClause.index}
                          placeholder="Ex.: 1.1.2"
                          on:input={(event) =>
                            updateCustomClause(customIndex, "index", inputValue(event))}
                        />
                      </div>
                      <div class="field">
                        <label for={`custom_clause_title_${customIndex}`}>Titulo da clausula</label>
                        <input
                          id={`custom_clause_title_${customIndex}`}
                          value={customClause.title}
                          placeholder="Ex.: Clausula X"
                          on:input={(event) =>
                            updateCustomClause(customIndex, "title", inputValue(event))}
                        />
                      </div>
                      <div class="field span-all">
                        <label for={`custom_clause_content_${customIndex}`}>Conteudo da clausula</label>
                        <textarea
                          id={`custom_clause_content_${customIndex}`}
                          value={customClause.content}
                          placeholder="Texto completo da clausula customizada."
                          on:input={(event) =>
                            updateCustomClause(customIndex, "content", textareaValue(event))}
                        ></textarea>
                      </div>
                    </div>

                    <div class="inline-actions">
                      <button class="btn ghost btn-sm" on:click={() => removeCustomClause(customIndex)}>
                        Remover
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

        </section>

      </div>

      <div class="actions">
        <button class="btn primary" disabled={saving || hasPartyConditionalBlockers} on:click={saveNewVersion}>
          {saving ? "Salvando..." : "Salvar nova versao"}
        </button>
        <button
          class="btn ghost"
          disabled={previewing || hasPartyConditionalBlockers}
          on:click={() => refreshPreview(true)}
        >
          {previewing ? "Gerando previa..." : "Gerar previa agora"}
        </button>
        <button
          class="btn ghost"
          disabled={downloadingDocx || previewing || hasPartyConditionalBlockers}
          on:click={downloadContractDocx}
        >
          {downloadingDocx ? "Gerando DOCX..." : "Baixar DOCX"}
        </button>
      </div>

      {#if success}
        <div class="notice success">{success}</div>
      {/if}
      {#if error}
        <div class="notice error">{error}</div>
      {/if}
    {/if}
  </div>

  <div class="panel">
    <h2>Previa do Contrato</h2>
    {#if previewing}
      <p>Atualizando previa...</p>
    {/if}
    {#if preview}
      <h3>{preview.title}</h3>
      {#if preview.full_text}
        <article class="full-contract-preview">{preview.full_text}</article>
      {:else}
        <ol class="list-tight">
          {#each preview.sections as section}
            <li>{section}</li>
          {/each}
        </ol>
      {/if}
    {:else}
      <p>Sem previa disponivel.</p>
    {/if}
  </div>

  <div class="panel">
    <h2>Historico de versoes</h2>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Versao</th>
            <th>Criada em</th>
            <th>Acoes</th>
          </tr>
        </thead>
        <tbody>
          {#if details && details.versions.length > 0}
            {#each details.versions as version}
              <tr class:active-version={version.version_number === selectedVersion}>
                <td>v{version.version_number}</td>
                <td>{new Date(version.created_at).toLocaleString()}</td>
                <td>
                  <button class="btn ghost btn-sm" on:click={() => loadVersion(version)}>
                    Carregar
                  </button>
                </td>
              </tr>
            {/each}
          {:else}
            <tr><td colspan="3">Nenhuma versao</td></tr>
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</section>

<style>
  .editor-meta {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 10px;
    margin-bottom: 12px;
  }

  .meta-card {
    border: 1px solid #d3e0f0;
    border-radius: 12px;
    background: rgba(245, 250, 255, 0.9);
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .meta-label {
    color: #445f81;
    font-size: 0.8rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .editor-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 14px;
  }

  .editor-form {
    display: grid;
    gap: 12px;
  }

  .editor-section {
    border: 1px solid #d4e2f4;
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.55);
    padding: 14px;
    display: grid;
    gap: 12px;
  }

  .editor-section-head {
    display: grid;
    gap: 4px;
  }

  .editor-section-head h3 {
    margin: 0;
  }

  .editor-section-head p {
    margin: 0;
    color: #4a6281;
  }

  .party-columns {
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  }

  .party-column {
    display: grid;
    gap: 10px;
  }

  .party-column-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .party-column-head h4 {
    margin: 0;
    font-size: 1rem;
  }

  .party-card {
    border: 1px solid #d7e5f6;
    border-radius: 12px;
    background: #f9fbff;
    padding: 10px;
    display: grid;
    gap: 10px;
  }

  .property-layout {
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  }

  .property-column {
    display: grid;
    gap: 10px;
  }

  .property-column h4 {
    margin: 0;
    font-size: 1rem;
  }

  .property-toggle-grid {
    display: grid;
    gap: 10px;
    grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  }

  .radio-inline {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .field-static-label {
    font-weight: 700;
    font-size: 0.92rem;
    color: #304b6c;
  }

  .field-feedback {
    display: block;
    margin-top: 4px;
    font-size: 0.82rem;
    font-weight: 600;
  }

  .field-feedback.info {
    color: #1d4ed8;
  }

  .field-feedback.error {
    color: #b91c1c;
  }

  .radio-inline label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-weight: 600;
    color: #304b6c;
  }

  .radio-inline input {
    width: auto;
    min-width: auto;
    max-width: none;
    margin: 0;
  }

  .inline-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 8px;
  }

  .payment-summary {
    border-top: 1px solid #dce8f6;
    padding-top: 10px;
    display: grid;
    gap: 8px;
  }

  .clause-suggestion-list {
    border: 1px solid #d8e5f6;
    border-radius: 12px;
    background: #f8fbff;
    display: grid;
    gap: 6px;
    padding: 8px;
    max-height: 260px;
    overflow-y: auto;
  }

  .clause-suggestion-item {
    border: 1px solid #d0dff2;
    border-radius: 10px;
    padding: 8px 10px;
    background: #fff;
    color: inherit;
    text-align: left;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .clause-suggestion-item:hover {
    border-color: #22a18f;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
  }

  .clause-suggestion-item small {
    color: #4f6686;
    font-size: 0.8rem;
  }

  .clause-tags-area {
    border-top: 1px solid #dce8f6;
    padding-top: 10px;
  }

  .linked-clause-list {
    display: grid;
    gap: 8px;
  }

  .linked-clause-card {
    border: 1px solid #d7e5f6;
    border-radius: 12px;
    background: #f8fbff;
    padding: 10px;
    display: grid;
    gap: 8px;
  }

  .linked-clause-title {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .linked-clause-title small {
    color: #4f6686;
    font-size: 0.8rem;
  }

  .linked-clause-card .inline-actions {
    margin-top: 0;
  }

  .editor-subsection {
    border-top: 1px solid #dce8f6;
    padding-top: 12px;
    display: grid;
    gap: 10px;
  }

  .editor-subsection h5 {
    margin: 0;
    font-size: 0.96rem;
  }

  .delivery-clause-list {
    display: grid;
    gap: 10px;
  }

  .delivery-clause-card {
    border: 1px solid #d9e6f6;
    border-radius: 12px;
    background: #f8fbff;
    padding: 10px;
  }

  .hint {
    margin: 0;
    color: #5a6f8c;
    font-size: 0.92rem;
  }

  .full-contract-preview {
    margin-top: 10px;
    border: 1px solid #d5e3f5;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.92);
    padding: 14px;
    white-space: pre-wrap;
    line-height: 1.6;
    font-size: 0.95rem;
  }

  .btn-sm {
    padding: 7px 10px;
    font-size: 0.84rem;
  }

  .active-version {
    background: rgba(15, 118, 110, 0.12);
  }

  @media (max-width: 760px) {
    .editor-section {
      padding: 12px;
    }

    .party-column-head {
      flex-wrap: wrap;
    }

  }
</style>
