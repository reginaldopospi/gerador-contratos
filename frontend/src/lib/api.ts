import { clearAuth, getAuthState, setAuth } from "./stores/auth";
import type {
  AuthResponse,
  Broker,
  ClauseTemplate,
  Contract,
  ContractDetails
} from "./types";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api/v1";

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

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined
  });

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
    const code = payload?.error?.code ?? "request_error";
    const message = payload?.error?.message ?? "Erro na requisicao";
    throw new APIError(message, response.status, code);
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
  }): Promise<AuthResponse> {
    return request<AuthResponse>("/auth/register", {
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

  async previewContract(contractID: string): Promise<{ title: string; sections: string[] }> {
    const response = await request<{ preview: { title: string; sections: string[] } }>(
      `/contracts/${contractID}/preview`
    );
    return response.preview;
  },

  async previewContractFromData(payload: {
    numero: string;
    tipo: string;
    data: Record<string, unknown>;
  }): Promise<{ title: string; sections: string[] }> {
    const response = await request<{ preview: { title: string; sections: string[] } }>(
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
  }
};

export { APIError };
