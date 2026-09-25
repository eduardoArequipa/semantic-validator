# Semantic Validator

[Español](README.es.md) · [Field rules](docs/RULES.md) · [Install a release](docs/RELEASE.md) · [Self-hosted setup](docs/GETTING_STARTED.md) · [Open-source status](docs/OPEN_SOURCE.md)

Semantic Validator checks what text **means**, alongside your usual format and
required-field checks. It provides a Go REST API and SDKs for TypeScript,
Python, Go, and Java. Jev performs semantic
inference; the server handles the API contract, reusable rules, API keys,
quotas, and an in-memory cache. Run it with **your own Jev key** (BYOK).
Each SDK now also has a direct client that calls Jev with your key, without running a server.

Use it to ask whether a product description identifies a product, whether a
chatbot answer addresses a question, or whether a support request explains an
observable problem. These are examples of custom questions through
`POST /v1/check`. The beta includes `person_name`, `product_description`, and
`address`; see the [rule definitions and limits](docs/RULES.md) and the
[90-case evaluation](evals/RESULTS-2026-09-25.md). The latter is a small
fictional benchmark, not a general accuracy claim.

The code is available under [Apache License 2.0](LICENSE) in the
[GitHub repository](https://github.com/eduardoArequipa/semantic-validator);
Jev itself is not included. The TypeScript, Python, and Java SDKs are not yet
published to npm, PyPI, or Maven Central. The Go SDK is part of this module
at `github.com/eduardoArequipa/semantic-validator/sdk/go`.

## Bring your own Jev key

Create a Jev key in your own TypeSafe account. For direct mode, set
`TYPESAFE_API_KEY` in a trusted backend and use `DirectValidator` (TypeScript,
Python, Java) or `NewDirectClient` (Go). No Semantic Validator server or key is
required. For self-hosted API mode, run the Go server with your Jev key and
give callers a separate `SEMANTIC_VALIDATOR_API_KEY`. See the
[direct SDK guide](https://validator.trialsur.cloud/docs/byok-en.html) or the
[self-hosted guide](docs/GETTING_STARTED.md).

```ts
import { DirectValidator } from "@semantic-validator/sdk";

const jevApiKey = process.env.TYPESAFE_API_KEY;
if (!jevApiKey) throw new Error("Missing TYPESAFE_API_KEY");
const validator = new DirectValidator({ jevApiKey });
const result = await validator.check("My order arrived damaged", "Is this a complaint?");
console.log(result.valid, result.confidence, result.status);
```

The public TypeSafe [customer agreement](https://typesafe.ai/legal/mca)
distinguishes integration into customer applications from making its services
available as a standalone service. **Do not treat this project's Apache-2.0
license as permission to resell or publicly proxy Jev.** Check your own
agreement and obtain written authorization before offering a hosted service
to third parties. This is a project caution, not legal advice.

A response has `valid`, `confidence`, and `status`. For example,
`{"valid":true,"confidence":0.96,"status":"valid"}` is illustrative;
actual results can differ. A provider score below 0.80 currently produces
`{"valid":null,"status":"uncertain"}`. Confidence is the provider's score,
not a calibrated probability of correctness. Handle API failures separately
from negative decisions.

In self-hosted individual-key mode, each accepted check consumes one daily quota unit,
including cache hits. Direct calls use your Jev account and do not use our quotas.
Keep keys on trusted backends; do not bundle them into browser or mobile code.
In direct mode text goes to Jev; in self-hosted mode it first goes to your server.
Use fictional data in public examples.

## Run the server locally

Requires Go 1.22 or newer, a Jev token, and an available port 8080.

```bash
cp .env.example .env
# Edit .env: set TYPESAFE_API_KEY to your private Jev token.
set -a
source .env
set +a
go run ./cmd/server
```

The example configuration uses `SEMANTIC_VALIDATOR_API_KEY=local-dev` for
local development. Change it before exposing the service. `.env` is excluded
from Git and Docker build context. Docker Compose binds only to `127.0.0.1`
by default.
From another terminal:

```bash
curl --fail-with-body http://localhost:8080/v1/validate \
  -H 'Authorization: Bearer local-dev' \
  -H 'Content-Type: application/json' \
  -d '{"rule":"person_name","value":"Jorge Eduardo"}'
```

The three built-in field rules are documented [here](docs/RULES.md).
`POST /v1/validate/batch` accepts
1–100 rule checks; `POST /v1/check` accepts a custom yes/no question. See the
[OpenAPI contract](internal/api/openapi.yaml) for request and error schemas.

## SDKs and examples

The [online SDK guide](https://validator.trialsur.cloud/docs/byok-en.html)
covers direct and self-hosted setup plus downloadable TypeScript, Python, Go, and Java SDKs. The
previous product landing page and hosted demo are no longer served. The
production Compose file is configured with `DOCS_ONLY=true`: only SDK/API
documentation stays online, while `/v1/*` returns 503
without requiring a Jev key. The development Compose and direct Go server
remain in validation mode unless you explicitly set `DOCS_ONLY=true`.

## Develop and contribute

```bash
go test ./...
go vet ./...
python3 -m unittest discover -s tools -p 'test_*.py'
```

These tests do not need live API keys. For repeatable semantic quality checks,
use [labeled evaluation cases](evals/README.md); the sample cases are fictional
and a dry run sends no requests.

See [contribution guidelines](CONTRIBUTING.md), the
[open-source notes](docs/OPEN_SOURCE.md), and the
[Spanish README](README.es.md). This project is free and open source;
there is no paid hosted plan. Jev usage is governed by each user's own
TypeSafe account and terms.
