import { test } from "node:test";
import assert from "node:assert/strict";
import { DirectValidator, SemanticValidatorError } from "@semantic-validator/sdk";
import { describe, errorMessage, evaluate } from "./demo.js";

test("el SDK consulta Jev directamente y preserva uncertain y decisiones independientes", async () => {
  const received: string[] = [];
  const client = new DirectValidator({ jevApiKey: "test-only", fetcher: async (url, options) => {
    assert.equal(String(url), "https://api.typesafe.ai/v1/systemone");
    assert.equal(new Headers(options?.headers).get("Authorization"), "Bearer test-only");
    const body = JSON.parse(String(options?.body));
    assert.equal(body.state, "mensaje de prueba");
    received.push(body.questions.result.instructions);
    const confidence = body.questions.result.instructions.includes("comprar") ? 0.6 : 0.95;
    return Response.json({ answers: { result: { choice: "true", confidence } } });
  } });
  const result = await evaluate(client, "mensaje de prueba");
  assert.equal(received.length, 2);
  assert.equal(result.purchase.valid, null);
  assert.equal(result.complaint.valid, true);
  assert.match(describe(result.purchase), /INCIERTO/);
  assert.match(describe({ valid: false, confidence: 0.98, status: "invalid" }), /^NO/);
});

test("explica 401 y 429 de Jev", () => {
  assert.match(errorMessage(new SemanticValidatorError("", { statusCode: 401 })), /revocada/);
  assert.match(errorMessage(new SemanticValidatorError("", { statusCode: 429 })), /Jev/);
});

test("espera ambas consultas aunque una falle y no reintenta", async () => {
  let calls = 0;
  let completed = false;
  await assert.rejects(evaluate({ check: async () => {
    calls++;
    if (calls === 1) throw new SemanticValidatorError("", { statusCode: 401 });
    await new Promise(resolve => setTimeout(resolve, 10));
    completed = true;
    return { valid: true, confidence: 1, status: "valid" };
  } }, "test"), SemanticValidatorError);
  assert.equal(completed, true);
  assert.equal(calls, 2);
});
