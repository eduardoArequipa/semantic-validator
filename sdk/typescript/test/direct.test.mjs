import assert from "node:assert/strict";
import test from "node:test";
import { DirectValidator, Validator } from "../dist/index.js";

function mock(choice = "true", confidence = 0.95) {
  const calls = [];
  const fetcher = async (url, options) => {
    calls.push({ url, options });
    return new Response(JSON.stringify({ answers: { result: { type: "choice", choice, confidence } } }), {
      status: 200, headers: { "Content-Type": "application/json" },
    });
  };
  return { client: new DirectValidator({ jevApiKey: "test-only", fetcher }), calls };
}

test("direct client calls Jev with own key and returns statuses", async () => {
  const { client, calls } = mock();
  assert.deepEqual(await client.name(" Jorge Eduardo "), { valid: true, confidence: 0.95, status: "valid" });
  assert.equal(calls[0].url, "https://api.typesafe.ai/v1/systemone");
  assert.equal(calls[0].options.headers.Authorization, "Bearer test-only");
  const body = JSON.parse(calls[0].options.body);
  assert.equal(body.model, "jev-latest");
  assert.equal(body.state, "Jorge Eduardo");
  assert.equal(body.questions.result.type, "choice");
  assert.match(body.questions.result.instructions, /nombre de una persona/);
  assert.equal((await mock("false", 0.98).client.check("No", "¿Compra?")).status, "invalid");
  assert.deepEqual(await mock("true", 0.63).client.check("texto", "pregunta"), { valid: null, confidence: 0.63, status: "uncertain" });
});

test("direct batch keeps order and reports item errors", async () => {
  const { client, calls } = mock();
  const result = await client.validateBatch([
    { id: "a", rule: "person_name", value: "Jorge" },
    { id: "b", rule: "missing", value: "No" },
    { id: "c", rule: "person_name", value: " " },
  ]);
  assert.equal(result.length, 3);
  assert.deepEqual(result.map((item) => item.id), ["a", "b", "c"]);
  assert.equal(result[1].error.code, "unknown_rule");
  assert.equal(result[2].error.code, "invalid_request");
  assert.equal(calls.length, 1);
});

test("direct client sends versioned field-rule questions", async () => {
  const { client, calls } = mock();
  await client.validate("product_description", "Taladro de 18 V");
  await client.validate("address", "Av. Bolívar 123");
  assert.match(JSON.parse(calls[0].options.body).questions.result.instructions, /característica concreta/);
  assert.match(JSON.parse(calls[1].options.body).questions.result.instructions, /dirección física/);
});

test("direct client rejects missing keys and invalid provider responses", async () => {
  assert.throws(() => new DirectValidator({ jevApiKey: "" }), /jevApiKey/);
  assert.throws(() => new DirectValidator({ jevApiKey: "key", baseURL: "http://example.com" }), /api.typesafe.ai/);
  assert.throws(() => new DirectValidator({ jevApiKey: "key", baseURL: "https://validator.trialsur.cloud" }), /api.typesafe.ai/);
  const client = new DirectValidator({ jevApiKey: "key", fetcher: async () => new Response("{}") });
  await assert.rejects(() => client.check("value", "question"), { code: "invalid_response" });
  const fail = new DirectValidator({ jevApiKey: "key", fetcher: async () => new Response("error", { status: 401 }) });
  await assert.rejects(() => fail.check("value", "question"), { code: "provider_error", statusCode: 401 });
  const legacy = new Validator({ apiKey: "local-dev", fetcher: async () => new Response(JSON.stringify({ valid: true, confidence: 1, status: "valid" })) });
  assert.equal((await legacy.name("Jorge")).valid, true);
});
