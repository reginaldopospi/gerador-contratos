export interface CepAddress {
  cep: string;
  logradouro: string;
  complemento: string;
  bairro: string;
  cidade: string;
  uf: string;
}

interface ViaCepResponse {
  cep?: string;
  logradouro?: string;
  complemento?: string;
  bairro?: string;
  localidade?: string;
  uf?: string;
  erro?: boolean;
}

const CEP_DIGITS = 8;

export function onlyCepDigits(value: string): string {
  return value.replace(/\D/g, "").slice(0, CEP_DIGITS);
}

export function formatCep(value: string): string {
  const digits = onlyCepDigits(value);
  if (digits.length <= 5) {
    return digits;
  }
  return `${digits.slice(0, 5)}-${digits.slice(5)}`;
}

export function isCompleteCep(value: string): boolean {
  return onlyCepDigits(value).length === CEP_DIGITS;
}

export function mapViaCepResponse(payload: unknown): CepAddress | null {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return null;
  }

  const data = payload as ViaCepResponse;
  if (data.erro) {
    return null;
  }

  return {
    cep: formatCep(String(data.cep ?? "")),
    logradouro: String(data.logradouro ?? "").trim(),
    complemento: String(data.complemento ?? "").trim(),
    bairro: String(data.bairro ?? "").trim(),
    cidade: String(data.localidade ?? "").trim(),
    uf: String(data.uf ?? "").trim()
  };
}

// ViaCEP e um servico gratuito e nao exige autenticacao.
export async function lookupAddressByCep(cep: string): Promise<CepAddress | null> {
  const digits = onlyCepDigits(cep);
  if (digits.length !== CEP_DIGITS) {
    return null;
  }

  const response = await fetch(`https://viacep.com.br/ws/${digits}/json/`, {
    method: "GET",
    headers: {
      Accept: "application/json"
    }
  });

  if (!response.ok) {
    throw new Error("Falha ao consultar CEP no ViaCEP.");
  }

  const payload = (await response.json()) as unknown;
  return mapViaCepResponse(payload);
}
