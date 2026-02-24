type APIErrorPayload = {
  error?: {
    code?: string;
    message?: string;
  };
};

type ResolvedAPIError = {
  code: string;
  message: string;
};

function pickCode(payload: APIErrorPayload | undefined, fallback = "request_error"): string {
  const rawCode = payload?.error?.code?.trim();
  return rawCode && rawCode.length > 0 ? rawCode : fallback;
}

function pickMessage(payload: APIErrorPayload | undefined): string {
  const rawMessage = payload?.error?.message?.trim();
  return rawMessage && rawMessage.length > 0 ? rawMessage : "";
}

export function resolveAPIError(
  payload: APIErrorPayload | undefined,
  status: number,
  statusText: string
): ResolvedAPIError {
  const messageFromServer = pickMessage(payload);
  const code = pickCode(payload);
  if (messageFromServer) {
    return { code, message: messageFromServer };
  }

  if (status === 0) {
    // Erro sem resposta HTTP: API offline, DNS, bloqueio de rede, etc.
    return {
      code: "network_error",
      message: "Nao foi possivel conectar ao servidor. Verifique se a API esta ativa."
    };
  }

  if (status === 401) {
    return {
      code,
      message: "Sessao expirada ou invalida. Faca login novamente."
    };
  }

  if (status >= 500) {
    return {
      code,
      message: "Servidor indisponivel no momento. Verifique se a API esta ativa."
    };
  }

  if (status === 404) {
    return {
      code,
      message: "Endpoint nao encontrado. Verifique a configuracao da API."
    };
  }

  const normalizedStatusText = statusText.trim();
  if (status > 0 && normalizedStatusText !== "") {
    return { code, message: `Erro na requisicao (HTTP ${status} - ${normalizedStatusText}).` };
  }
  if (status > 0) {
    return { code, message: `Erro na requisicao (HTTP ${status}).` };
  }

  return { code, message: "Erro na requisicao." };
}
