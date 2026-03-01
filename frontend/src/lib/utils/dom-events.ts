export function readInputValue(event: Event): string {
  // Prioriza currentTarget para manter consistencia quando target vier de um filho interno.
  const target = (event.currentTarget ?? event.target) as EventTarget | null;
  return target instanceof HTMLInputElement ? target.value : "";
}

export function readSelectValue(event: Event): string {
  // Em alguns navegadores o target do change pode ser OPTION; usamos currentTarget primeiro.
  const current = event.currentTarget as EventTarget | null;
  if (current instanceof HTMLSelectElement) {
    return current.value;
  }
  const target = event.target as EventTarget | null;
  return target instanceof HTMLSelectElement ? target.value : "";
}

export function readTextareaValue(event: Event): string {
  // Prioriza currentTarget para manter consistencia quando target vier de um filho interno.
  const target = (event.currentTarget ?? event.target) as EventTarget | null;
  return target instanceof HTMLTextAreaElement ? target.value : "";
}
