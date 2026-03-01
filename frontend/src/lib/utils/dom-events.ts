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
  if (target instanceof HTMLSelectElement) {
    return target.value;
  }

  // Em eventos delegados, target pode ser OPTION; subimos para o SELECT pai.
  if (target instanceof HTMLOptionElement) {
    const selectFromOption = target.closest("select");
    if (selectFromOption instanceof HTMLSelectElement) {
      return selectFromOption.value;
    }
  }

  // Fallback defensivo para alvos internos que estejam dentro de um SELECT.
  if (target instanceof Element) {
    const closestSelect = target.closest("select");
    if (closestSelect instanceof HTMLSelectElement) {
      return closestSelect.value;
    }
  }

  return "";
}

export function readTextareaValue(event: Event): string {
  // Prioriza currentTarget para manter consistencia quando target vier de um filho interno.
  const target = (event.currentTarget ?? event.target) as EventTarget | null;
  return target instanceof HTMLTextAreaElement ? target.value : "";
}
