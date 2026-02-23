export interface CompanyAddressData {
  cep: string;
  logradouro: string;
  numero: string;
  complemento: string;
  bairro: string;
  cidade: string;
  uf: string;
}

export interface CompanyByCnpj {
  cnpj: string;
  razaoSocial: string;
  endereco: CompanyAddressData;
}

interface BrasilApiCnpjResponse {
  cnpj?: string;
  razao_social?: string;
  descricao_tipo_de_logradouro?: string;
  logradouro?: string;
  numero?: string;
  complemento?: string;
  bairro?: string;
  cep?: string;
  municipio?: string;
  uf?: string;
  message?: string;
  type?: string;
  name?: string;
}

const CNPJ_DIGITS = 14;

export function onlyCnpjDigits(value: string): string {
  return value.replace(/\D/g, "").slice(0, CNPJ_DIGITS);
}

export function formatCnpj(value: string): string {
  const digits = onlyCnpjDigits(value);
  if (digits.length <= 2) {
    return digits;
  }
  if (digits.length <= 5) {
    return `${digits.slice(0, 2)}.${digits.slice(2)}`;
  }
  if (digits.length <= 8) {
    return `${digits.slice(0, 2)}.${digits.slice(2, 5)}.${digits.slice(5)}`;
  }
  if (digits.length <= 12) {
    return `${digits.slice(0, 2)}.${digits.slice(2, 5)}.${digits.slice(5, 8)}/${digits.slice(8)}`;
  }
  return `${digits.slice(0, 2)}.${digits.slice(2, 5)}.${digits.slice(5, 8)}/${digits.slice(8, 12)}-${digits.slice(12)}`;
}

export function isCompleteCnpj(value: string): boolean {
  return onlyCnpjDigits(value).length === CNPJ_DIGITS;
}

function normalizeText(value: unknown): string {
  return String(value ?? "").trim();
}

// Consolida os campos da BrasilAPI no mesmo formato usado pelo formulario.
export function mapBrasilApiCnpjResponse(payload: unknown): CompanyByCnpj | null {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return null;
  }

  const data = payload as BrasilApiCnpjResponse;
  const tipoLogradouro = normalizeText(data.descricao_tipo_de_logradouro);
  const nomeLogradouro = normalizeText(data.logradouro);
  const logradouro =
    tipoLogradouro !== "" && nomeLogradouro !== ""
      ? `${tipoLogradouro} ${nomeLogradouro}`
      : tipoLogradouro || nomeLogradouro;

  return {
    cnpj: formatCnpj(normalizeText(data.cnpj)),
    razaoSocial: normalizeText(data.razao_social),
    endereco: {
      cep: normalizeText(data.cep),
      logradouro,
      numero: normalizeText(data.numero),
      complemento: normalizeText(data.complemento),
      bairro: normalizeText(data.bairro),
      cidade: normalizeText(data.municipio),
      uf: normalizeText(data.uf).toUpperCase()
    }
  };
}

// Consulta gratuita de CNPJ (BrasilAPI) para reproduzir o comportamento do app Python.
export async function lookupCompanyByCnpj(cnpj: string): Promise<CompanyByCnpj | null> {
  const digits = onlyCnpjDigits(cnpj);
  if (digits.length !== CNPJ_DIGITS) {
    return null;
  }

  const response = await fetch(`https://brasilapi.com.br/api/cnpj/v1/${digits}`, {
    method: "GET",
    headers: {
      Accept: "application/json"
    }
  });

  if (response.status === 404) {
    return null;
  }
  if (!response.ok) {
    throw new Error("Falha ao consultar CNPJ na BrasilAPI.");
  }

  const payload = (await response.json()) as unknown;
  return mapBrasilApiCnpjResponse(payload);
}
