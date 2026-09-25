import { SemanticValidatorError, type ValidationResult, type DirectValidator } from "@semantic-validator/sdk";

export const cases = [
  { id: "compra", message: "Quiero comprar dos taladros. ¿Cómo puedo pagar?", expected: "Compra: sí; reclamo: no." },
  { id: "reclamo", message: "Mi pedido llegó roto. Quiero presentar un reclamo.", expected: "Compra: no; reclamo: sí." },
  { id: "ambos", message: "Quiero comprar otro taladro, pero también reclamar porque el anterior llegó roto.", expected: "Compra: sí; reclamo: sí." },
  { id: "ninguno", message: "Gracias por la información. Que tengan buen día.", expected: "Compra: no; reclamo: no." },
  { id: "ambiguo", message: "No sé si pedir otro. El anterior no era lo que esperaba.", expected: "Revisión humana: la intención depende del contexto; no se exige una etiqueta concreta." },
];

export async function evaluate(client: Pick<DirectValidator, "check">, message: string) {
  const results = await Promise.allSettled([
    client.check(message, "¿El cliente manifiesta intención de comprar?"),
    client.check(message, "¿El cliente está presentando un reclamo?"),
  ]);
  const failed = results.find(result => result.status === "rejected");
  // Wait for both requests before reporting an error. One may have consumed Jev usage.
  if (failed?.status === "rejected") throw failed.reason;
  const [purchase, complaint] = results as [PromiseFulfilledResult<ValidationResult>, PromiseFulfilledResult<ValidationResult>];
  return { purchase: purchase.value, complaint: complaint.value };
}

export function describe(result: ValidationResult): string {
  const decision = result.status === "uncertain" || result.valid === null
    ? "INCIERTO — requiere revisión"
    : result.valid ? "SÍ" : "NO";
  return `${decision} | status=${result.status} | confidence=${result.confidence}`;
}

export function errorMessage(error: unknown): string {
  if (!(error instanceof SemanticValidatorError)) return "No se pudo ejecutar la prueba. Revisa la configuración y el ejemplo.";
  if (error.statusCode === 401) return "401: la clave de Jev es incorrecta o está revocada. Revisa TYPESAFE_API_KEY.";
  if (error.statusCode === 429) return "429: Jev limitó las solicitudes. Revisa tu cuenta y espera antes de repetir.";
  if (error.code === "provider_timeout") return "Se agotó el tiempo de espera de Jev. No se reintenta automáticamente.";
  if (error.code === "provider_error") return "No se pudo consultar Jev. Revisa tu conexión y el estado de tu cuenta.";
  return "Jev devolvió una respuesta no válida. Comprueba la versión del SDK.";
}
