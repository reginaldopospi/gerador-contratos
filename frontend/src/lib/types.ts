export interface AuthUser {
  id: string;
  tenant_id: string;
  email: string;
  role: "admin" | "gestor" | "operador";
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  access_expires_at: string;
  refresh_expires_at: string;
}

export interface AuthResponse {
  user: AuthUser;
  tokens: AuthTokens;
}

export interface Contract {
  id: string;
  tenant_id: string;
  numero: string;
  tipo: string;
  status: string;
  updated_at: string;
  created_at: string;
}

export interface ContractVersion {
  id: string;
  contract_id: string;
  version_number: number;
  data: Record<string, unknown>;
  created_at: string;
}

export interface ContractDetails {
  contract: Contract;
  latest_version?: ContractVersion;
  versions: ContractVersion[];
}

export interface Broker {
  id: string;
  nome: string;
  cpf?: string;
  creci?: string;
  banco?: string;
  agencia?: string;
  conta?: string;
  pix?: string;
}

export interface ClauseTemplate {
  id: string;
  clause_key: string;
  title: string;
  content: string;
  is_active: boolean;
}
