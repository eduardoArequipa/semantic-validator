import type { DirectValidator, ValidationResult } from "@semantic-validator/sdk";
import { errorMessage } from "./demo.js";

export interface FormFields {
  name: string;
  productDescription: string;
  address: string;
}

export interface FieldDecision {
  state: "accepted" | "rejected" | "review" | "error";
  confidence?: number;
  message: string;
}

function decide(result: PromiseSettledResult<ValidationResult>): FieldDecision {
  if (result.status === "rejected") {
    return { state: "error", message: errorMessage(result.reason) };
  }
  const { valid, confidence, status } = result.value;
  if (status === "uncertain" || valid === null) {
    return { state: "review", confidence, message: "Revisión manual: la respuesta no es concluyente." };
  }
  if (valid) {
    return { state: "accepted", confidence, message: "El campo parece válido." };
  }
  return { state: "rejected", confidence, message: "El campo no cumple la regla semántica." };
}

/** Run only on a trusted backend: each field can consume one Jev request. */
export async function validateForm(client: Pick<DirectValidator, "validate">, fields: FormFields) {
  const results = await Promise.allSettled([
    client.validate("person_name", fields.name),
    client.validate("product_description", fields.productDescription),
    client.validate("address", fields.address),
  ]);
  return {
    name: decide(results[0]!),
    productDescription: decide(results[1]!),
    address: decide(results[2]!),
  };
}
