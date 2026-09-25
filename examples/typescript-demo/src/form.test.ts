import { test } from "node:test";
import assert from "node:assert/strict";
import { SemanticValidatorError, type ValidationResult } from "@semantic-validator/sdk";
import { validateForm } from "./form.js";

test("form accepts, rejects and requests review without collapsing the states", async () => {
  const asked: string[] = [];
  const answers: Record<string, ValidationResult> = {
    person_name: { valid: true, confidence: 0.96, status: "valid" },
    product_description: { valid: false, confidence: 0.93, status: "invalid" },
    address: { valid: null, confidence: 0.61, status: "uncertain" },
  };
  const result = await validateForm({ validate: async (rule) => {
    asked.push(rule);
    return answers[rule]!;
  } }, { name: "Ana Pérez", productDescription: "Bonito", address: "Cerca del parque" });
  assert.deepEqual(asked, ["person_name", "product_description", "address"]);
  assert.equal(result.name.state, "accepted");
  assert.equal(result.productDescription.state, "rejected");
  assert.equal(result.address.state, "review");
});

test("provider failure is an error, not an invalid field", async () => {
  const result = await validateForm({ validate: async (rule) => {
    if (rule === "address") throw new SemanticValidatorError("private details", { statusCode: 401 });
    return { valid: true, confidence: 0.95, status: "valid" };
  } }, { name: "Ana Pérez", productDescription: "Taladro 18 V", address: "Av. Libertad 123" });
  assert.equal(result.name.state, "accepted");
  assert.equal(result.address.state, "error");
  assert.doesNotMatch(result.address.message, /private details/);
});
