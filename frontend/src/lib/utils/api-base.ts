const DEFAULT_API_BASE = "/api/v1";

function removeTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

export function resolveApiBase(rawBase?: string): string {
  const normalized = (rawBase ?? "").trim();
  if (normalized === "") {
    // Default relativo para funcionar com proxy local e deploy no mesmo dominio.
    return DEFAULT_API_BASE;
  }

  if (
    normalized.startsWith("/") ||
    normalized.startsWith("http://") ||
    normalized.startsWith("https://")
  ) {
    return removeTrailingSlash(normalized);
  }

  // Suporta configuracoes sem protocolo, ex.: localhost:8080/api/v1.
  return removeTrailingSlash(`http://${normalized}`);
}
