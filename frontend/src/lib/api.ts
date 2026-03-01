import { clearAuth, getAuthState, setAuth } from "./stores/auth";
import { resolveApiBase } from "./utils/api-base";
import { resolveAPIError } from "./utils/api-error";
import { isCompleteCnpj, onlyCnpjDigits, type CompanyByCnpj } from "./utils/cnpj";
import type {
  AuthResponse,
  Broker,
  ClauseTemplate,
  Contract,
  ContractDetails,
  ContractPreview,
  PendingRegistration,
  RegisterTenantResponse,
  TenantSummary
} from "./types";

const API_BASE = resolveApiBase(import.meta.env.VITE_API_BASE_URL);

class APIError extends Error {
  status: number;
  code: string;

  constructor(message: string, status: number, code: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

type RequestOptions = {
  method?: string;
  body?: unknown;
  auth?: boolean;
  retryOnAuth?: boolean;
};

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const {
    method = "GET",
    body,
    auth = true,
    retryOnAuth = true
  } = options;

  const headers: Record<string, string> = {
    "Content-Type": "application/json"
  };

  const authState = getAuthState();
  if (auth && authState.accessToken) {
    headers.Authorization = `Bearer ${authState.accessToken}`;
  }

  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined
    });
  } catch {
    const resolved = resolveAPIError(undefined, 0, "");
    throw new APIError(resolved.message, 0, resolved.code);
  }

  if (response.status === 401 && auth && retryOnAuth && authState.refreshToken) {
    const refreshed = await refreshSession(authState.refreshToken);
    if (refreshed) {
      return request<T>(path, {
        method,
        body,
        auth,
        retryOnAuth: false
      });
    }
  }

  const payload = await safeJson(response);
  if (!response.ok) {
    const resolved = resolveAPIError(payload, response.status, response.statusText);
    throw new APIError(resolved.message, response.status, resolved.code);
  }

  return payload as T;
}

async function safeJson(response: Response): Promise<any> {
  try {
    return await response.json();
  } catch {
    return {};
  }
}

async function refreshSession(refreshToken: string): Promise<boolean> {
  try {
    const payload = await request<AuthResponse>("/auth/refresh", {
      method: "POST",
      auth: false,
      retryOnAuth: false,
      body: { refresh_token: refreshToken }
    });

    setAuth({
      accessToken: payload.tokens.access_token,
      refreshToken: payload.tokens.refresh_token,
      user: payload.user
    });
    return true;
  } catch {
    clearAuth();
    return false;
  }
}

export const api = {
  async registerTenant(input: {
    tenant_name: string;
    tenant_cnpj: string;
    name: string;
    email: string;
    password: string;
  }): Promise<RegisterTenantResponse> {
    return request<RegisterTenantResponse>("/auth/register", {
      method: "POST",
      auth: false,
      body: input
    });
  },

  async login(input: { email: string; password: string }): Promise<AuthResponse> {
    return request<AuthResponse>("/auth/login", {
      method: "POST",
      auth: false,
      body: input
    });
  },

  async me(): Promise<{ user: AuthResponse["user"] }> {
    return request<{ user: AuthResponse["user"] }>("/auth/me");
  },

  async registerUser(input: {
    name: string;
    email: string;
    password: string;
    role: "admin" | "gestor" | "operador";
  }): Promise<{ user: AuthResponse["user"] }> {
    return request("/auth/users", { method: "POST", body: input });
  },

  async listPendingRegistrations(): Promise<PendingRegistration[]> {
    const response = await request<{ items: PendingRegistration[] }>("/auth/pending-registrations");
    return response.items;
  },

  async listTenants(): Promise<TenantSummary[]> {
    // Lista as imobiliarias cadastradas para o painel do admin da plataforma.
    const response = await request<{ items: TenantSummary[] }>("/auth/tenants");
    return response.items;
  },

  async updateTenant(
    tenantID: string,
    input: { tenant_name: string; tenant_cnpj?: string }
  ): Promise<{ tenant: { id: string; nome_fantasia: string; cnpj?: string }; message: string }> {
    // Permite ajustar os dados cadastrais da imobiliaria pelo painel de plataforma.
    return request<{ tenant: { id: string; nome_fantasia: string; cnpj?: string }; message: string }>(
      `/auth/tenants/${encodeURIComponent(tenantID)}`,
      {
        method: "PUT",
        body: input
      }
    );
  },

  async deleteTenant(tenantID: string): Promise<{ message: string }> {
    // Remove a imobiliaria e seus registros associados via backend.
    return request<{ message: string }>(`/auth/tenants/${encodeURIComponent(tenantID)}`, {
      method: "DELETE"
    });
  },

  async resetTenantAdminPassword(
    tenantID: string,
    newPassword: string
  ): Promise<{ message: string }> {
    // Redefine a senha do admin principal da imobiliaria selecionada.
    return request<{ message: string }>(
      `/auth/tenants/${encodeURIComponent(tenantID)}/admin-password`,
      {
        method: "POST",
        body: { new_password: newPassword }
      }
    );
  },

  async approvePendingRegistration(userID: string): Promise<{ user: AuthResponse["user"]; message: string }> {
    // Aprovacao apenas ativa o cadastro; troca de senha ocorre depois pelo fluxo apropriado.
    return request<{ user: AuthResponse["user"]; message: string }>(
      `/auth/pending-registrations/${encodeURIComponent(userID)}/approve`,
      {
        method: "POST"
      }
    );
  },

  async forgotPassword(email: string): Promise<{ message: string; dev_reset_token?: string }> {
    return request("/auth/forgot-password", {
      method: "POST",
      auth: false,
      body: { email }
    });
  },

  async resetPassword(input: { token: string; new_password: string }): Promise<{ message: string }> {
    return request("/auth/reset-password", {
      method: "POST",
      auth: false,
      body: input
    });
  },

  async listContracts(query = ""): Promise<Contract[]> {
    const suffix = query ? `?q=${encodeURIComponent(query)}` : "";
    const response = await request<{ items: Contract[] }>(`/contracts${suffix}`);
    return response.items;
  },

  async createContract(payload: {
    numero: string;
    tipo: string;
    status: string;
    data: Record<string, unknown>;
  }): Promise<ContractDetails> {
    return request<ContractDetails>("/contracts", {
      method: "POST",
      body: payload
    });
  },

  async getContract(contractID: string): Promise<ContractDetails> {
    return request<ContractDetails>(`/contracts/${contractID}`);
  },

  async addContractVersion(contractID: string, data: Record<string, unknown>): Promise<void> {
    await request(`/contracts/${contractID}/versions`, {
      method: "POST",
      body: { data }
    });
  },

  async previewContract(contractID: string): Promise<ContractPreview> {
    const response = await request<{ preview: ContractPreview }>(
      `/contracts/${contractID}/preview`
    );
    return response.preview;
  },

  async previewContractFromData(payload: {
    numero: string;
    tipo: string;
    data: Record<string, unknown>;
  }): Promise<ContractPreview> {
    const response = await request<{ preview: ContractPreview }>(
      "/contracts/preview",
      {
        method: "POST",
        body: payload
      }
    );
    return response.preview;
  },

  async listBrokers(): Promise<Broker[]> {
    const response = await request<{ items: Broker[] }>("/brokers");
    return response.items;
  },

  async createBroker(payload: {
    nome: string;
    cpf: string;
    creci: string;
    banco: string;
    agencia: string;
    conta: string;
    pix: string;
  }): Promise<Broker> {
    const response = await request<{ broker: Broker }>("/brokers", {
      method: "POST",
      body: payload
    });
    return response.broker;
  },

  async updateBroker(
    brokerID: string,
    payload: {
      nome: string;
      cpf: string;
      creci: string;
      banco: string;
      agencia: string;
      conta: string;
      pix: string;
    }
  ): Promise<Broker> {
    const response = await request<{ broker: Broker }>(`/brokers/${brokerID}`, {
      method: "PUT",
      body: payload
    });
    return response.broker;
  },

  async deleteBroker(brokerID: string): Promise<void> {
    await request(`/brokers/${brokerID}`, { method: "DELETE" });
  },

  async listClauses(): Promise<ClauseTemplate[]> {
    const response = await request<{ items: ClauseTemplate[] }>("/clauses");
    return response.items;
  },

  async upsertClause(payload: {
    clause_key: string;
    title: string;
    content: string;
    is_active: boolean;
  }): Promise<ClauseTemplate> {
    const response = await request<{ clause: ClauseTemplate }>("/clauses", {
      method: "POST",
      body: payload
    });
    return response.clause;
  },

  async deleteClause(clauseID: string): Promise<void> {
    await request(`/clauses/${clauseID}`, { method: "DELETE" });
  },

  async lookupCompanyByCnpj(cnpj: string): Promise<CompanyByCnpj | null> {
    const digits = onlyCnpjDigits(cnpj);
    if (!isCompleteCnpj(digits)) {
      return null;
    }

    try {
      const response = await request<{
        company: {
          cnpj: string;
          razao_social: string;
          endereco: {
            cep: string;
            logradouro: string;
            numero: string;
            complemento: string;
            bairro: string;
            cidade: string;
            uf: string;
          };
        };
      }>(`/cnpj/${digits}`);

      if (!response.company) {
        return null;
      }

      // Mantem o formato usado no editor para evitar alterações no restante do fluxo.
      return {
        cnpj: response.company.cnpj ?? "",
        razaoSocial: response.company.razao_social ?? "",
        endereco: {
          cep: response.company.endereco?.cep ?? "",
          logradouro: response.company.endereco?.logradouro ?? "",
          numero: response.company.endereco?.numero ?? "",
          complemento: response.company.endereco?.complemento ?? "",
          bairro: response.company.endereco?.bairro ?? "",
          cidade: response.company.endereco?.cidade ?? "",
          uf: response.company.endereco?.uf ?? ""
        }
      };
    } catch (err) {
      if (
        err instanceof APIError &&
        err.status === 404 &&
        err.code === "cnpj_not_found"
      ) {
        return null;
      }
      throw err;
    }
  }
};

export { APIError };
