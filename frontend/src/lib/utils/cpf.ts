const CPF_DIGITS = 11;

// Mantem apenas digitos e limita ao tamanho oficial de CPF.
export function onlyCpfDigits(value: string): string {
  return value.replace(/\D/g, "").slice(0, CPF_DIGITS);
}

// Aplica mascara progressiva para nao travar a digitacao no formulario.
export function formatCpf(value: string): string {
  const digits = onlyCpfDigits(value);
  if (digits.length <= 3) {
    return digits;
  }
  if (digits.length <= 6) {
    return `${digits.slice(0, 3)}.${digits.slice(3)}`;
  }
  if (digits.length <= 9) {
    return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6)}`;
  }
  return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6, 9)}-${digits.slice(9)}`;
}

export function isCompleteCpf(value: string): boolean {
  return onlyCpfDigits(value).length === CPF_DIGITS;
}
