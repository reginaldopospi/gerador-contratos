export function readInputValue(event: Event): string {
  // Prioriza currentTarget para manter consistencia quando target vier de um filho interno.
  const target = (event.currentTarget ?? event.target) as EventTarget | null;
  return target instanceof HTMLInputElement ? target.value : "";
}

export function readSelectValue(event: Event): string {
  // Prioriza target para capturar o valor mais recente selecionado.
  const target = event.target as EventTarget | null;
  if (target instanceof HTMLSelectElement) {
    return target.value;
  }

  // Em alguns navegadores o target do change pode ser OPTION.
  if (target instanceof HTMLOptionElement) {
    if (target.value !== "") {
      return target.value;
    }
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

  // Fallback para casos em que apenas currentTarget venha tipado como SELECT.
  const current = event.currentTarget as EventTarget | null;
  if (current instanceof HTMLSelectElement) {
    return current.value;
  }

  return "";
}

export function readTextareaValue(event: Event): string {
  // Prioriza currentTarget para manter consistencia quando target vier de um filho interno.
  const target = (event.currentTarget ?? event.target) as EventTarget | null;
  return target instanceof HTMLTextAreaElement ? target.value : "";
}
