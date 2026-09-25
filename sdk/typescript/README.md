# Semantic Validator SDK for TypeScript

TypeScript clients for direct Jev BYOK and for a self-hosted Semantic Validator REST API. Requires Node.js 18+.

The built-in field rules are `person_name`, `product_description`, and `address` (v1).
Use `validate(rule, value)` for them; see the [definitions](../../docs/RULES.md).

## Direct Jev mode (recommended)

```ts
import { DirectValidator } from "@semantic-validator/sdk";

const jevApiKey = process.env.TYPESAFE_API_KEY;
if (!jevApiKey) throw new Error("Missing TYPESAFE_API_KEY");
const validator = new DirectValidator({ jevApiKey });
const result = await validator.check("My order arrived damaged", "Is this a complaint?");
console.log(result.valid, result.confidence, result.status);
```

Use only in trusted Node.js backend code. `name`, `validate`, and `validateBatch` are also available. Each batch item calls Jev separately; confidence below 0.80 returns `valid: null`.

## Install locally

From this directory:

```bash
npm install
npm run build
```

In an application in the same repository, install the generated package with:

```bash
npm install /path/to/semantic-validator/sdk/typescript
```

## Self-hosted server mode (optional)

```ts
import { Validator } from "@semantic-validator/sdk";

const validator = new Validator({
  apiKey: process.env.SEMANTIC_VALIDATOR_API_KEY!,
  baseURL: "http://localhost:8080",
});

const result = await validator.name("Jorge Eduardo");
console.log(result.valid, result.confidence, result.status);

const answer = await validator.check(
  "Necesito devolver el producto porque llegó roto",
  "¿El cliente está solicitando una devolución?",
);
console.log(answer.valid);

const batch = await validator.validateBatch([
  { id: "name", rule: "person_name", value: "Jorge Eduardo" },
  { id: "invalid-name", rule: "person_name", value: "sdgfxcdg" },
]);
console.log(batch);
```

The SDK sends `Authorization: Bearer <apiKey>`. Keep the Semantic Validator key on your server; never expose it in browser code. `valid` can be `null` when the semantic result is uncertain. HTTP and transport failures throw `SemanticValidatorError`, which exposes `code` and, for HTTP errors, `statusCode`.
