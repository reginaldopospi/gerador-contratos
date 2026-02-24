function eventTarget(event: Event): EventTarget | null {
  // Prioriza target para suportar cenarios com delegacao de eventos no runtime.
  return (event.target ?? event.currentTarget) as EventTarget | null;
}

export function readInputValue(event: Event): string {
  const target = eventTarget(event);
  return target instanceof HTMLInputElement ? target.value : "";
}

export function readSelectValue(event: Event): string {
  const target = eventTarget(event);
  return target instanceof HTMLSelectElement ? target.value : "";
}

export function readTextareaValue(event: Event): string {
  const target = eventTarget(event);
  return target instanceof HTMLTextAreaElement ? target.value : "";
}
